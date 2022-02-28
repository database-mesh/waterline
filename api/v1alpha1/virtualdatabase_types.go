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

type DatabaseServerProtocol string

const (
	DatabaseServerProtocolMySQL DatabaseServerProtocol = "MySQL"
)

type VirtualDatabaseServer struct {
	Port       int                    `json:"port"`
	Protocol   DatabaseServerProtocol `json:"protocol"`
	Credential string                 `json:"credentialName"`
	Backends   []DatabaseSource       `json:"backends"`
}

type DatabaseSource struct {
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Credential string `json:"credentialName"`
}

type SQLTrafficQoS string

// VirtualDatabaseSpec defines the desired state of VirtualDatabase
type VirtualDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VirtualDatabase. Edit virtualdatabase_types.go to remove/update
	Server VirtualDatabaseServer `json:"server"`
	QoS    SQLTrafficQoS         `json:"qos"`
}

// VirtualDatabaseStatus defines the observed state of VirtualDatabase
type VirtualDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VirtualDatabase is the Schema for the virtualdatabases API
type VirtualDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualDatabaseSpec   `json:"spec,omitempty"`
	Status VirtualDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualDatabaseList contains a list of VirtualDatabase
type VirtualDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualDatabase `json:"items"`
}
