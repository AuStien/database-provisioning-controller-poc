/*


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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Postgres struct {
	// Host is the hostname of the postgres server
	Host string `json:"host,omitempty"`
	// Username is the username associated with the server
	Username string `json:"username,omitempty"`
	// Port is the port of the server
	Port int32 `json:"port,omitempty"`
	// UseSsl is a bool deciding if ssl will be used
	UseSsl bool `json:"useSsl,omitempty"`
}

type Mysql struct {
	// Host is the hostname of the postgres server
	Host string `json:"host,omitempty"`
	// Username is the username associated with the server
	Username string `json:"username,omitempty"`
	// Port is the port of the server
	Port int32 `json:"port,omitempty"`
	// UseSsl is a bool deciding if ssl will be used
	UseSsl bool `json:"useSsl,omitempty"`
}

type Mongo struct {
	// Host is the hostname of the postgres server
	Host string `json:"host,omitempty"`
	// Username is the username associated with the server
	Username string `json:"username,omitempty"`
	// Port is the port of the server
	Port int32 `json:"port,omitempty"`
	// UseSsl is a bool deciding if ssl will be used
	UseSsl bool `json:"useSsl,omitempty"`
}

// DatabaseServerSpec defines the desired state of DatabaseServer
type DatabaseServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Type is the type of database server. Postgres, mongo etc.
	Type string `json:"type,omitempty"`
	// SecretName is the name of the secret stored in the cluster
	SecretName string   `json:"secretName,omitempty"`
	Postgres   Postgres `json:"postgres,omitempty"`
	Mysql      Mysql    `json:"mysql,omitempty"`
	Mongo      Mongo    `json:"mongo,omitempty"`
}

// DatabaseServerStatus defines the observed state of DatabaseServer
type DatabaseServerStatus struct {
	Connected bool `json:"connected,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DatabaseServer is the Schema for the databaseservers API
type DatabaseServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseServerSpec   `json:"spec,omitempty"`
	Status DatabaseServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseServerList contains a list of DatabaseServer
type DatabaseServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseServer{}, &DatabaseServerList{})
}
