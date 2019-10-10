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
	"time"

	databricksv1beta1 "github.com/microsoft/azure-databricks-operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("WorkspaceItem Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Workspace Item", func() {
		It("Should create successfully", func() {

			key := types.NamespacedName{
				Name:      "block-greater-than-1mb",
				Namespace: "default",
			}

			created := &databricksv1beta1.WorkspaceItem{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: &databricksv1beta1.WorkspaceItemSpec{
					Content:  "MSsx",
					Path:     "/test-notebook",
					Language: "SCALA",
					Format:   "SOURCE",
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1beta1.WorkspaceItem{}
				k8sClient.Get(context.Background(), key, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &databricksv1beta1.WorkspaceItem{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec.Content = "MSsy"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1beta1.WorkspaceItem{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1beta1.WorkspaceItem{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
