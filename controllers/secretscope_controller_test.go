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

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	databricks "github.com/xinsnake/databricks-sdk-golang"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
)

var _ = Describe("SecretScope Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	const aclKeyName = "secretscope-with-acls"
	const secretsKeyName = "secretscope-with-secrets"

	BeforeEach(func() {
		// failed test runs that don't clean up leave resources behind.
		keys := []string{aclKeyName, secretsKeyName}
		for _, value := range keys {
			apiClient.Secrets().DeleteSecretScope(value)

			ss := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			k8sClient.Delete(context.Background(), ss)
		}
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		keys := []string{aclKeyName, secretsKeyName}
		for _, value := range keys {
			apiClient.Secrets().DeleteSecretScope(value)

			ss := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			k8sClient.Delete(context.Background(), ss)
		}
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Secret Scope with ACLs", func() {
		It("Should handle scope and ACLs correctly", func() {
			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     make([]databricksv1alpha1.SecretScopeSecret, 0),
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "WRITE"},
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "READ"},
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "MANAGE"},
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
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating ACLs successfully")
			updatedACLs := []databricksv1alpha1.SecretScopeACL{
				databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "READ"},
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
				k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeACLs
			}, timeout, interval).Should(Equal(updatedACLs))

			By("Deleting the scope")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())

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
			byteSecretValue := "aGVsbG8="
			initialSecrets := []databricksv1alpha1.SecretScopeSecret{
				databricksv1alpha1.SecretScopeSecret{Key: "secretKey", StringValue: secretValue},
				databricksv1alpha1.SecretScopeSecret{
					Key: "secretFromSecret",
					ValueFrom: &databricksv1alpha1.SecretScopeValueFrom{
						SecretKeyRef: databricksv1alpha1.SecretScopeKeyRef{
							Name: "k8secret",
							Key:  "username",
						},
					},
				},
				databricksv1alpha1.SecretScopeSecret{Key: "byteSecretKey", ByteValue: byteSecretValue},
			}

			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     initialSecrets,
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "WRITE"},
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "READ"},
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "MANAGE"},
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
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Updating secrets successfully")
			newSecretValue := "newSecretValue"
			updatedSecrets := []databricksv1alpha1.SecretScopeSecret{
				databricksv1alpha1.SecretScopeSecret{Key: "newSecretKey", StringValue: newSecretValue},
			}

			updateSpec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets:     updatedSecrets,
				SecretScopeACLs: []databricksv1alpha1.SecretScopeACL{
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "WRITE"},
					databricksv1alpha1.SecretScopeACL{Principal: "admins", Permission: "READ"},
				},
			}

			fetched.Spec = updateSpec

			Expect(k8sClient.Update(context.Background(), fetched)).Should(Succeed())
			fetchedUpdated := &databricksv1alpha1.SecretScope{}
			Eventually(func() []databricksv1alpha1.SecretScopeSecret {
				k8sClient.Get(context.Background(), key, fetchedUpdated)
				return fetchedUpdated.Spec.SecretScopeSecrets
			}, timeout, interval).Should(Equal(updatedSecrets))

			By("Deleting the scope")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())

			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Secret Scope with ACLs", func() {
		It("Should handle missing k8s secrets", func() {
			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets: []databricksv1alpha1.SecretScopeSecret{
					databricksv1alpha1.SecretScopeSecret{
						Key: "secretFromSecret",
						ValueFrom: &databricksv1alpha1.SecretScopeValueFrom{
							SecretKeyRef: databricksv1alpha1.SecretScopeKeyRef{
								Name: "k8secret",
								Key:  "username",
							},
						},
					},
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

			By("Creating the scope successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			By("Scope has not been marked as IsSubmitted")
			fetched := &databricksv1alpha1.SecretScope{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeFalse())

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

			By("Scope has been marked as IsSubmitted")

			fetched = &databricksv1alpha1.SecretScope{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

			By("Deleting the scope")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())

			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Secret Scope with ACLs", func() {
		It("Should fail if secret scope exist in Databricks", func() {

			var o databricks.DBClientOption
			o.Host = os.Getenv("DATABRICKS_HOST")
			o.Token = os.Getenv("DATABRICKS_TOKEN")

			var APIClient dbazure.DBClient
			APIClient.Init(o)

			Expect(APIClient.Secrets().CreateSecretScope(aclKeyName, "users")).Should(Succeed())

			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets: []databricksv1alpha1.SecretScopeSecret{
					databricksv1alpha1.SecretScopeSecret{
						Key: "secretFromSecret",
						ValueFrom: &databricksv1alpha1.SecretScopeValueFrom{
							SecretKeyRef: databricksv1alpha1.SecretScopeKeyRef{
								Name: "k8secret",
								Key:  "username",
							},
						},
					},
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

			By("Creating the scope successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			By("Scope has been marked as IsSubmitted")
			fetched := &databricksv1alpha1.SecretScope{}
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeFalse())

			By("Deleting the scope")
			Expect(k8sClient.Delete(context.Background(), toCreate)).Should(Succeed())

			Eventually(func() error {
				f := &databricksv1alpha1.SecretScope{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
