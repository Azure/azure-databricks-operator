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
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	db "github.com/polar-rams/databricks-sdk-golang"
	dbazure "github.com/polar-rams/databricks-sdk-golang/azure"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var apiClient dbazure.DBClient

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	t := true
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
		}
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = scheme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = databricksv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	host, token := os.Getenv("DATABRICKS_HOST"), os.Getenv("DATABRICKS_TOKEN")
	if host == "" || token == "" {
		Fail("Missing environment variable required for tests. DATABRICKS_HOST and DATABRICKS_TOKEN must both be set.")
	}

	opt := db.NewDBClientOption("", "", host, token, nil, false, 0)
	apiClient := *(dbazure.NewDBClient(opt))

	err = (&SecretScopeReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("SecretScope"),
		Recorder:  k8sManager.GetEventRecorderFor("secretscope-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&DjobReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Djob"),
		Recorder:  k8sManager.GetEventRecorderFor("djob-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&RunReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Run"),
		Recorder:  k8sManager.GetEventRecorderFor("run-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&DclusterReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Run"),
		Recorder:  k8sManager.GetEventRecorderFor("dcluster-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&DbfsBlockReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Run"),
		Recorder:  k8sManager.GetEventRecorderFor("dbfsblock-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&WorkspaceItemReconciler{
		Client:    k8sManager.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Run"),
		Recorder:  k8sManager.GetEventRecorderFor("workspaceitem-controller"),
		APIClient: apiClient,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

const charset = "abcdefghijklmnopqrstuvwxyz"

func randomStringWithCharset(length int, charset string) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
