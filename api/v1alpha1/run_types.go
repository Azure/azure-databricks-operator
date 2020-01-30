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
