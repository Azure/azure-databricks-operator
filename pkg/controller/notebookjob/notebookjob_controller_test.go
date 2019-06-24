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

package notebookjob

import (
	"fmt"
	"testing"
	"time"

	microsoftv1beta1 "microsoft/azure-databricks-operator/pkg/apis/microsoft/v1beta1"
	randStr "microsoft/azure-databricks-operator/pkg/rand"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var namespacedName = types.NamespacedName{Name: randStr.String(10), Namespace: "default"}
var expectedRequest = reconcile.Request{NamespacedName: namespacedName}

const timeout = time.Second * 60

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secrets := make(map[string][]byte)
	secrets["secret1"] = []byte("value1")
	secret1 := &v1.Secret{Data: secrets, ObjectMeta: metav1.ObjectMeta{
		Name: "test-secret", Namespace: "default",
	}}

	instance := &microsoftv1beta1.NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: microsoftv1beta1.NotebookJobSpec{
			NotebookTask: microsoftv1beta1.NotebookTask{
				NotebookPath: "/test-notebook",
			},
			NotebookSpec: map[string]string{
				"TestSpec": fmt.Sprintf("%v", time.Now().String()),
			},
			NotebookSpecSecrets: []microsoftv1beta1.NotebookSpecSecret{
				{
					SecretName: "test-secret",
					Mapping: []microsoftv1beta1.KeyMapping{
						{
							SecretKey: "secret1",
							OutputKey: "SECRET_VALUE",
						},
					},
				},
			},
			NotebookAdditionalLibraries: []microsoftv1beta1.NotebookAdditionalLibrary{
				{
					Type: "maven",
					Properties: map[string]string{
						"coordinates": "com.microsoft.azure:azure-eventhubs-spark_2.11:2.3.9",
					},
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the secrets needed
	c.Create(context.TODO(), secret1)
	defer func() {
		c.Delete(context.TODO(), secret1)
	}()

	time.Sleep(1 * time.Second)

	// Create the NotebookJob object and expect the Reconcile to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() bool {
		_ = c.Get(context.TODO(), namespacedName, instance)
		return instance.HasFinalizer(finalizerName)
	}, timeout,
	).Should(gomega.BeTrue())

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() bool {
		_ = c.Get(context.TODO(), namespacedName, instance)
		return instance.IsSubmitted()
	}, timeout,
	).Should(gomega.BeTrue())

	time.Sleep(10 * time.Second)

	instance2 := &microsoftv1beta1.NotebookJob{}
	err = c.Get(context.TODO(), namespacedName, instance2)
	if err != nil {
		t.Logf("failed to get object to be deleted: %v", err)
	}
	err = c.Delete(context.TODO(), instance2)
	if err != nil {
		t.Logf("failed to delete object: %v", err)
		return
	}

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error { return c.Get(context.TODO(), namespacedName, instance2) }, timeout).
		Should(gomega.MatchError("NotebookJob.microsoft.k8s.io \"" + namespacedName.Name + "\" not found"))
}
