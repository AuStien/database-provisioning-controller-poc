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

// Server is the server on which the database is hosted
type Server struct {
	// Name is the name of the database server
	Name string `json:"name"`
	// Namespace is the namespace of the database server
	Namespace string `json:"namespace"`
}

// Secret is the secret containing credentials
type Secret struct {
	// Name is the name of the secret
	Name string `json:"name"`
	// Namespace is the namespace of the secret
	Namespace string `json:"namespace"`
}

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Server is the namespaced name of databaseServer on which this database is to be created
	Server Server `json:"server"`
	// Name is the name of the database
	Name string `json:"name"`
	// Secret is the secret containing credentials
	Secret Secret `json:"secret"`
	// Username is the username to be assigned to the database (default is name of database)
	Username string `json:"username,omitempty"`
	// +kubebuilder:validation:Enum=delete;retain
	// ReclaimPolicy tells if database will be retained or deleted
	ReclaimPolicy string `json:"reclaimPolicy"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// Secret is the status of secret containg credentials
	CreatedSecret bool `json:"secret,omitempty"`
	// User is status of user on server
	CreatedUser bool `json:"user,omitempty"`
	// DB is status of the new database on server
	CreatedDatabase bool `json:"db,omitempty"`
	// Permissions is status of permissions given to new user
	GrantedPermissions bool `json:"permissions,omitempty"`
	// Connection is status of connection to new database with new user
	Connection bool `json:"connection,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Database Name",type=string,JSONPath=".spec.name",description="name of database"
// +kubebuilder:printcolumn:name="Server",type=string,JSONPath=".spec.server.name",description="name of database server"
// +kubebuilder:printcolumn:name="Reclaim Policy",type=string,JSONPath=".spec.reclaimPolicy",description="reclaim policy"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
