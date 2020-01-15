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

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
)

func (r *DbfsBlockReconciler) submit(instance *databricksv1alpha1.DbfsBlock) error {
	r.Log.Info(fmt.Sprintf("Create block %s", instance.GetName()))

	data, err := base64.StdEncoding.DecodeString(instance.Spec.Data)
	if err != nil {
		return err
	}

	// Open handler
	execution := NewExecution("dbfsblocks", "create")
	createResponse, err := r.APIClient.Dbfs().Create(instance.Spec.Path, true)
	execution.Finish(err)

	if err != nil {
		return err
	}

	// DataBricks limits the AddBlock size to be 1024KB
	var g = 1000
	for i := 0; i < len(data); i += g {
		execution = NewExecution("dbfsblocks", "add_block")

		if i+g <= len(data) {
			err = r.APIClient.Dbfs().AddBlock(createResponse.Handle, data[i:i+g])
		} else {
			err = r.APIClient.Dbfs().AddBlock(createResponse.Handle, data[i:])
		}

		execution.Finish(err)

		if err != nil {
			return err
		}
	}

	// Close handler
	execution = NewExecution("dbfsblocks", "close")
	err = r.APIClient.Dbfs().Close(createResponse.Handle)
	execution.Finish(err)

	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// Refresh info
	execution = NewExecution("dbfsblocks", "get_status")
	fileInfo, err := r.APIClient.Dbfs().GetStatus(instance.Spec.Path)
	execution.Finish(err)

	if err != nil {
		return err
	}

	instance.Status = &databricksv1alpha1.DbfsBlockStatus{
		FileInfo: &fileInfo,
		FileHash: instance.GetHash(),
	}

	return r.Update(context.Background(), instance)
}

func (r *DbfsBlockReconciler) delete(instance *databricksv1alpha1.DbfsBlock) error {
	r.Log.Info(fmt.Sprintf("Deleting block %s", instance.GetName()))

	if instance.Status == nil || instance.Status.FileInfo == nil {
		return nil
	}

	path := instance.Status.FileInfo.Path

	execution := NewExecution("dbfsblocks", "delete")
	err := r.APIClient.Dbfs().Delete(path, true)
	execution.Finish(err)
	return err
}
