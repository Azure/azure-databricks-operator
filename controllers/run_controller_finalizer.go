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

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
)

func (r *RunReconciler) addFinalizer(instance *databricksv1alpha1.Run) error {
	instance.AddFinalizer(databricksv1alpha1.RunFinalizerName)
	return r.Update(context.Background(), instance)
}

// handleFinalizer returns a bool and an error. If error is set then the attempt failed, otherwise boolean indicates whether it completed
func (r *RunReconciler) handleFinalizer(instance *databricksv1alpha1.Run) (bool, error) {
	if !instance.HasFinalizer(databricksv1alpha1.RunFinalizerName) {
		return true, nil
	}

	completed, err := r.delete(instance)
	if err != nil {
		return false, err
	}
	if completed {
		instance.RemoveFinalizer(databricksv1alpha1.RunFinalizerName)
		err := r.Update(context.Background(), instance)
		return err != nil, err
	}
	return false, nil // no error, but indicate not completed to trigger a requeue to delete once cancelled
}
