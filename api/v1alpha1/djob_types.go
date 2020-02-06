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

// DjobStatus is the status object for the Djob
type DjobStatus struct {
	JobStatus  *dbmodels.Job  `json:"job_status,omitempty"`
	Last10Runs []dbmodels.Run `json:"last_10_runs,omitempty"`
}

// +kubebuilder:object:root=true

// Djob is the Schema for the djobs API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="JobID",type="integer",JSONPath=".status.job_status.job_id"
type Djob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *JobSettings `json:"spec,omitempty"`
	Status *DjobStatus  `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (djob *Djob) IsBeingDeleted() bool {
	return !djob.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if the item has been submitted to DataBricks
func (djob *Djob) IsSubmitted() bool {
	if djob.Status == nil || djob.Status.JobStatus == nil || djob.Status.JobStatus.JobID == 0 {
		return false
	}
	return djob.Status.JobStatus.JobID > 0
}

// DjobFinalizerName is the name of the djob finalizer
const DjobFinalizerName = "djob.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (djob *Djob) HasFinalizer(finalizerName string) bool {
	return containsString(djob.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (djob *Djob) AddFinalizer(finalizerName string) {
	djob.ObjectMeta.Finalizers = append(djob.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
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
