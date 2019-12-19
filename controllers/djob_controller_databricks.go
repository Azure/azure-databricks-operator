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
	"strings"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DjobReconciler) submit(instance *databricksv1alpha1.Djob) error {
	r.Log.Info(fmt.Sprintf("Submitting job %s", instance.GetName()))
	instance.Spec.Name = instance.GetName()
	//Get exisiting dbricks cluster by cluster name and set ExistingClusterID or
	//Get exisiting dbricks cluster by cluster id
	var ownerInstance databricksv1alpha1.Dcluster
	if len(instance.Spec.ExistingClusterName) > 0 {
		dClusterNamespacedName := types.NamespacedName{Name: instance.Spec.ExistingClusterName, Namespace: instance.Namespace}
		err := r.Get(context.Background(), dClusterNamespacedName, &ownerInstance)
		if err != nil {
			return err
		}
		if (ownerInstance.Status != nil) && (ownerInstance.Status.ClusterInfo != nil) && len(ownerInstance.Status.ClusterInfo.ClusterID) > 0 {
			instance.Spec.ExistingClusterID = ownerInstance.Status.ClusterInfo.ClusterID
		} else {
			return fmt.Errorf("failed to get ClusterID of %v", instance.Spec.ExistingClusterName)
		}
	} else if len(instance.Spec.ExistingClusterID) > 0 {
		var dclusters databricksv1alpha1.DclusterList
		err := r.List(context.Background(), &dclusters, client.InNamespace(instance.Namespace), client.MatchingFields{dclusterIndexKey: instance.Spec.ExistingClusterID})
		if err != nil {
			return err
		}
		if len(dclusters.Items) == 1 {
			ownerInstance = dclusters.Items[0]
		} else {
			return fmt.Errorf("failed to get ClusterID of %v", instance.Spec.ExistingClusterID)
		}
	}
	//Set Exisiting cluster as Owner of JOb
	if &ownerInstance != nil && len(ownerInstance.APIVersion) > 0 && len(ownerInstance.Kind) > 0 && len(ownerInstance.GetName()) > 0 {
		references := []metav1.OwnerReference{
			{
				APIVersion: ownerInstance.APIVersion,
				Kind:       ownerInstance.Kind,
				Name:       ownerInstance.GetName(),
				UID:        ownerInstance.GetUID(),
			},
		}
		instance.ObjectMeta.SetOwnerReferences(references)
	}
	jobSettings := databricksv1alpha1.ToDatabricksJobSettings(instance.Spec)
	job, err := r.createJob(jobSettings)

	if err != nil {
		return err
	}

	instance.Spec.Name = instance.GetName()
	instance.Status = &databricksv1alpha1.DjobStatus{
		JobStatus: &job,
	}
	return r.Update(context.Background(), instance)
}

func (r *DjobReconciler) refresh(instance *databricksv1alpha1.Djob) error {
	r.Log.Info(fmt.Sprintf("Refreshing job %s", instance.GetName()))

	jobID := instance.Status.JobStatus.JobID

	job, err := r.getJob(jobID)

	if err != nil {
		return err
	}

	// Refresh job also needs to get a list of historic runs under this job
	jobRunListResponse, err := r.APIClient.Jobs().RunsList(false, false, jobID, 0, 10)
	if err != nil {
		return err
	}

	err = r.Get(context.Background(), types.NamespacedName{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}, instance)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(instance.Status.JobStatus, &job) &&
		reflect.DeepEqual(instance.Status.Last10Runs, jobRunListResponse.Runs) {
		return nil
	}

	instance.Status = &databricksv1alpha1.DjobStatus{
		JobStatus:  &job,
		Last10Runs: jobRunListResponse.Runs,
	}
	return r.Update(context.Background(), instance)
}

func (r *DjobReconciler) delete(instance *databricksv1alpha1.Djob) error {
	r.Log.Info(fmt.Sprintf("Deleting job %s", instance.GetName()))

	if instance.Status == nil || instance.Status.JobStatus == nil {
		return nil
	}

	jobID := instance.Status.JobStatus.JobID

	// Check if the job exists before trying to delete it
	if _, err := r.getJob(jobID); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}

	execution := NewExecution("djobs", "delete")
	err := r.APIClient.Jobs().Delete(jobID)
	execution.Finish(err)
	return err
}

func (r *DjobReconciler) getJob(jobID int64) (job dbmodels.Job, err error) {
	execution := NewExecution("djobs", "get")
	job, err = r.APIClient.Jobs().Get(jobID)
	execution.Finish(err)
	return job, err
}

func (r *DjobReconciler) createJob(jobSettings dbmodels.JobSettings) (job dbmodels.Job, err error) {
	execution := NewExecution("djobs", "create")
	job, err = r.APIClient.Jobs().Create(jobSettings)
	execution.Finish(err)
	return job, err
}
