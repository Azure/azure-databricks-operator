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
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbfsBlockSpec defines the desired state of DbfsBlock
type DbfsBlockSpec struct {
	Path string `json:"path,omitempty"`
	Data string `json:"data,omitempty"`
}

// DbfsBlockStatus defines the observed state of DbfsBlock
type DbfsBlockStatus struct {
	FileInfo *dbmodels.FileInfo `json:"file_info,omitempty"`
	FileHash string             `json:"file_hash,omitempty"`
}

// +kubebuilder:object:root=true

// DbfsBlock is the Schema for the dbfsblocks API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SHA1SUM",type="string",JSONPath=".status.file_hash"
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".status.file_info.path"
// +kubebuilder:printcolumn:name="Size",type="integer",JSONPath=".status.file_info.file_size"
type DbfsBlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *DbfsBlockSpec   `json:"spec,omitempty"`
	Status *DbfsBlockStatus `json:"status,omitempty"`
}

func (dbfsBlock *DbfsBlock) IsBeingDeleted() bool {
	return !dbfsBlock.ObjectMeta.DeletionTimestamp.IsZero()
}

func (dbfsBlock *DbfsBlock) IsSubmitted() bool {
	if dbfsBlock.Status == nil ||
		dbfsBlock.Status.FileInfo == nil ||
		dbfsBlock.Status.FileInfo.Path == "" {
		return false
	}
	return true
}

// IsUpToDate tells you whether the data is up-to-date with the status
func (dbfsBlock *DbfsBlock) IsUpToDate() bool {
	h := dbfsBlock.GetHash()
	return h == dbfsBlock.Status.FileHash
}

// GetHash returns the sha1 hash of the decoded data attribute
func (dbfsBlock *DbfsBlock) GetHash() string {
	data, err := base64.StdEncoding.DecodeString(dbfsBlock.Spec.Data)
	if err != nil {
		return ""
	}
	h := sha1.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

const DbfsBlockFinalizerName = "dbfsBlock.finalizers.databricks.microsoft.com"

func (dbfsBlock *DbfsBlock) HasFinalizer(finalizerName string) bool {
	return containsString(dbfsBlock.ObjectMeta.Finalizers, finalizerName)
}

func (dbfsBlock *DbfsBlock) AddFinalizer(finalizerName string) {
	dbfsBlock.ObjectMeta.Finalizers = append(dbfsBlock.ObjectMeta.Finalizers, finalizerName)
}

func (dbfsBlock *DbfsBlock) RemoveFinalizer(finalizerName string) {
	dbfsBlock.ObjectMeta.Finalizers = removeString(dbfsBlock.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// DbfsBlockList contains a list of DbfsBlock
type DbfsBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DbfsBlock `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DbfsBlock{}, &DbfsBlockList{})
}
