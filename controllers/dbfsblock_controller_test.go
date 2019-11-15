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
	"crypto/rand"
	"encoding/base64"
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("DbfsBlock Controller", func() {

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
	Context("Block greater than 1MB", func() {
		It("Should create successfully", func() {

			data := make([]byte, 5000)
			rand.Read(data)
			dataStr := base64.StdEncoding.EncodeToString(data)

			data2 := make([]byte, 5500)
			rand.Read(data2)
			dataStr2 := base64.StdEncoding.EncodeToString(data2)

			key := types.NamespacedName{
				Name:      "block-greater-than-1mb",
				Namespace: "default",
			}

			created := &databricksv1alpha1.DbfsBlock{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: &databricksv1alpha1.DbfsBlockSpec{
					Path: "/some-path/test-block",
					Data: dataStr,
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &databricksv1alpha1.DbfsBlock{}
				k8sClient.Get(context.Background(), key, f)
				return f.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Expecting size to be 5000")
			Eventually(func() int64 {
				f := &databricksv1alpha1.DbfsBlock{}
				k8sClient.Get(context.Background(), key, f)
				return f.Status.FileInfo.FileSize
			}, timeout, interval).Should(Equal(int64(5000)))

			// Update
			updated := &databricksv1alpha1.DbfsBlock{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec.Data = dataStr2
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			By("Expecting size to be 5500")
			Eventually(func() int64 {
				f := &databricksv1alpha1.DbfsBlock{}
				k8sClient.Get(context.Background(), key, f)
				return f.Status.FileInfo.FileSize
			}, timeout, interval).Should(Equal(int64(5500)))

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &databricksv1alpha1.DbfsBlock{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &databricksv1alpha1.DbfsBlock{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
