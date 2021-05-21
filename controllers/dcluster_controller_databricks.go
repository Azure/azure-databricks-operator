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
	dbhttpmodels "github.com/polar-rams/databricks-sdk-golang/azure/clusters/httpmodels"
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/clusters/models"
	dbjobsmodels "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

func (r *DclusterReconciler) submit(instance *databricksv1alpha1.Dcluster) error {
	r.Log.Info(fmt.Sprintf("Create cluster %s", instance.GetName()))

	// ClusterName field was removed from Databricks API (https://docs.databricks.com/dev-tools/api/latest/jobs.html#newcluster)
	//instance.Spec.ClusterName = instance.GetName()

	if instance.Status != nil && instance.Status.ClusterInfo != nil && instance.Status.ClusterInfo.ClusterID != "" {
		permanentDeleteReq := dbhttpmodels.PermanentDeleteReq{
			ClusterID: instance.Status.ClusterInfo.ClusterID,
		}
		err := r.APIClient.Clusters().PermanentDelete(permanentDeleteReq)
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
	permanentDeleteReq := dbhttpmodels.PermanentDeleteReq{
		ClusterID: instance.Status.ClusterInfo.ClusterID,
	}
	err := r.APIClient.Clusters().PermanentDelete(permanentDeleteReq)
	execution.Finish(err)
	return err
}

func (r *DclusterReconciler) getCluster(clusterID string) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dclusters", "get")
	getReq := dbhttpmodels.GetReq{
		ClusterID: clusterID,
	}
	getRes, err := r.APIClient.Clusters().Get(getReq)
	execution.Finish(err)

	cluster = mapGetRespToClusterInfo(getRes)
	return cluster, err
}

func (r *DclusterReconciler) createCluster(instance *databricksv1alpha1.Dcluster) (cluster dbmodels.ClusterInfo, err error) {
	execution := NewExecution("dclusters", "create")
	createReq := mapNewClusterToCreateReq(*instance.Spec)
	createRes, err := r.APIClient.Clusters().Create(createReq)
	execution.Finish(err)

	cluster, err = r.getCluster(createRes.ClusterID)
	return cluster, err
}

func mapNewClusterToCreateReq(cluster dbjobsmodels.NewCluster) dbhttpmodels.CreateReq {
	return dbhttpmodels.CreateReq{
		NumWorkers:       cluster.NumWorkers,
		Autoscale:        cluster.Autoscale,
		SparkVersion:     cluster.SparkVersion,
		SparkConf:        *cluster.SparkConf,
		NodeTypeID:       cluster.NodeTypeID,
		DriverNodeTypeID: cluster.DriverNodeTypeID,
		CustomTags:       cluster.CustomTags,
		ClusterLogConf:   cluster.ClusterLogConf,
		InitScripts:      cluster.InitScripts,
		SparkEnvVars:     *cluster.SparkEnvVars,
		InstancePoolID:   cluster.InstancePoolID,
		// ClusterName:            cluster.ClusterName,
		// DockerImage:            cluster.DockerImage,
		// AutoterminationMinutes: cluster.AutoterminationMinutes,
		// IdempotencyToken:       cluster.IdempotencyToken,
		// ApplyPolicyDefVal:      cluster.ApplyPolicyDefVal,
		// EnableLocalDiskEncr:    cluster.EnableLocalDiskEncr,
	}
}

func mapGetRespToClusterInfo(res dbhttpmodels.GetResp) dbmodels.ClusterInfo {
	return dbmodels.ClusterInfo{
		NumWorkers:             res.NumWorkers,
		AutoScale:              &res.AutoScale,
		ClusterID:              res.ClusterID,
		CreatorUserName:        res.CreatorUserName,
		Driver:                 res.Driver,
		Executors:              &res.Executors,
		SparkContextID:         res.SparkContextID,
		JdbcPort:               res.JdbcPort,
		ClusterName:            res.ClusterName,
		SparkVersion:           res.SparkVersion,
		SparkConf:              res.SparkConf,
		NodeTypeID:             res.NodeTypeID,
		DriverNodeTypeID:       res.DriverNodeTypeID,
		ClusterLogConf:         res.ClusterLogConf,
		InitScripts:            &res.InitScripts,
		SparkEnvVars:           res.SparkEnvVars,
		AutoterminationMinutes: res.AutoterminationMinutes,
		State:                  res.State,
		StateMessage:           res.StateMessage,
		StartTime:              res.StartTime,
		TerminateTime:          res.TerminateTime,
		LastStateLossTime:      res.LastStateLossTime,
		LastActivityTime:       res.LastActivityTime,
		ClusterMemoryMb:        res.ClusterMemoryMb,
		ClusterCores:           res.ClusterCores,
		DefaultTags:            res.DefaultTags,
		ClusterLogStatus:       res.ClusterLogStatus,
		TerminationReason:      res.TerminationReason,
	}
}
