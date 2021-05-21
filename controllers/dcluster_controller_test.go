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
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbcltrsmodels "github.com/polar-rams/databricks-sdk-golang/azure/clusters/models"
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Dcluster Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Cluster with autho-scaling", func() {
		It("Should create successfully", func() {

			key := types.NamespacedName{
				Name:      "t-cluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}

			created := &databricksv1alpha1.Dcluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: &dbmodels.NewCluster{
					Autoscale: &dbcltrsmodels.AutoScale{
						MinWorkers: 2,
						MaxWorkers: 5,
					},
					// AutoterminationMinutes: 10,
					NodeTypeID:   "Standard_D3_v2",
					SparkVersion: "5.3.x-scala2.11",
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1alpha1.Dcluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1alpha1.Dcluster{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1alpha1.Dcluster{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
