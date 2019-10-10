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

package v1beta1

import (
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

var _ = Describe("DbfsBlock", func() {
	var (
		key              types.NamespacedName
		created, fetched *DbfsBlock
	)

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

			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}
			created = &DbfsBlock{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				}}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &DbfsBlock{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle isSubmitted", func() {
			dbfsBlock := &DbfsBlock{
				Status: &DbfsBlockStatus{
					FileInfo: &dbmodels.FileInfo{
						FileSize: 0,
					},
				},
			}
			Expect(dbfsBlock.IsSubmitted()).To(BeFalse())

			dbfsBlock2 := &DbfsBlock{
				Status: &DbfsBlockStatus{
					FileInfo: &dbmodels.FileInfo{
						Path: "/test-path",
					},
				},
			}
			Expect(dbfsBlock2.IsSubmitted()).To(BeTrue())

			dbfsBlock3 := &DbfsBlock{
				Status: &DbfsBlockStatus{
					FileInfo: nil,
				},
			}
			Expect(dbfsBlock3.IsSubmitted()).To(BeFalse())
		})

		It("should correctly handle finalizers", func() {
			dbfsBlock := &DbfsBlock{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(dbfsBlock.IsBeingDeleted()).To(BeTrue())

			dbfsBlock.AddFinalizer(DbfsBlockFinalizerName)
			Expect(len(dbfsBlock.GetFinalizers())).To(Equal(1))
			Expect(dbfsBlock.HasFinalizer(DbfsBlockFinalizerName)).To(BeTrue())

			dbfsBlock.RemoveFinalizer(DbfsBlockFinalizerName)
			Expect(len(dbfsBlock.GetFinalizers())).To(Equal(0))
			Expect(dbfsBlock.HasFinalizer(DbfsBlockFinalizerName)).To(BeFalse())
		})

		It("should correctly handle file hash", func() {
			dbfsBlock := &DbfsBlock{
				Spec: &DbfsBlockSpec{
					Data: "dGVzdA==",
				},
			}

			Expect(dbfsBlock.GetHash()).To(Equal("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"))
			Expect(dbfsBlock.IsUpToDate()).To(BeFalse())

			dbfsBlock.Status = &DbfsBlockStatus{
				FileHash: "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
			}
			Expect(dbfsBlock.IsUpToDate()).To(BeTrue())

			dbfsBlockError := &DbfsBlock{
				Spec: &DbfsBlockSpec{
					Data: "invalid_base64",
				},
			}
			Expect(dbfsBlockError.GetHash()).To(Equal(""))
		})
	})

})
