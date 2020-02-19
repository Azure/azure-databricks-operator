/*
The MIT License (MIT)

Copyright (c) 2019  Microsoft

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

	execution := NewExecution("dclusters", "delete")
	err := r.APIClient.Clusters().PermanentDelete(instance.Status.ClusterInfo.ClusterID)
	execution.Finish(err)
	return err
}

func (r *DclusterReconciler) getCluster(clusterID string) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dclusters", "get")
	cluster, err = r.APIClient.Clusters().Get(clusterID)
	execution.Finish(err)
	return cluster, err
}

func (r *DclusterReconciler) createCluster(instance *databricksv1alpha1.Dcluster) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dclusters", "create")
	cluster, err = r.APIClient.Clusters().Create(*instance.Spec)
	execution.Finish(err)
	return cluster, err
}
