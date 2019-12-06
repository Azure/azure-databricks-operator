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

package controllers

import (
	"context"
	"fmt"
	"reflect"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

func (r *DclusterReconciler) submit(instance *databricksv1alpha1.Dcluster) error {
	r.Log.Info(fmt.Sprintf("Create cluster %s", instance.GetName()))

	instance.Spec.ClusterName = instance.GetName()

	if instance.Status != nil && instance.Status.ClusterInfo != nil && instance.Status.ClusterInfo.ClusterID != "" {
		err := r.APIClient.Clusters().PermanentDelete(instance.Status.ClusterInfo.ClusterID)
		if err != nil {
			return err
		}
	}

	clusterInfo, err := r.createCluster(instance)
	if err != nil {
		return err
	}

	var info databricksv1alpha1.DclusterInfo
	instance.Status = &databricksv1alpha1.DclusterStatus{
		ClusterInfo: info.FromDataBricksClusterInfo(clusterInfo),
	}
	return r.Update(context.Background(), instance)
}

func (r *DclusterReconciler) refresh(instance *databricksv1alpha1.Dcluster) error {
	r.Log.Info(fmt.Sprintf("Refresh cluster %s", instance.GetName()))

	if instance.Status == nil || instance.Status.ClusterInfo == nil {
		return nil
	}

	clusterInfo, err := r.getCluster(instance.Status.ClusterInfo.ClusterID)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(instance.Status.ClusterInfo, &clusterInfo) {
		return nil
	}

	var info databricksv1alpha1.DclusterInfo
	instance.Status = &databricksv1alpha1.DclusterStatus{
		ClusterInfo: info.FromDataBricksClusterInfo(clusterInfo),
	}
	return r.Update(context.Background(), instance)
}

func (r *DclusterReconciler) delete(instance *databricksv1alpha1.Dcluster) error {
	r.Log.Info(fmt.Sprintf("Deleting cluster %s", instance.GetName()))

	if instance.Status == nil || instance.Status.ClusterInfo == nil {
		return nil
	}

	execution := NewExecution("dcluster", "delete")
	err := r.APIClient.Clusters().PermanentDelete(instance.Status.ClusterInfo.ClusterID)
	execution.Finish(err)
	return err
}

func (r *DclusterReconciler) getCluster(clusterID string) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dcluster", "get")
	cluster, err = r.APIClient.Clusters().Get(clusterID)
	execution.Finish(err)
	return cluster, err
}

func (r *DclusterReconciler) createCluster(instance *databricksv1alpha1.Dcluster) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dcluster", "create")
	cluster, err = r.APIClient.Clusters().Create(*instance.Spec)
	execution.Finish(err)
	return cluster, err
}
