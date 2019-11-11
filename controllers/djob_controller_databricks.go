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
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	models "github.com/xinsnake/databricks-sdk-golang/azure/models"
	"k8s.io/apimachinery/pkg/types"
)

func (r *DjobReconciler) submit(instance *databricksv1alpha1.Djob) error {
	r.Log.Info(fmt.Sprintf("Submitting job %s", instance.GetName()))

	instance.Spec.Name = instance.GetName()

	job, err := createJob(r, instance)

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

	job, err := getJob(r, jobID)

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
	if _, err := getJob(r, jobID); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}

	return trackExecutionTime(djobDeleteDuration, func() error {
		return r.APIClient.Jobs().Delete(jobID)
	})
}

func getJob(r *DjobReconciler, jobID int64) (job models.Job, err error) {
	defer trackMillisecondsTaken(time.Now(), djobGetDuration)

	job, err = r.APIClient.Jobs().Get(jobID)

	trackSuccessFailure(err, djobGetSuccess, djobGetFailure)

	return job, err
}

func createJob(r *DjobReconciler, instance *databricksv1alpha1.Djob) (job models.Job, err error) {
	defer trackMillisecondsTaken(time.Now(), djobCreateDuration)

	job, err = r.APIClient.Jobs().Create(*instance.Spec)

	trackSuccessFailure(err, djobCreateSuccess, djobCreateFailure)

	return job, err
}
