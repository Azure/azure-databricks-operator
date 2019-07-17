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

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Djob Controller", func() {

	const timeout = time.Second * 15
	const interval = time.Millisecond * 200

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
	Context("Job with schedule", func() {
		It("Should create successfully", func() {

			key := types.NamespacedName{
				Name:      "integreation-test-job-with-schedule",
				Namespace: "default",
			}

			spec := &dbmodels.JobSettings{
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
				Schedule: &dbmodels.CronSchedule{
					QuartzCronExpression: "0 15 22 ? * *",
					TimezoneID:           "America/Los_Angeles",
				},
				SparkJarTask: &dbmodels.SparkJarTask{
					MainClassName: "com.databricks.ComputeModels",
				},
			}

			created := &databricksv1.Djob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			// Create
			k8sClient.Create(context.TODO(), created)

			fetched := &databricksv1.Djob{}
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Delete
			Eventually(func() error {
				k8sClient.Get(context.TODO(), key, fetched)
				return k8sClient.Delete(context.TODO(), fetched)
			}, timeout, interval).Should(Succeed())
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), key, fetched)
				return fetched.IsBeingDeleted()
			}, timeout, interval).Should(BeTrue())
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), key, fetched)
				return fetched.HasFinalizer(databricksv1.DjobFinalizerName)
			}, timeout, interval).Should(BeFalse())
		})
	})
})
