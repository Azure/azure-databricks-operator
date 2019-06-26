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

package v1

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("NotebookJob", func() {

	const timeout = time.Second * 5

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
	Context("Create API", func() {
		It("should create an object successfully", func() {
			key := types.NamespacedName{
				Name:      RandomString(10),
				Namespace: "default",
			}
			created := &NotebookJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				}}

			// Test Create
			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched := &NotebookJob{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})
	})

	Context("API Status", func() {
		It("should return being deleted", func() {
			finalizer := "finalizer.domain.k8s.org"
			finalizers := []string{finalizer}

			key := types.NamespacedName{
				Name:      RandomString(10),
				Namespace: "default",
			}
			notebookjob := &NotebookJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:       key.Name,
					Namespace:  key.Namespace,
					Finalizers: finalizers,
				}}

			defer func() {
				k8sClient.Get(context.TODO(), key, notebookjob)
				notebookjob.RemoveFinalizer(finalizer)
				k8sClient.Update(context.TODO(), notebookjob)
			}()

			k8sClient.Create(context.TODO(), notebookjob)
			k8sClient.Delete(context.TODO(), notebookjob)

			Eventually(func() bool {
				_ = k8sClient.Get(context.TODO(), key, notebookjob)
				return notebookjob.IsBeingDeleted()
			}, timeout,
			).Should(BeTrue())
		})

		It("should return is running", func() {
			notebookJob := &NotebookJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RandomString(10),
					Namespace: "default",
				}}
			Expect(notebookJob.IsSubmitted()).To(BeFalse())

			notebookJob.Status.Run = &dbmodels.Run{
				RunID: int64(rand.Intn(100) + 1),
			}

			Expect(notebookJob.IsSubmitted()).To(BeTrue())
		})
	})
})
