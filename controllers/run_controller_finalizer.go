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
