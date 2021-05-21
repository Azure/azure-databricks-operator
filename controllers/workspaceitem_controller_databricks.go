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
	"fmt"
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"

	dbhttpmodels "github.com/polar-rams/databricks-sdk-golang/azure/workspace/httpmodels"
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/workspace/models"
)

func (r *WorkspaceItemReconciler) submit(instance *databricksv1alpha1.WorkspaceItem) error {
	r.Log.Info(fmt.Sprintf("Create item %s", instance.GetName()))

	if instance.Spec == nil || len(instance.Spec.Content) <= 0 {
		return fmt.Errorf("Workspace Content is empty")
	}

	execution := NewExecution("workspaceitems", "import")
	importReq := dbhttpmodels.ImportReq{
		Path:      instance.Spec.Path,
		Language:  instance.Spec.Language,
		Format:    instance.Spec.Format,
		Content:   instance.Spec.Content,
		Overwrite: true,
	}
	err := r.APIClient.Workspace().Import(importReq)
	execution.Finish(err)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// Refresh info
	execution = NewExecution("workspaceitems", "get_status")
	getStatusReq := dbhttpmodels.GetStatusReq{
		Path: instance.Spec.Path,
	}
	getStatusRes, err := r.APIClient.Workspace().GetStatus(getStatusReq)
	execution.Finish(err)
	if err != nil {
		return err
	}

	objectInfo := dbmodels.ObjectInfo(getStatusRes)
	instance.Status = &databricksv1alpha1.WorkspaceItemStatus{
		ObjectInfo: &objectInfo,
		ObjectHash: instance.GetHash(),
	}

	return r.Update(context.Background(), instance)
}

func (r *WorkspaceItemReconciler) delete(instance *databricksv1alpha1.WorkspaceItem) error {
	r.Log.Info(fmt.Sprintf("Deleting item %s", instance.GetName()))

	if instance.Status == nil || instance.Status.ObjectInfo == nil {
		return nil
	}

	path := instance.Status.ObjectInfo.Path

	execution := NewExecution("workspaceitems", "delete")
	deleteReq := dbhttpmodels.DeleteReq{
		Path:      path,
		Recursive: true,
	}
	err := r.APIClient.Workspace().Delete(deleteReq)
	execution.Finish(err)
	return err
}
