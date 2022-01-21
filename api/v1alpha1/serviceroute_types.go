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

type TrafficRouteSpec struct {
	TrafficSubset []*Subset         `json:"routes"`
	Default       map[string]string `json:"default"`
}

type HTTPMatchRequest struct {
	Headers map[string]*StringMatch `json:"headers,omitempty"`
}

type Subset struct {
	// Name of the subset. The service name and the subset name can
	// be used for traffic splitting in a route rule.
	Name string `json:"name,omitempty"`
	// Labels apply a filter over the endpoints of a service in the
	// service registry. See route rules for examples of usage.
	Labels  map[string]string       `json:"labels,omitempty"`
	Headers map[string]*StringMatch `json:"headers,omitempty"`
}

// ServiceRouteSpec defines the desired state of ServiceRoute
type ServiceRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TrafficRoutes TrafficRouteSpec `json:"trafficRoutes"`
	Name          string           `json:"name"`
}

// ServiceRouteStatus defines the observed state of ServiceRoute
type ServiceRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceRoute is the Schema for the serviceroutes API
type ServiceRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceRouteSpec   `json:"spec,omitempty"`
	Status ServiceRouteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceRouteList contains a list of ServiceRoute
type ServiceRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceRoute `json:"items"`
}

func (c *ServiceRoute) GetServiceName() (serviceName string) {
	if c.Spec.Name != "" {
		serviceName = c.Spec.Name
	}
	return
}

type StringMatch struct {
	// Specified exactly one of the fields below.

	// exact string match
	Exact string `json:"exact,omitempty"`

	// prefix-based match
	Prefix string `json:"prefix,omitempty"`

	// suffix-based match.
	Suffix string `json:"suffix,omitempty"`

	// ECMAscript style regex-based match
	Regex string `json:"regex,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ServiceRoute{}, &ServiceRouteList{})
}
