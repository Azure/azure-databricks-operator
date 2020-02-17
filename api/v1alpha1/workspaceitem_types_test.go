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

package v1alpha1

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

var _ = Describe("WorkspaceItem", func() {
	var (
		key              types.NamespacedName
		created, fetched *WorkspaceItem
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
			created = &WorkspaceItem{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				}}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &WorkspaceItem{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle isSubmitted", func() {
			wiItem := &WorkspaceItem{
				Status: &WorkspaceItemStatus{
					ObjectInfo: &dbmodels.ObjectInfo{
						Path: "",
					},
				},
			}
			Expect(wiItem.IsSubmitted()).To(BeFalse())

			wiItem2 := &WorkspaceItem{
				Status: &WorkspaceItemStatus{
					ObjectInfo: &dbmodels.ObjectInfo{
						Path: "/test-path",
					},
				},
			}
			Expect(wiItem2.IsSubmitted()).To(BeTrue())

			wiItem3 := &WorkspaceItem{
				Status: &WorkspaceItemStatus{
					ObjectInfo: nil,
				},
			}
			Expect(wiItem3.IsSubmitted()).To(BeFalse())
		})

		It("should correctly handle finalizers", func() {
			wiItem := &WorkspaceItem{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(wiItem.IsBeingDeleted()).To(BeTrue())

			wiItem.AddFinalizer(WorkspaceItemFinalizerName)
			Expect(len(wiItem.GetFinalizers())).To(Equal(1))
			Expect(wiItem.HasFinalizer(WorkspaceItemFinalizerName)).To(BeTrue())

			wiItem.RemoveFinalizer(WorkspaceItemFinalizerName)
			Expect(len(wiItem.GetFinalizers())).To(Equal(0))
			Expect(wiItem.HasFinalizer(WorkspaceItemFinalizerName)).To(BeFalse())
		})

		It("should correctly handle file hash", func() {
			wiItem := &WorkspaceItem{
				Spec: &WorkspaceItemSpec{
					Content: "dGVzdA==",
				},
			}

			Expect(wiItem.GetHash()).To(Equal("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"))
			Expect(wiItem.IsUpToDate()).To(BeFalse())

			wiItem.Status = &WorkspaceItemStatus{
				ObjectHash: "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
			}
			Expect(wiItem.IsUpToDate()).To(BeTrue())

			wiItemError := &WorkspaceItem{
				Spec: &WorkspaceItemSpec{
					Content: "invalid_base64",
				},
			}
			Expect(wiItemError.GetHash()).To(Equal(""))
		})

	})

})
