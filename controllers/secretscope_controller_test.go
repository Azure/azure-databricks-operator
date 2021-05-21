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

package controllers

import (
	"context"
	"fmt"
	"time"

	"os"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	databricks "github.com/polar-rams/databricks-sdk-golang"
	dbazure "github.com/polar-rams/databricks-sdk-golang/azure"
	dbhttpmodels "github.com/polar-rams/databricks-sdk-golang/azure/secrets/httpmodels"
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
			req := dbhttpmodels.DeleteSecretScopeReq{
				Scope: value,
			}
			_ = apiClient.Secrets().DeleteSecretScope(req)

			ss := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			_ = k8sClient.Delete(context.Background(), ss)
		}
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		keys := []string{aclKeyName, secretsKeyName}
		for _, value := range keys {
			req := dbhttpmodels.DeleteSecretScopeReq{
				Scope: value,
			}
			_ = apiClient.Secrets().DeleteSecretScope(req)

			ss := &databricksv1alpha1.SecretScope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			_ = k8sClient.Delete(context.Background(), ss)
		}
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

			_ = k8sClient.Get(context.Background(), key, fetched)

			fmt.Println(fetched.IsSubmitted())

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

	Context("Secret Scope with ACLs", func() {
		It("Should handle missing k8s secrets", func() {
			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets: []databricksv1alpha1.SecretScopeSecret{
					{
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
				_ = k8sClient.Get(context.Background(), key, fetched)
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
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeTrue())

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

	Context("Secret Scope with ACLs", func() {
		It("Should fail if secret scope exist in Databricks", func() {
			opt := databricks.NewDBClientOption("", "", os.Getenv("DATABRICKS_HOST"), os.Getenv("DATABRICKS_TOKEN"), nil, false, 0)
			APIClient := dbazure.NewDBClient(opt)

			createReq := dbhttpmodels.CreateSecretScopeReq{
				Scope:                  aclKeyName,
				InitialManagePrincipal: "users",
			}
			Expect(APIClient.Secrets().CreateSecretScope(createReq)).Should(Succeed())
			defer func() {
				deleteReq := dbhttpmodels.DeleteSecretScopeReq{
					Scope: aclKeyName,
				}
				Expect(APIClient.Secrets().DeleteSecretScope(deleteReq)).Should(Succeed())
				time.Sleep(time.Second * 5)
			}()

			spec := databricksv1alpha1.SecretScopeSpec{
				InitialManagePrincipal: "users",
				SecretScopeSecrets: []databricksv1alpha1.SecretScopeSecret{
					{
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
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.IsSubmitted()
			}, timeout, interval).Should(BeFalse())

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
