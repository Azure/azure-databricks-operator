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

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Djob Controller", func() {

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
	Context("Job with schedule on New Cluster", func() {
		It("Should create successfully", func() {

			testDjobkey := types.NamespacedName{
				Name:      "t-job-with-schedule-new-cluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}

			spec := databricksv1alpha1.JobSettings{
				NewCluster: &dbmodels.NewCluster{
					SparkVersion: "5.3.x-scala2.11",
					NodeTypeID:   "Standard_D3_v2",
					NumWorkers:   2,
				},
				Libraries: []dbmodels.Library{
					{
						Jar: "dbfs:/my-jar.jar",
					},
					{
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

			created := &databricksv1alpha1.Djob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDjobkey.Name,
					Namespace: testDjobkey.Namespace,
				},
				Spec: &spec,
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), testDjobkey, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), testDjobkey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				return k8sClient.Get(context.Background(), testDjobkey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
	Context("Job with schedule on Exsisting Cluster", func() {

		var testDclusterKey types.NamespacedName

		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
			testDclusterKey = types.NamespacedName{
				Name:      "t-cluster" + "-" + randomStringWithCharset(10, charset),
				Namespace: "default",
			}

			dcluster := &databricksv1alpha1.Dcluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDclusterKey.Name,
					Namespace: testDclusterKey.Namespace,
				},
				Spec: &dbmodels.NewCluster{
					Autoscale: &dbmodels.AutoScale{
						MinWorkers: 2,
						MaxWorkers: 3,
					},
					AutoterminationMinutes: 10,
					NodeTypeID:             "Standard_D3_v2",
					SparkVersion:           "5.3.x-scala2.11",
				},
			}

			// Create testDcluster
			_ = k8sClient.Create(context.Background(), dcluster)
			testK8sDcluster := &databricksv1alpha1.Dcluster{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), testDclusterKey, testK8sDcluster)
			}, timeout, interval).Should(Succeed())

		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// Delete test Dcluster
			f := &databricksv1alpha1.Dcluster{}
			_ = k8sClient.Get(context.Background(), testDclusterKey, f)
			_ = k8sClient.Delete(context.Background(), f)
		})

		It("Should create successfully on Exsisting Cluster by name", func() {

			testK8sDcluster := &databricksv1alpha1.Dcluster{}
			By("Expecting Dcluster submitted")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), testDclusterKey, testK8sDcluster)
				return testK8sDcluster.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			Expect(testK8sDcluster.Status).ShouldNot(BeNil())
			Expect(testK8sDcluster.Status.ClusterInfo).ShouldNot(BeNil())
			key := types.NamespacedName{
				Name:      "t-job-with-schedule-exsisting-cluster" + "-" + testK8sDcluster.GetName(),
				Namespace: "default",
			}

			spec := databricksv1alpha1.JobSettings{
				ExistingClusterName: testK8sDcluster.GetName(),
				Libraries: []dbmodels.Library{
					{
						Jar: "dbfs:/my-jar.jar",
					},
					{
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

			created := &databricksv1alpha1.Djob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: &spec,
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), key, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

		})

		It("Should create successfully on Exsisting Cluster using ID", func() {

			testK8sDcluster := &databricksv1alpha1.Dcluster{}
			By("Expecting Dcluster submitted")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), testDclusterKey, testK8sDcluster)
				return testK8sDcluster.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			Expect(testK8sDcluster.Status).ShouldNot(BeNil())
			Expect(testK8sDcluster.Status.ClusterInfo).ShouldNot(BeNil())
			key := types.NamespacedName{
				Name:      "t-job-with-schedule-exsisting-cluster" + "-" + testK8sDcluster.Status.ClusterInfo.ClusterID,
				Namespace: "default",
			}

			spec := databricksv1alpha1.JobSettings{
				ExistingClusterID: testK8sDcluster.Status.ClusterInfo.ClusterID,
				Libraries: []dbmodels.Library{
					{
						Jar: "dbfs:/my-jar.jar",
					},
					{
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

			created := &databricksv1alpha1.Djob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: &spec,
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), key, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1alpha1.Djob{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

		})

	})

})
