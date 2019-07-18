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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("SecretScope Controller", func() {

	const timeout = time.Second * 60
	const interval = time.Second * 1

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
	Context("Secret Scope with ACLs", func() {
		It("Should handle scope and ACLs correctly", func() {
			spec := databricksv1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     make([]databricksv1.SecretScopeSecret, 0),
				SecretScopeACLs: []databricksv1.SecretScopeACL{
					databricksv1.SecretScopeACL{Principal: "joshua.agudo@team.telstra.com", Permission: "WRITE"},
					databricksv1.SecretScopeACL{Principal: "joshua.agudo@team.telstra.com", Permission: "READ"},
					databricksv1.SecretScopeACL{Principal: "joshua.agudo@team.telstra.com", Permission: "MANAGE"},
				},
			}

			key := types.NamespacedName{
				Name:      "secretscope-with-acls",
				Namespace: "default",
			}

			toCreate := &databricksv1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating the scope with ACLs successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &databricksv1.SecretScope{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating ACLs successfully")
			updatedACLs := []databricksv1.SecretScopeACL{
				databricksv1.SecretScopeACL{Principal: "joshua.agudo@team.telstra.com", Permission: "READ"},
			}

			updateSpec := databricksv1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     make([]databricksv1.SecretScopeSecret, 0),
				SecretScopeACLs:        updatedACLs,
			}

			fetched.Spec = updateSpec

			Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
			fetchedUpdated := &databricksv1.SecretScope{}
			Eventually(func() []databricksv1.SecretScopeACL {
				k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeACLs
			}, timeout, interval).Should(Equal(updatedACLs))

			By("Deleting the scope")
			Eventually(func() error {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() bool {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return f.IsBeingDeleted()
			}, timeout, interval).Should(BeTrue())

			time.Sleep(time.Second * 5)
			By("Removing the finaliser")
			Eventually(func() bool {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return f.HasFinalizer(databricksv1.SecretScopeFinalizerName)
			}, timeout, interval).Should(BeFalse())
		})
	})

	Context("Secret Scope with secrets", func() {
		It("Should handle scope and secrets correctly", func() {

			// setup k8s secret
			k8SecretKey := types.NamespacedName{
				Name:      "k8secret",
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
			byteSecretValue := []byte("hello")
			initialSecrets := []databricksv1.SecretScopeSecret{
				databricksv1.SecretScopeSecret{Key: "secretKey", StringValue: &secretValue},
				databricksv1.SecretScopeSecret{
					Key: "secretFromSecret",
					ValueFrom: &databricksv1.SecretScopeValueFrom{
						SecretKeyRef: databricksv1.SecretScopeKeyRef{
							Name: "k8secret",
							Key:  "username",
						},
					},
				},
				databricksv1.SecretScopeSecret{Key: "byteSecretKey", ByteValue: &byteSecretValue},
			}

			spec := databricksv1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     initialSecrets,
				SecretScopeACLs:        make([]databricksv1.SecretScopeACL, 0),
			}

			key := types.NamespacedName{
				Name:      "secretscope-with-secrets",
				Namespace: "default",
			}

			toCreate := &databricksv1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating the scope with secrets successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			fetched := &databricksv1.SecretScope{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating secrets successfully")
			newSecretValue := "newSecretValue"
			updatedSecrets := []databricksv1.SecretScopeSecret{
				databricksv1.SecretScopeSecret{Key: "newSecretKey", StringValue: &newSecretValue},
			}

			updateSpec := databricksv1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     updatedSecrets,
				SecretScopeACLs:        make([]databricksv1.SecretScopeACL, 0),
			}

			fetched.Spec = updateSpec

			Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
			fetchedUpdated := &databricksv1.SecretScope{}
			Eventually(func() []databricksv1.SecretScopeSecret {
				k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeSecrets
			}, timeout, interval).Should(Equal(updatedSecrets))

			By("Deleting the scope")
			Eventually(func() error {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() bool {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return f.IsBeingDeleted()
			}, timeout, interval).Should(BeTrue())

			time.Sleep(time.Second * 5)
			By("Removing the finaliser")
			Eventually(func() bool {
				f := &databricksv1.SecretScope{}
				k8sClient.Get(context.Background(), key, f)
				return f.HasFinalizer(databricksv1.SecretScopeFinalizerName)
			}, timeout, interval).Should(BeFalse())
		})
	})
})
