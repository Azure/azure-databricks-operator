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
	"encoding/base64"
	"fmt"
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	httpmodels "github.com/polar-rams/databricks-sdk-golang/azure/dbfs/httpmodels"
	models "github.com/polar-rams/databricks-sdk-golang/azure/dbfs/models"
)

func (r *DbfsBlockReconciler) submit(instance *databricksv1alpha1.DbfsBlock) error {
	r.Log.Info(fmt.Sprintf("Create block %s", instance.GetName()))

	data, err := base64.StdEncoding.DecodeString(instance.Spec.Data)
	if err != nil {
		return err
	}

	// Open handler
	execution := NewExecution("dbfsblocks", "create")
	createReq := httpmodels.CreateReq{
		Path:      instance.Spec.Path,
		Overwrite: true,
	}
	createResponse, err := r.APIClient.Dbfs().Create(createReq)
	execution.Finish(err)

	if err != nil {
		return err
	}

	// DataBricks limits the AddBlock size to be 1024KB
	var g = 1000
	for i := 0; i < len(data); i += g {
		execution = NewExecution("dbfsblocks", "add_block")

		addBlockReq := httpmodels.AddBlockReq{
			Handle: createResponse.Handle,
		}

		if i+g <= len(data) {
			addBlockReq.Data = string(data[i : i+g])
			err = r.APIClient.Dbfs().AddBlock(addBlockReq)
		} else {
			addBlockReq.Data = string(data[i:])
			err = r.APIClient.Dbfs().AddBlock(addBlockReq)
		}

		execution.Finish(err)

		if err != nil {
			return err
		}
	}

	// Close handler
	execution = NewExecution("dbfsblocks", "close")
	closeReq := httpmodels.CloseReq{
		Handle: createResponse.Handle,
	}
	err = r.APIClient.Dbfs().Close(closeReq)
	execution.Finish(err)

	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// Refresh info
	execution = NewExecution("dbfsblocks", "get_status")
	statusReq := httpmodels.GetStatusReq{
		Path: instance.Spec.Path,
	}
	statusRes, err := r.APIClient.Dbfs().GetStatus(statusReq)
	execution.Finish(err)

	if err != nil {
		return err
	}
	fileInfo := models.FileInfo(statusRes)
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
	deleteReq := httpmodels.DeleteReq{
		Path:      path,
		Recursive: true,
	}
	err := r.APIClient.Dbfs().Delete(deleteReq)
	execution.Finish(err)
	return err
}
