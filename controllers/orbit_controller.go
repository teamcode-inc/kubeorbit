/*
Copyright 2022 The TeamCode authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/types"
	"istio.io/api/networking/v1alpha3"
	istiov1 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	orbitv1alpha1 "kubeorbit.io/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OrbitReconciler reconciles a Orbit object
type OrbitReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=orbits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=orbits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=orbits/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Orbit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *OrbitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("orbit", req.NamespacedName)
	obj := &orbitv1alpha1.Orbit{}

	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Println(err, "unable to fetch object")
	} else {
		if err := r.reconcileEnvoyFilter(obj, req); err != nil {
			obj.Status.Status = "failed"
			if err := r.Status().Update(context.Background(), obj); err != nil {
				log.Println(err, "unable to update status")
			}
			r.Recorder.Event(obj, corev1.EventTypeWarning, "FailedCreate", "Unable to create EnvoyFilter on kubeorbit")
			return ctrl.Result{}, fmt.Errorf("reconcileEnvoyFilter failed: %w", err)
		}
	}
	obj.Status.Status = "success"
	if err := r.Status().Update(context.Background(), obj); err != nil {
		log.Println(err, "unable to update status")

	}
	r.Recorder.Event(obj, corev1.EventTypeNormal, "Created", "EnvoyFilter created successfully on kubeorbit")
	return ctrl.Result{}, nil
}

func (r *OrbitReconciler) reconcileEnvoyFilter(orbit *orbitv1alpha1.Orbit, req ctrl.Request) error {
	envoyName := orbit.Name
	outboundSpec, err := generateOutboudValue(orbit)
	if err != nil {
		return fmt.Errorf("failed to generate outbound proxy: %w", err)
	}

	newSpec := buildHttpFilter(outboundSpec)
	envoyFilter := &istiov1.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyName,
			Namespace: orbit.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(orbit, schema.GroupVersionKind{
					Group:   orbit.GroupVersionKind().Group,
					Version: orbit.GroupVersionKind().Version,
					Kind:    orbit.Kind,
				}),
			},
		},
		Spec: newSpec,
	}

	err = r.Get(context.TODO(), req.NamespacedName, envoyFilter)
	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), envoyFilter)
		if err != nil {
			return fmt.Errorf("EnvoyFilter %s.%s create error: %w", envoyName, orbit.Namespace, err)
		}
		r.Log.WithValues("orbit", fmt.Sprintf("%s.%s", orbit.Name, orbit.Namespace)).
			Info("EnvoyFilter created", envoyFilter.GetName(), orbit.Namespace)
		return nil
	} else if err != nil {
		return fmt.Errorf("EnvoyFilter %s.%s get query error: %w", envoyName, orbit.Namespace, err)
	}

	if envoyFilter != nil {
		if diff := cmp.Diff(newSpec, envoyFilter.Spec); diff != "" {
			clone := envoyFilter.DeepCopy()
			clone.Spec = newSpec
			err = r.Update(context.TODO(), clone)
			if err != nil {
				r.Recorder.Event(orbit, corev1.EventTypeWarning, "FailedUpdate", "Unable to update EnvoyFilter on kubeorbit")
				return fmt.Errorf("EnvoyFilter %s.%s update error: %w", envoyName, orbit.Namespace, err)
			}
			r.Log.WithValues("orbit", fmt.Sprintf("%s.%s", orbit.Name, orbit.Namespace)).
				Info("EnvoyFilter updated", envoyFilter.GetName(), orbit.Namespace)
		}
	}

	return nil
}

func generateOutboudValue(orbit *orbitv1alpha1.Orbit) (*types.Struct, error) {
	var out = &types.Struct{}

	out.Fields = map[string]*types.Value{}
	out.Fields["@type"] = &types.Value{Kind: &types.Value_StringValue{
		StringValue: "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua",
	}}

	headerKey := ""
	for k, _ := range orbit.Spec.TrafficRules.Headers {
		headerKey = k
	}
	out.Fields["inlineCode"] = &types.Value{Kind: &types.Value_StringValue{
		StringValue: fmt.Sprintf(`function envoy_on_request(handle)
  local tag = handle:headers():get("` + headerKey + `")
  local channelValue = os.getenv("ORBIT_CHANNEL_TAG")
  if tag == nil and channelValue ~= nil then
     handle:headers():add("` + headerKey + `", channelValue)
  end
end`),
	}}

	return &types.Struct{
		Fields: map[string]*types.Value{
			"name": {
				Kind: &types.Value_StringValue{
					StringValue: "envoy.lua",
				},
			},
			"typed_config": {
				Kind: &types.Value_StructValue{StructValue: out},
			},
		},
	}, nil
}

func buildHttpFilter(outboundSpec *types.Struct) v1alpha3.EnvoyFilter {
	return v1alpha3.EnvoyFilter{
		ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
			{
				ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
				Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
					Context: v1alpha3.EnvoyFilter_SIDECAR_OUTBOUND,
					ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
						Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
							FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
								Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
									Name: "envoy.filters.network.http_connection_manager",
									SubFilter: &v1alpha3.EnvoyFilter_ListenerMatch_SubFilterMatch{
										Name: "envoy.filters.http.router",
									},
								},
							},
						}},
				},
				Patch: &v1alpha3.EnvoyFilter_Patch{
					Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
					Value:     outboundSpec,
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrbitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orbitv1alpha1.Orbit{}).
		Owns(&istiov1.EnvoyFilter{}).
		Complete(r)
}
