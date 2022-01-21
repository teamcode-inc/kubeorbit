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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TrafficRulesSpec struct {
	Headers map[string]string `json:"headers,omitempty"`
}

// OrbitSpec defines the desired state of Orbit
type OrbitSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	MeshProvider string           `json:"provider"`
	TrafficRules TrafficRulesSpec `json:"trafficRules"`
}

// OrbitStatus defines the observed state of Orbit
type OrbitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"Status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Orbit is the Schema for the orbits API
type Orbit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrbitSpec   `json:"spec,omitempty"`
	Status OrbitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// OrbitList contains a list of Orbit
type OrbitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Orbit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Orbit{}, &OrbitList{})
}
