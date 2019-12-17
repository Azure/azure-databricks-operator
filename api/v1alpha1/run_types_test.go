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

package v1alpha1

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("Run", func() {
	var (
		key              types.NamespacedName
		created, fetched *Run
	)

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
	Context("Create API", func() {

		It("should create an object successfully", func() {

			key = types.NamespacedName{
				Name:      "foo-" + RandomString(5),
				Namespace: "default",
			}
			created = &Run{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				}}

			By("creating an API obj")
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())

			fetched = &Run{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.Background(), created)).To(Succeed())
			Expect(k8sClient.Get(context.Background(), key, created)).ToNot(Succeed())
		})

	})

	It("should correctly handle isSubmitted", func() {
		run := &Run{
			Status: &dbazure.JobsRunsGetOutputResponse{
				Metadata: dbmodels.Run{
					JobID: 23,
				},
			},
		}
		Expect(run.IsSubmitted()).To(BeTrue())

		run2 := &Run{
			Status: nil,
		}
		Expect(run2.IsSubmitted()).To(BeFalse())
	})

	It("should correctly handle finalizers", func() {
		run := &Run{
			ObjectMeta: metav1.ObjectMeta{
				DeletionTimestamp: &metav1.Time{
					Time: time.Now(),
				},
			},
		}
		Expect(run.IsBeingDeleted()).To(BeTrue())

		run.AddFinalizer(RunFinalizerName)
		Expect(len(run.GetFinalizers())).To(Equal(1))
		Expect(run.HasFinalizer(RunFinalizerName)).To(BeTrue())

		run.RemoveFinalizer(RunFinalizerName)
		Expect(len(run.GetFinalizers())).To(Equal(0))
		Expect(run.HasFinalizer(RunFinalizerName)).To(BeFalse())
	})
})
