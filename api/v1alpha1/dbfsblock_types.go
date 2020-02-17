/*
The MIT License (MIT)

Copyright (c) 2019 Microsoft

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

// IsBeingDeleted returns true if a deletion timestamp is set
func (dbfsBlock *DbfsBlock) IsBeingDeleted() bool {
	return !dbfsBlock.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if the item has been submitted to DataBricks
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
	if dbfsBlock.Status == nil {
		return false
	}
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
	_, err = h.Write(data)
	if err != nil {
		return ""
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// DbfsBlockFinalizerName is the name of the dbfs block finalizer
const DbfsBlockFinalizerName = "dbfsBlock.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (dbfsBlock *DbfsBlock) HasFinalizer(finalizerName string) bool {
	return containsString(dbfsBlock.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (dbfsBlock *DbfsBlock) AddFinalizer(finalizerName string) {
	dbfsBlock.ObjectMeta.Finalizers = append(dbfsBlock.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
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
