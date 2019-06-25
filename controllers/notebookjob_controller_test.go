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
	"fmt"
	"time"

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Notebookjob Controller", func() {

	var namespacedName = types.NamespacedName{Name: databricksv1.RandomString(10), Namespace: "default"}
	const timeout = time.Second * 60

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
	Context("Create and Delete", func() {
		It("should create and delete real jobs with secrets", func() {
			secrets := make(map[string][]byte)
			secrets["secret1"] = []byte("value1")
			secret1 := &v1.Secret{Data: secrets, ObjectMeta: metav1.ObjectMeta{
				Name: "test-secret", Namespace: "default",
			}}

			instance := &databricksv1.NotebookJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespacedName.Name,
					Namespace: namespacedName.Namespace,
				},
				Spec: databricksv1.NotebookJobSpec{
					NotebookTask: databricksv1.NotebookTask{
						NotebookPath: "/test-notebook",
					},
					NotebookSpec: map[string]string{
						"TestSpec": fmt.Sprintf("%v", time.Now().String()),
					},
					NotebookSpecSecrets: []databricksv1.NotebookSpecSecret{
						{
							SecretName: "test-secret",
							Mapping: []databricksv1.KeyMapping{
								{
									SecretKey: "secret1",
									OutputKey: "SECRET_VALUE",
								},
							},
						},
					},
					NotebookAdditionalLibraries: []databricksv1.NotebookAdditionalLibrary{
						{
							Type: "maven",
							Properties: map[string]string{
								"coordinates": "com.microsoft.azure:azure-eventhubs-spark_2.11:2.3.9",
							},
						},
					},
				},
			}

			// Create the secrets needed
			k8sClient.Create(context.TODO(), secret1)
			defer func() {
				k8sClient.Delete(context.TODO(), secret1)
			}()

			time.Sleep(1 * time.Second)

			// Create the NotebookJob object and expect the Reconcile to be created
			err := k8sClient.Create(context.TODO(), instance)
			// The instance object may not be a valid object because it might be missing some required fields.
			// Please modify the instance object by adding required fields and then remove the following if statement.
			Expect(apierrors.IsInvalid(err)).To(Equal(false))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				_ = k8sClient.Get(context.TODO(), namespacedName, instance)
				return instance.HasFinalizer(finalizerName)
			}, timeout,
			).Should(BeTrue())

			Eventually(func() bool {
				_ = k8sClient.Get(context.TODO(), namespacedName, instance)
				return instance.IsSubmitted()
			}, timeout,
			).Should(BeTrue())

			time.Sleep(10 * time.Second)

			instance2 := &databricksv1.NotebookJob{}
			err = k8sClient.Get(context.TODO(), namespacedName, instance2)
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.Delete(context.TODO(), instance2)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error { return k8sClient.Get(context.TODO(), namespacedName, instance2) }, timeout).
				Should(MatchError("NotebookJob.microsoft.k8s.io \"" + namespacedName.Name + "\" not found"))
		})
	})
})
