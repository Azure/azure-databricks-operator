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

// NotebookJobSpec defines the desired state of NotebookJob
type NotebookJobSpec struct {
	NotebookTask                NotebookTask                `json:"notebookTask,omitempty"`
	TimeoutSeconds              int                         `json:"timeoutSeconds,omitempty"`
	NotebookSpec                map[string]string           `json:"notebookSpec,omitempty"`
	NotebookSpecSecrets         []NotebookSpecSecret        `json:"notebookSpecSecrets,omitempty"`
	NotebookAdditionalLibraries []NotebookAdditionalLibrary `json:"notebookAdditionalLibraries,omitempty"`
	ClusterSpec                 ClusterSpec                 `json:"clusterSpec,omitempty"`
}

// NotebookJobStatus defines the observed state of NotebookJob
type NotebookJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run *dbmodels.Run `json:"run,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// NotebookJob is the Schema for the notebookjobs API
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=nbj,path=notebookjobs
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="JobID",type="integer",JSONPath=".status.run.job_id"
// +kubebuilder:printcolumn:name="RunID",type="integer",JSONPath=".status.run.run_id"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.run.state.life_cycle_state"
// +kubebuilder:printcolumn:name="NotebookPath",type="string",JSONPath=".status.run.task.notebook_task.notebook_path"
type NotebookJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotebookJobSpec   `json:"spec,omitempty"`
	Status NotebookJobStatus `json:"status,omitempty"`
}

func (nj *NotebookJob) IsBeingDeleted() bool {
	return !nj.ObjectMeta.DeletionTimestamp.IsZero()
}

func (nj *NotebookJob) IsSubmitted() bool {
	if nj.Status.Run == nil {
		return false
	}
	return nj.Status.Run.RunID > 0
}

func (nj *NotebookJob) HasFinalizer(finalizerName string) bool {
	return containsString(nj.ObjectMeta.Finalizers, finalizerName)
}

func (nj *NotebookJob) AddFinalizer(finalizerName string) {
	nj.ObjectMeta.Finalizers = append(nj.ObjectMeta.Finalizers, finalizerName)
}

func (nj *NotebookJob) RemoveFinalizer(finalizerName string) {
	nj.ObjectMeta.Finalizers = removeString(nj.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// NotebookJobList contains a list of NotebookJob
type NotebookJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotebookJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NotebookJob{}, &NotebookJobList{})
}
