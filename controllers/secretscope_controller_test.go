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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("SecretScope Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1
	const charset = "abcdefghijklmnopqrstuvwxyz"

	var aclKeyName = "t-secretscope-with-acls" + randomStringWithCharset(10, charset)
	var secretsKeyName = "t-secretscope-with-secrets" + randomStringWithCharset(10, charset)

	BeforeEach(func() {
		// failed test runs that don't clean up leave resources behind.
		keys := []string{aclKeyName, secretsKeyName}
		for _, value := range keys {
			_ = apiClient.Secrets().DeleteSecretScope(value)
		}
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Secret Scope with ACLs", func() {
		It("Should handle scope and ACLs correctly", func() {
			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     make([]databricksv1alpha1.SecretScopeSecret, 0),
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					{Principal: "admins", Permission: "WRITE"},
					{Principal: "admins", Permission: "READ"},
					{Principal: "admins", Permission: "MANAGE"},
				},
			}

			key := types.NamespacedName{
				Name:      aclKeyName,
				Namespace: "default",
			}

			toCreate := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating the scope with ACLs successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &databricksv1alpha1.SecretScope{}
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating ACLs successfully")
			updatedACLs := []databricksv1alpha1.SecretScopeACL{
				{Principal: "admins", Permission: "READ"},
			}

			updateSpec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     make([]databricksv1alpha1.SecretScopeSecret, 0),
				SecretScopeACLs:        updatedACLs,
			}

			fetched.Spec = updateSpec

			Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
			fetchedUpdated := &databricksv1alpha1.SecretScope{}
			Eventually(func() []databricksv1alpha1.SecretScopeACL {
				_ = k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeACLs
			}, timeout, interval).Should(Equal(updatedACLs))

			By("Deleting the scope")
			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Secret Scope with secrets", func() {
		It("Should handle scope and secrets correctly", func() {

			// setup k8s secret
			k8SecretKey := types.NamespacedName{
				Name:      "t-k8secret",
				Namespace: "default",
			}

			data := make(map[string][]byte)
			data["username"] = []byte("Josh")
			k8Secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      k8SecretKey.Name,
					Namespace: k8SecretKey.Namespace,
				},
				Data: data,
			}
			Expect(k8sClient.Create(context.Background(), k8Secret)).Should(Succeed())
			time.Sleep(time.Second * 8)
			defer func() {
				Expect(k8sClient.Delete(context.Background(), k8Secret)).Should(Succeed())
				time.Sleep(time.Second * 5)
			}()

			secretValue := "secretValue"
			byteSecretValue := "aGVsbG8="
			initialSecrets := []databricksv1alpha1.SecretScopeSecret{
				{Key: "secretKey", StringValue: secretValue},
				{
					Key: "secretFromSecret",
					ValueFrom: &databricksv1alpha1.SecretScopeValueFrom{
						SecretKeyRef: databricksv1alpha1.SecretScopeKeyRef{
							Name: "t-k8secret",
							Key:  "username",
						},
					},
				},
				{Key: "byteSecretKey", ByteValue: byteSecretValue},
			}

			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     initialSecrets,
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					{Principal: "admins", Permission: "WRITE"},
					{Principal: "admins", Permission: "READ"},
					{Principal: "admins", Permission: "MANAGE"},
				},
			}

			key := types.NamespacedName{
				Name:      secretsKeyName,
				Namespace: "default",
			}

			toCreate := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating the scope with secrets successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &databricksv1alpha1.SecretScope{}
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating secrets successfully")
			newSecretValue := "newSecretValue"
			updatedSecrets := []databricksv1alpha1.SecretScopeSecret{
				{Key: "newSecretKey", StringValue: newSecretValue},
			}

			updateSpec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     updatedSecrets,
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					{Principal: "admins", Permission: "WRITE"},
					{Principal: "admins", Permission: "READ"},
				},
			}

			fetched.Spec = updateSpec

			Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
			fetchedUpdated := &databricksv1alpha1.SecretScope{}
			Eventually(func() []databricksv1alpha1.SecretScopeSecret {
				_ = k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeSecrets
			}, timeout, interval).Should(Equal(updatedSecrets))

			By("Deleting the scope")
			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
