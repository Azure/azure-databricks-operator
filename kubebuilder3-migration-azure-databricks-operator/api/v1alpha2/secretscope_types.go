/*
Copyright 2021.

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

package v1alpha2

import (
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/secrets/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretScopeSpec defines the desired state of SecretScope
type SecretScopeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	InitialManagePrincipal string              `json:"initial_manage_permission,omitempty"`
	SecretScopeSecrets     []SecretScopeSecret `json:"secrets,omitempty"`
	SecretScopeACLs        []SecretScopeACL    `json:"acls,omitempty"`
}

// SecretScopeStatus defines the observed state of SecretScope
type SecretScopeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SecretScope              *dbmodels.SecretScope `json:"secretscope,omitempty"`
	SecretInClusterAvailable bool                  `json:"secretinclusteravailable,omitempty"`
}

// +kubebuilder:object:root=true

// SecretScope is the Schema for the secretscopes API
type SecretScope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretScopeSpec   `json:"spec,omitempty"`
	Status SecretScopeStatus `json:"status,omitempty"`
}

// IsSecretAvailable returns true if secret in cluster is available
func (ss *SecretScope) IsSecretAvailable() bool {
	return ss.Status.SecretInClusterAvailable
}

// IsSubmitted returns true if the item has been submitted to DataBricks
func (ss *SecretScope) IsSubmitted() bool {
	return ss.Status.SecretScope != nil
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (ss *SecretScope) IsBeingDeleted() bool {
	return !ss.ObjectMeta.DeletionTimestamp.IsZero()
}

// SecretScopeFinalizerName is the name of the secretscope finalizer
const SecretScopeFinalizerName = "secretscope.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (ss *SecretScope) HasFinalizer(finalizerName string) bool {
	return containsString(ss.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (ss *SecretScope) AddFinalizer(finalizerName string) {
	ss.ObjectMeta.Finalizers = append(ss.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
func (ss *SecretScope) RemoveFinalizer(finalizerName string) {
	ss.ObjectMeta.Finalizers = removeString(ss.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// SecretScopeList contains a list of SecretScope
type SecretScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretScope `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretScope{}, &SecretScopeList{})
}