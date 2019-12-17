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
