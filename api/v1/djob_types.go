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

type DjobStatus struct {
	JobStatus  *dbmodels.Job  `json:"job_status,omitempty"`
	Last10Runs []dbmodels.Run `json:"last_10_runs,omitempty"`
}

// +kubebuilder:object:root=true

// Djob is the Schema for the djobs API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".status.created_time"
// +kubebuilder:printcolumn:name="JobID",type="integer",JSONPath=".status.job_id"
type Djob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *dbmodels.JobSettings `json:"spec,omitempty"`
	Status *DjobStatus           `json:"status,omitempty"`
}

func (djob *Djob) IsBeingDeleted() bool {
	return !djob.ObjectMeta.DeletionTimestamp.IsZero()
}

func (djob *Djob) IsSubmitted() bool {
	if djob.Status == nil || djob.Status.JobStatus.JobID == 0 {
		return false
	}
	return djob.Status.JobStatus.JobID > 0
}

const DjobFinalizerName = "djob.finalizers.databricks.microsoft.com"

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
