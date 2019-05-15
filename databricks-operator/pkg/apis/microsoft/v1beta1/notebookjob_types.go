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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NotebookTask struct {
	NotebookPath string `json:"notebookPath,omitempty"`
	RunID        int    `json:"runID,omitempty"`
}

type NotebookStream struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type KeyMapping struct {
	SecretKey string `json:"secretKey"`
	OutputKey string `json:"outputKey"`
}
type NotebookSpecSecret struct {
	SecretName string       `json:"secretName"`
	Mapping    []KeyMapping `json:"mapping"`
}

type ClusterSpec struct {
	SparkVersion string `json:"sparkVersion,omitempty"`
	NodeTypeId   string `json:"nodeTypeId,omitempty"`
	NumWorkers   int    `json:"numWorkers,omitempty"`
}

// NotebookJobSpec defines the desired state of NotebookJob
type NotebookJobSpec struct {
	NotebookTask                NotebookTask                `json:"notebookTask,omitempty"`
	TimeoutSeconds              int                         `json:"timeoutSeconds,omitempty"`
	NotebookSpec                map[string]string           `json:"notebookSpec,omitempty"`
	NotebookSpecSecrets         []NotebookSpecSecret        `json:"notebookSpecSecrets,omitempty"`
	NotebookAdditionalLibraries []NotebookAdditionalLibrary `json:"notebookAdditionalLibraries,omitempty"`
	ClusterSpec                 ClusterSpec                 `json:"clusterSpec,omitempty"`
}

type NotebookAdditionalLibrary struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// NotebookJobStatus defines the observed state of NotebookJob
type NotebookJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NotebookJob is the Schema for the notebookjobs API
// +k8s:openapi-gen=true
type NotebookJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotebookJobSpec   `json:"spec,omitempty"`
	Status NotebookJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NotebookJobList contains a list of NotebookJob
type NotebookJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotebookJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NotebookJob{}, &NotebookJobList{})
}

func (nj *NotebookJob) IsBeingDeleted() bool {
	return !nj.ObjectMeta.DeletionTimestamp.IsZero()
}

func (nj *NotebookJob) IsRunning() bool {
	return nj.Spec.NotebookTask.RunID > 0
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

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
