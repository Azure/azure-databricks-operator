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
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DclusterStatus represents the status for a Dcluster
type DclusterStatus struct {
	ClusterInfo *DclusterInfo `json:"cluster_info,omitempty"`
}

// +kubebuilder:object:root=true

// Dcluster is the Schema for the dclusters API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="ClusterID",type="string",JSONPath=".status.cluster_info.cluster_id"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.cluster_info.state"
type Dcluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *dbmodels.NewCluster `json:"spec,omitempty"`
	Status *DclusterStatus      `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (dcluster *Dcluster) IsBeingDeleted() bool {
	return !dcluster.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if the item has been submitted to DataBricks
func (dcluster *Dcluster) IsSubmitted() bool {
	if dcluster.Status == nil ||
		dcluster.Status.ClusterInfo == nil ||
		dcluster.Status.ClusterInfo.ClusterID == "" {
		return false
	}
	return true
}

// DclusterFinalizerName is the name of the finalizer for the Dcluster operator
const DclusterFinalizerName = "dcluster.finalizers.databricks.microsoft.com"

// HasFinalizer returns true if the item has the specified finalizer
func (dcluster *Dcluster) HasFinalizer(finalizerName string) bool {
	return containsString(dcluster.ObjectMeta.Finalizers, finalizerName)
}

// AddFinalizer adds the specified finalizer
func (dcluster *Dcluster) AddFinalizer(finalizerName string) {
	dcluster.ObjectMeta.Finalizers = append(dcluster.ObjectMeta.Finalizers, finalizerName)
}

// RemoveFinalizer removes the specified finalizer
func (dcluster *Dcluster) RemoveFinalizer(finalizerName string) {
	dcluster.ObjectMeta.Finalizers = removeString(dcluster.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// DclusterList contains a list of Dcluster
type DclusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dcluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dcluster{}, &DclusterList{})
}
