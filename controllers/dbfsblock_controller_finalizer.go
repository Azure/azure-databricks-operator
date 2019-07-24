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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
)

func (r *DbfsBlockReconciler) addFinalizer(instance *databricksv1.DbfsBlock) error {
	instance.AddFinalizer(databricksv1.DbfsBlockFinalizerName)
	return r.Update(context.Background(), instance)
}

func (r *DbfsBlockReconciler) handleFinalizer(instance *databricksv1.DbfsBlock) error {
	if !instance.HasFinalizer(databricksv1.DbfsBlockFinalizerName) {
		return nil
	}

	if err := r.delete(instance); err != nil {
		return err
	}
	instance.RemoveFinalizer(databricksv1.DbfsBlockFinalizerName)
	return r.Update(context.Background(), instance)
}