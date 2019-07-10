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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
)

func (r *DjobReconciler) submitJobToDatabricks(instance *databricksv1.Djob) error {
	r.Log.Info("Submitting job " + instance.GetName())

	instance.Spec.Name = instance.GetName()
	job, err := r.APIClient.Jobs().Create(*instance.Spec)
	if err != nil {
		return err
	}
	if job.JobID == 0 {
		return fmt.Errorf("result from API didn't return any values")
	}

	instance.Status = &job
	err = r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("error when updating job after submitting to API: %v", err)
	}

	r.Recorder.Event(instance, "Normal", "Updated", "jobID added")
	return nil
}

func (r *DjobReconciler) refreshDatabricksJob(instance *databricksv1.Djob) error {
	r.Log.Info("Refreshing job " + instance.GetName())

	jobID := instance.Status.JobID
	job, err := r.APIClient.Jobs().Get(jobID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(instance.Status, &job) {
		instance.Status = &job
		err = r.Update(context.Background(), instance)
		if err != nil {
			return fmt.Errorf("error when updating job: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Updated", "job status updated")
	}
	return nil
}

func (r *DjobReconciler) deleteJobFromDatabricks(instance *databricksv1.Djob) error {
	r.Log.Info("Deleting job " + instance.GetName())

	if instance.Status == nil {
		return nil
	}
	jobID := instance.Status.JobID
	_, err := r.APIClient.Jobs().Get(jobID)
	if err == nil {
		err = r.APIClient.Jobs().Delete(jobID)
	}
	if err == nil || strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}
