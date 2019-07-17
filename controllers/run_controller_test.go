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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Run Controller", func() {

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
	Context("Run directly without existing job", func() {
		It("Should create successfully", func() {
			Expect(1).To(Equal(1))
		})
	})

	Context("Run existing job", func() {
		It("Should create successfully", func() {
			By("Create job for run")
			jobKey := types.NamespacedName{
				Name:      "integreation-test-job-for-run",
				Namespace: "default",
			}

			jobSpec := &dbmodels.JobSettings{
				NewCluster: &dbmodels.NewCluster{
					SparkVersion: "5.3.x-scala2.11",
					NodeTypeID:   "Standard_D3_v2",
					NumWorkers:   10,
				},
				Libraries: []dbmodels.Library{
					dbmodels.Library{
						Jar: "dbfs:/my-jar.jar",
					},
					dbmodels.Library{
						Maven: &dbmodels.MavenLibrary{
							Coordinates: "org.jsoup:jsoup:1.7.2",
						},
					},
				},
				TimeoutSeconds: 3600,
				MaxRetries:     1,
				SparkJarTask: &dbmodels.SparkJarTask{
					MainClassName: "com.databricks.ComputeModels",
				},
			}

			job := &databricksv1.Djob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobKey.Name,
					Namespace: jobKey.Namespace,
				},
				Spec: jobSpec,
			}

			Expect(k8sClient.Create(context.Background(), job)).Should(Succeed())
			time.Sleep(time.Second * 5)
			defer func() {
				Expect(k8sClient.Delete(context.Background(), job)).Should(Succeed())
				time.Sleep(time.Second * 5)
			}()

			By("Create the run itself")
			runKey := types.NamespacedName{
				Name:      "integreation-test-job-for-run-run",
				Namespace: "default",
			}

			runSpec := &databricksv1.RunSpec{
				JobName: jobKey.Name,
				RunParameters: &dbmodels.RunParameters{
					JarParams: []string{"test"},
				},
			}

			run := &databricksv1.Run{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runKey.Name,
					Namespace: runKey.Namespace,
				},
				Spec: runSpec,
			}

			Expect(k8sClient.Create(context.Background(), run)).Should(Succeed())

			// Create
			By("Expecting run to be submitted")
			Eventually(func() bool {
				f := &databricksv1.Run{}
				k8sClient.Get(context.Background(), runKey, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			time.Sleep(time.Second * 5)

			// Delete
			By("Expecting run to be deleted successfully")
			Eventually(func() error {
				f := &databricksv1.Run{}
				k8sClient.Get(context.Background(), runKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting run to be deleted true")
			Eventually(func() bool {
				f := &databricksv1.Run{}
				k8sClient.Get(context.Background(), runKey, f)
				return f.IsBeingDeleted()
			}, timeout, interval).Should(BeTrue())

			By("Expecting run finaliser be removed")
			Eventually(func() bool {
				f := &databricksv1.Run{}
				k8sClient.Get(context.Background(), runKey, f)
				return f.HasFinalizer(databricksv1.RunFinalizerName)
			}, timeout, interval).Should(BeFalse())
		})
	})
})
