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

package v1

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var podlog = logf.Log.WithName("pod-resource")

// +kubebuilder:webhook:path=/mutate-core-v1-pod,mutating=true,failurePolicy=fail,groups=core,resources=pods,verbs=create;update,versions=v1,admissionReviewVersions=v1,sideEffects=none,name=mpod.kb.io

type PodLabelMutate struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewPodSideCarMutate(c client.Client) admission.Handler {
	return &PodLabelMutate{Client: c}
}

const (
	istioProxyName = "istio-proxy"
	channelEnv     = "ORBIT_CHANNEL_TAG"
)

// PodLabelMutate injects a key-value pair to istio-proxy sidecar if a specific label exists.
func (v *PodLabelMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	tag := ""
	if val, ok := pod.Labels[KUBEORBIT_CHANNEL_LABEL]; ok {
		tag = val
	}

	sidecarIndex := -1
	for k, container := range pod.Spec.Containers {
		if container.Name == istioProxyName {
			sidecarIndex = k
		}
	}

	if tag != "" && sidecarIndex > 0 {
		pod.Spec.Containers[sidecarIndex].Env = append(
			pod.Spec.Containers[sidecarIndex].Env,
			corev1.EnvVar{
				Name:  channelEnv,
				Value: tag,
			})
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// PodLabelMutate implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *PodLabelMutate) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
