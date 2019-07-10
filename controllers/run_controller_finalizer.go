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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
)

func (r *RunReconciler) addFinalizer(instance *databricksv1.Run) error {
	instance.AddFinalizer(databricksv1.RunFinalizerName)
	err := r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("failed to update finalizer: %v", err)
	}
	r.Recorder.Event(instance, "Normal", "Updated", fmt.Sprintf("finalizer %s added", databricksv1.RunFinalizerName))
	return nil
}

func (r *RunReconciler) handleFinalizer(instance *databricksv1.Run) error {
	if instance.HasFinalizer(databricksv1.RunFinalizerName) {
		// our finalizer is present, so lets handle our external dependency
		if err := r.deleteExternalDependency(instance); err != nil {
			return err
		}

		instance.RemoveFinalizer(databricksv1.RunFinalizerName)
		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}
	}
	// Our finalizer has finished, so the reconciler can do nothing.
	return nil
}

func (r *RunReconciler) deleteExternalDependency(instance *databricksv1.Run) error {
	if instance.Status != nil {
		r.Log.Info(fmt.Sprintf("Deleting external dependencies (job_id: %d)", instance.Status.Metadata.JobID))
	}
	return r.deleteRunFromDatabricks(instance)
}
