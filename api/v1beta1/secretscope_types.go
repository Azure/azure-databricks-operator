/*
Copyright 2019 microsoft.

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

package v1beta1

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
	SecretScope *dbmodels.SecretScope `json:"secretscope,omitempty"`
}

// +kubebuilder:object:root=true

// SecretScope is the Schema for the secretscopes API
type SecretScope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretScopeSpec   `json:"spec,omitempty"`
	Status SecretScopeStatus `json:"status,omitempty"`
}

func (ss *SecretScope) IsSubmitted() bool {
	return ss.Status.SecretScope != nil
}

func (ss *SecretScope) IsBeingDeleted() bool {
	return !ss.ObjectMeta.DeletionTimestamp.IsZero()
}

const SecretScopeFinalizerName = "secretscope.finalizers.databricks.microsoft.com"

func (ss *SecretScope) HasFinalizer(finalizerName string) bool {
	return containsString(ss.ObjectMeta.Finalizers, finalizerName)
}

func (ss *SecretScope) AddFinalizer(finalizerName string) {
	ss.ObjectMeta.Finalizers = append(ss.ObjectMeta.Finalizers, finalizerName)
}

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
