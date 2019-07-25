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

var _ = Describe("Dcluster", func() {
	var (
		key              types.NamespacedName
		created, fetched *Dcluster
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
			created = &Dcluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				}}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &Dcluster{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle isSubmitted", func() {
			dcluster := &Dcluster{
				Status: &DclusterStatus{
					ClusterInfo: &DclusterInfo{
						ClusterID: "221-322-djaj2",
					},
				},
			}
			Expect(dcluster.IsSubmitted()).To(BeTrue())

			dcluster2 := &Dcluster{
				Status: &DclusterStatus{
					ClusterInfo: nil,
				},
			}
			Expect(dcluster2.IsSubmitted()).To(BeFalse())
		})

		It("should correctly handle finalizers", func() {
			dcluster := &Dcluster{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(dcluster.IsBeingDeleted()).To(BeTrue())

			dcluster.AddFinalizer(DclusterFinalizerName)
			Expect(len(dcluster.GetFinalizers())).To(Equal(1))
			Expect(dcluster.HasFinalizer(DclusterFinalizerName)).To(BeTrue())

			dcluster.RemoveFinalizer(DclusterFinalizerName)
			Expect(len(dcluster.GetFinalizers())).To(Equal(0))
			Expect(dcluster.HasFinalizer(DclusterFinalizerName)).To(BeFalse())
		})

		It("should correctly handle float to string", func() {
			clusterInfo := &dbmodels.ClusterInfo{
				ClusterCores: 20.32,
			}

			dclusterInfo := &DclusterInfo{}
			dclusterInfo.FromDataBricksClusterInfo(*clusterInfo)

			Expect(dclusterInfo.ClusterCores).To(Equal("20.32"))
		})
	})

})
