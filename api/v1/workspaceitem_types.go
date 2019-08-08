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

// WorkspaceItemSpec defines the desired state of WorkspaceItem
type WorkspaceItemSpec struct {
	Content  string `json:"contet,omitempty"`
	Path     string `json:"path,omitempty"`
	Language string `json:"language,omitempty"`
	Format   string `json:"format,omitempty"`
}

// WorkspaceItemStatus defines the observed state of WorkspaceItem
type WorkspaceItemStatus struct {
	ObjectInfo *dbmodels.ObjectInfo `json:"object_info,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceItem is the Schema for the workspaceitems API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".status.object_info.path"
// +kubebuilder:printcolumn:name="Language",type="string",JSONPath=".status.object_info.language"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".status.object_info.object_type"
type WorkspaceItem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *WorkspaceItemSpec   `json:"spec,omitempty"`
	Status *WorkspaceItemStatus `json:"status,omitempty"`
}

func (wi *WorkspaceItem) IsBeingDeleted() bool {
	return !wi.ObjectMeta.DeletionTimestamp.IsZero()
}

func (wi *WorkspaceItem) IsSubmitted() bool {
	if wi.Status == nil || wi.Status.ObjectInfo == nil || wi.Status.ObjectInfo.Path == "" {
		return false
	}
	return true
}

const WorkspaceItemFinalizerName = "workspaceitem.finalizers.databricks.microsoft.com"

func (wi *WorkspaceItem) HasFinalizer(finalizerName string) bool {
	return containsString(wi.ObjectMeta.Finalizers, finalizerName)
}

func (wi *WorkspaceItem) AddFinalizer(finalizerName string) {
	wi.ObjectMeta.Finalizers = append(wi.ObjectMeta.Finalizers, finalizerName)
}

func (wi *WorkspaceItem) RemoveFinalizer(finalizerName string) {
	wi.ObjectMeta.Finalizers = removeString(wi.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// WorkspaceItemList contains a list of WorkspaceItem
type WorkspaceItemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceItem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceItem{}, &WorkspaceItemList{})
}
