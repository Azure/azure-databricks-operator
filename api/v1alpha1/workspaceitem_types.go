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
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceItemSpec defines the desired state of WorkspaceItem
type WorkspaceItemSpec struct {
	Content  string                `json:"content,omitempty"`
	Path     string                `json:"path,omitempty"`
	Language dbmodels.Language     `json:"language,omitempty"`
	Format   dbmodels.ExportFormat `json:"format,omitempty"`
}

// WorkspaceItemStatus defines the observed state of WorkspaceItem
type WorkspaceItemStatus struct {
	ObjectInfo *dbmodels.ObjectInfo `json:"object_info,omitempty"`
	ObjectHash string               `json:"object_hash,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceItem is the Schema for the workspaceitems API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
// +kubebuilder:printcolumn:name="SHA1SUM",type="string",JSONPath=".status.object_hash",priority=0
// +kubebuilder:printcolumn:name="Language",type="string",JSONPath=".status.object_info.language",priority=0
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".status.object_info.object_type",priority=1
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".status.object_info.path",priority=1
type WorkspaceItem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *WorkspaceItemSpec   `json:"spec,omitempty"`
	Status *WorkspaceItemStatus `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (wi *WorkspaceItem) IsBeingDeleted() bool {
	return !wi.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if the item has been submitted to DataBricks
func (wi *WorkspaceItem) IsSubmitted() bool {
	if wi.Status == nil || wi.Status.ObjectInfo == nil || wi.Status.ObjectInfo.Path == "" {
		return false
	}
	return true
}

// IsUpToDate tells you whether the data is up-to-date with the status
func (wi *WorkspaceItem) IsUpToDate() bool {
	if wi.Status == nil {
		return false
	}
	h := wi.GetHash()
	return h == wi.Status.ObjectHash
}

// GetHash returns the sha1 hash of the decoded data attribute
func (wi *WorkspaceItem) GetHash() string {
	data, err := base64.StdEncoding.DecodeString(wi.Spec.Content)
	if err != nil {
		return ""
	}
	h := sha1.New()
	_, err = h.Write(data)
	if err != nil {
		return ""
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// WorkspaceItemFinalizerName is the name of the workspace item finalizer
const WorkspaceItemFinalizerName = "workspaceitem.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (wi *WorkspaceItem) HasFinalizer(finalizerName string) bool {
	return containsString(wi.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (wi *WorkspaceItem) AddFinalizer(finalizerName string) {
	wi.ObjectMeta.Finalizers = append(wi.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
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
