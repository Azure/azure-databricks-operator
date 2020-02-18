/*
The MIT License (MIT)

Copyright (c) 2019  Microsoft

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package v1alpha1

import (
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
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
