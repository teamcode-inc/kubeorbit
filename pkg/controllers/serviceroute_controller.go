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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"istio.io/api/networking/v1alpha3"
	istiov1 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	routev1alpha1 "kubeorbit.io/api/v1alpha1"
)

// ServiceRouteReconciler reconciles a ServiceRoute object
type ServiceRouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=serviceroutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=serviceroutes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.kubeorbit.io,resources=serviceroutes/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceRoute object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ServiceRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("serviceroute", req.NamespacedName)
	obj := &routev1alpha1.ServiceRoute{}

	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Println(err, "unable to fetch object")
	} else {
		if err := r.reconcileDestinationRule(obj, req); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcileDestinatinRule failed: %w", err)
		}

		if err := r.reconcileVirtualService(obj, req); err != nil {
			return ctrl.Result{}, fmt.Errorf("reconcileVirtualService failed: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceRouteReconciler) reconcileDestinationRule(tr *routev1alpha1.ServiceRoute, req ctrl.Request) error {
	svcName := tr.GetServiceName()
	newSpec := v1alpha3.DestinationRule{
		Host:    svcName,
		Subsets: buildRoute(tr),
	}

	destinationRule := &istiov1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tr.Name,
			Namespace: tr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(tr, schema.GroupVersionKind{
					Group:   tr.GroupVersionKind().Group,
					Version: tr.GroupVersionKind().Version,
					Kind:    tr.Kind,
				}),
			},
		},
		Spec: newSpec,
	}

	err := r.Get(context.TODO(), req.NamespacedName, destinationRule)
	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), destinationRule)
		if err != nil {
			return fmt.Errorf("DestinationRule %s.%s create error: %w", svcName, tr.Namespace, err)
		}
		r.Log.WithValues("serviceroute", fmt.Sprintf("%s.%s", tr.Name, tr.Namespace)).
			Info("DestinationRule created", destinationRule.GetName(), tr.Namespace)
		return nil
	} else if err != nil {
		return fmt.Errorf("DestinationRule %s.%s get query error: %w", tr.Name, tr.Namespace, err)
	}

	if destinationRule != nil {
		if diff := cmp.Diff(newSpec, destinationRule.Spec); diff != "" {
			clone := destinationRule.DeepCopy()
			clone.Spec = newSpec
			err = r.Update(context.TODO(), clone)
			if err != nil {
				return fmt.Errorf("DestinationRule %s.%s update error: %w", tr.Name, tr.Namespace, err)
			}
			r.Log.WithValues("serviceroute", fmt.Sprintf("%s.%s", tr.Name, tr.Namespace)).
				Info("DestinationRule updated", destinationRule.GetName(), tr.Namespace)
		}
	}

	return nil
}

func (r *ServiceRouteReconciler) reconcileVirtualService(tr *routev1alpha1.ServiceRoute, req ctrl.Request) error {
	svcName := tr.GetServiceName()

	newSpec := v1alpha3.VirtualService{
		Hosts: []string{
			svcName,
		},
		Http: buildHTTP(tr),
	}

	virtualService := &istiov1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tr.Name,
			Namespace: tr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(tr, schema.GroupVersionKind{
					Group:   tr.GroupVersionKind().Group,
					Version: tr.GroupVersionKind().Version,
					Kind:    tr.Kind,
				}),
			},
		},
		Spec: newSpec,
	}

	err := r.Get(context.TODO(), req.NamespacedName, virtualService)
	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), virtualService)
		if err != nil {
			return fmt.Errorf("VirtualService %s.%s create error: %w", tr.Name, tr.Namespace, err)
		}
		r.Log.WithValues("serviceroute", fmt.Sprintf("%s.%s", tr.Spec.Name, tr.Namespace)).
			Info("VirtualService created", virtualService.GetName(), tr.Namespace)
		return nil
	} else if err != nil {
		return fmt.Errorf("VirtualService %s.%s get query error %v", tr.Name, tr.Namespace, err)
	}

	if virtualService != nil {
		if diff := cmp.Diff(
			newSpec,
			virtualService.Spec,
			cmpopts.IgnoreFields(v1alpha3.HTTPRoute{}, "Mirror", "MirrorPercentage"),
		); diff != "" {
			vtClone := virtualService.DeepCopy()
			vtClone.Spec = newSpec
			err = r.Update(context.TODO(), vtClone)
			if err != nil {
				return fmt.Errorf("VirtualService %s.%s update error: %w", tr.Name, tr.Namespace, err)
			}
			r.Log.WithValues("serviceroute", fmt.Sprintf("%s.%s", tr.Spec.Name, tr.Namespace)).
				Info("VirtualService updated", virtualService.GetName(), tr.Namespace)
		}
	}

	return nil
}

func buildRoute(tr *routev1alpha1.ServiceRoute) []*v1alpha3.Subset {
	defaultRoute := tr.Spec.TrafficRoutes.Default
	subsets := make([]*v1alpha3.Subset, 0)

	for _, c := range tr.Spec.TrafficRoutes.TrafficSubset {
		if c.Labels != nil {
			subsets = append(subsets, &v1alpha3.Subset{
				Name:   c.Name,
				Labels: c.Labels,
			})
		}
	}

	for _, c := range defaultRoute {
		if c != "" {
			subsets = append(subsets, &v1alpha3.Subset{
				Name:   c,
				Labels: defaultRoute,
			})
		}
	}

	return subsets
}

func buildHTTP(tr *routev1alpha1.ServiceRoute) []*v1alpha3.HTTPRoute {
	httpRoutes := make([]*v1alpha3.HTTPRoute, 0)
	defaultRoute := tr.Spec.TrafficRoutes.Default

	for _, c := range tr.Spec.TrafficRoutes.TrafficSubset {
		if c.Labels != nil {
			headers := make(map[string]*v1alpha3.StringMatch)
			for k, match := range c.Headers {
				headers[k] = &v1alpha3.StringMatch{
					MatchType: &v1alpha3.StringMatch_Exact{Exact: match.Exact},
				}
			}
			httpRoutes = append(httpRoutes, &v1alpha3.HTTPRoute{
				Match: []*v1alpha3.HTTPMatchRequest{
					{
						Headers: headers,
					},
				},
				Route: []*v1alpha3.HTTPRouteDestination{
					{
						Destination: &v1alpha3.Destination{
							Host:   tr.Spec.Name,
							Subset: c.Name,
						},
					},
				},
			})
		}
	}

	for _, c := range defaultRoute {
		if c != "" {
			httpRoutes = append(httpRoutes, &v1alpha3.HTTPRoute{
				Route: []*v1alpha3.HTTPRouteDestination{
					{
						Destination: &v1alpha3.Destination{
							Host:   tr.Spec.Name,
							Subset: c,
						},
					},
				},
			})
		}
	}

	return httpRoutes
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routev1alpha1.ServiceRoute{}).
		Owns(&istiov1.DestinationRule{}).
		Owns(&istiov1.VirtualService{}).
		Complete(r)
}
