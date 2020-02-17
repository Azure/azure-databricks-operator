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
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RunSpec defines the desired state of Run
type RunSpec struct {
	// dedicated for job run
	JobName                 string `json:"job_name,omitempty"`
	*dbmodels.RunParameters `json:",inline"`
	// dedicated for direct run
	RunName           string `json:"run_name,omitempty"`
	ClusterSpec       `json:",inline"`
	*dbmodels.JobTask `json:",inline"`
	TimeoutSeconds    int32 `json:"timeout_seconds,omitempty"`
}

// +kubebuilder:object:root=true

// Run is the Schema for the runs API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="RunID",type="integer",JSONPath=".status.metadata.run_id"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.metadata.state.life_cycle_state"
type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *RunSpec                           `json:"spec,omitempty"`
	Status *dbazure.JobsRunsGetOutputResponse `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (run *Run) IsBeingDeleted() bool {
	return !run.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if the item has been submitted to DataBricks
func (run *Run) IsSubmitted() bool {
	if run.Status == nil || run.Status.Metadata.JobID == 0 {
		return false
	}
	return run.Status.Metadata.JobID > 0
}

// IsTerminated return true if item is in terminal state
func (run *Run) IsTerminated() bool {
	if run.Status == nil || run.Status.Metadata.State == nil || run.Status.Metadata.State.LifeCycleState == nil {
		return false
	}
	switch *run.Status.Metadata.State.LifeCycleState {
	case dbmodels.RunLifeCycleStateTerminated, dbmodels.RunLifeCycleStateSkipped, dbmodels.RunLifeCycleStateInternalError:
		return true
	}
	return false
}

// RunFinalizerName is the name of the run finalizer
const RunFinalizerName = "run.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (run *Run) HasFinalizer(finalizerName string) bool {
	return containsString(run.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (run *Run) AddFinalizer(finalizerName string) {
	run.ObjectMeta.Finalizers = append(run.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
func (run *Run) RemoveFinalizer(finalizerName string) {
	run.ObjectMeta.Finalizers = removeString(run.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// RunList contains a list of Run
type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Run{}, &RunList{})
}
