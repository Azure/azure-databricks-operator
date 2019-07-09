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

package v1

import (
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DjobSpec defines the desired state of Djob
type DjobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	*dbmodels.JobSettings `json:",inline"`
}

// DjobStatus defines the observed state of Djob
type DjobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Job *dbmodels.Job `json:"job"`
}

// +kubebuilder:object:root=true

// Djob is the Schema for the djobs API
type Djob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DjobSpec   `json:"spec,omitempty"`
	Status DjobStatus `json:"status,omitempty"`
}

func (djob *Djob) IsBeingDeleted() bool {
	return !djob.ObjectMeta.DeletionTimestamp.IsZero()
}

func (djob *Djob) IsSubmitted() bool {
	if djob.Status.Job == nil {
		return false
	}
	return djob.Status.Job.JobID > 0
}

func (djob *Djob) HasFinalizer(finalizerName string) bool {
	return containsString(djob.ObjectMeta.Finalizers, finalizerName)
}

func (djob *Djob) AddFinalizer(finalizerName string) {
	djob.ObjectMeta.Finalizers = append(djob.ObjectMeta.Finalizers, finalizerName)
}

func (djob *Djob) RemoveFinalizer(finalizerName string) {
	djob.ObjectMeta.Finalizers = removeString(djob.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// DjobList contains a list of Djob
type DjobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Djob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Djob{}, &DjobList{})
}
