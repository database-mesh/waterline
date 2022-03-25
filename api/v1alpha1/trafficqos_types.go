// Copyright 2022 Database Mesh Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type QoSClassType string

const (
	QoSClassGuaranteed QoSClassType = "Guaranteed"
	QoSClassBurstable  QoSClassType = "Burstable"
	QoSClassBestEffort QoSClassType = "BestEffort"
)

type TrafficQoSStrategy string

const (
	TrafficQoSStrategyDynamic    TrafficQoSStrategy = "Dynamic"
	TrafficQoSStrategyPreDefined TrafficQoSStrategy = "PreDefined"
)

type TrafficQoSGroup struct {
	Rate string `json:"rate"`
	Ceil string `json:"ceil",omitempty`
}

// TrafficQoSSpec defines the desired state of TrafficQoS
type TrafficQoSSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TrafficQoS. Edit trafficqos_types.go to remove/update
	NetworkDevice string             `json:"networkDevice"`
	QoSClass      QoSClassType       `json:"qosClass,omitempty"`
	Strategy      TrafficQoSStrategy `json:"strategy",omitempty`
	Groups        []TrafficQoSGroup  `json:"groups"`
}

// TrafficQoSStatus defines the observed state of TrafficQoS
type TrafficQoSStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//TODO: add ObservedGeneration
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TrafficQoS is the Schema for the trafficqos API
type TrafficQoS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrafficQoSSpec   `json:"spec,omitempty"`
	Status TrafficQoSStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TrafficQoSList contains a list of TrafficQoS
type TrafficQoSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrafficQoS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrafficQoS{}, &TrafficQoSList{})
}