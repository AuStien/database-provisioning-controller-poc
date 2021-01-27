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
	Name string `json:"name,omitempty"`
	// Namespace is the namespace of the database server
	Namespace string `json:"namespace,omitempty"`
}

// Secret is the secret containing credentials
type Secret struct {
	// Name is the name of the secret
	Name string `json:"name,omitempty"`
	// Namespace is the namespace of the secret
	Namespace string `json:"namespace,omitempty"`
}

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Server is the namespaced name of databaseServer on which this database is to be created
	Server Server `json:"server,omitempty"`
	// Name is the name of the database
	Name string `json:"name,omitempty"`
	// Secret is the secret containing credentials
	Secret Secret `json:"secret,omitempty"`
	// Username is the username to be assigned to the database (default is name of database)
	Username string `json:"username,omitempty"`
	// Deletable is if database is able to be deleted
	Deletable bool `json:"deletable,omitempty"`
	// MigrationURL provides URL to the migration
	MigrationURL string `json:"migrationURL,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// Secret is the status of secret containg credentials
	Secret string `json:"secret,omitempty"`
	// User is status of user on server
	User string `json:"user,omitempty"`
	// DB is status of the new database on server
	DB string `json:"db,omitempty"`
	// Permissions is status of permissions given to new user
	Permissions string `json:"permissions,omitempty"`
	// Connection is status of connection to new database with new user
	Connection bool `json:"connection,omitempty"`
	// Migrated is database has been migrated
	Migrated bool `json:"migrated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
