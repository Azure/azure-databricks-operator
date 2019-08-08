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
	"encoding/base64"
	"fmt"
	"time"

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
)

func (r *WorkspaceItemReconciler) submit(instance *databricksv1.WorkspaceItem) error {
	r.Log.Info(fmt.Sprintf("Create item %s", instance.GetName()))

	data, err := base64.StdEncoding.DecodeString(instance.Spec.Content)
	if err != nil {
		return err
	}

	err = r.APIClient.Workspace().Import(instance.Spec.Path, instance.Spec.Format, instance.Spec.Language, data, true)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// Refresh info
	objectInfo, err := r.APIClient.Workspace().GetStatus(instance.Spec.Path)
	if err != nil {
		return err
	}

	instance.Status = &databricksv1.WorkspaceItemStatus{
		ObjectInfo: &objectInfo,
		ObjectHash: instance.GetHash(),
	}

	return r.Update(context.Background(), instance)
}

func (r *WorkspaceItemReconciler) delete(instance *databricksv1.WorkspaceItem) error {
	r.Log.Info(fmt.Sprintf("Deleting item %s", instance.GetName()))

	if instance.Status == nil || instance.Status.ObjectInfo == nil {
		return nil
	}

	path := instance.Status.ObjectInfo.Path

	return r.APIClient.Workspace().Delete(path, true)
}
