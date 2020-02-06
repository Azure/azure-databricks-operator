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
