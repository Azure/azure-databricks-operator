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
	"fmt"
	"math/rand"
	"testing"
	"time"

	randStr "microsoft/azure-databricks-operator/pkg/rand"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const timeout = time.Second * 5

func TestStorageNotebookJob(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	key := types.NamespacedName{
		Name:      randStr.String(10),
		Namespace: "default",
	}
	created := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		}}

	// Test Create
	fetched := &NotebookJob{}
	g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
}

func TestAddFinalizers(t *testing.T) {
	finalizersCount := rand.Intn(100) + 1
	finalizers := []string{}
	for i := 0; i < finalizersCount; i++ {
		finalizers = append(finalizers, fmt.Sprintf("finalizer%d.domain.k8s.org", i))
	}

	notebookJob := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		}}

	g := gomega.NewGomegaWithT(t)

	g.Expect(len(notebookJob.GetFinalizers())).To(gomega.BeIdenticalTo(0))
	for _, fn := range finalizers {
		g.Expect(containsString(notebookJob.GetFinalizers(), fn)).To(gomega.BeFalse())
		notebookJob.AddFinalizer(fn)
		g.Expect(containsString(notebookJob.GetFinalizers(), fn)).To(gomega.BeTrue())
	}
	g.Expect(len(notebookJob.GetFinalizers())).To(gomega.BeIdenticalTo(finalizersCount))
}

func TestRemoveFinalizers(t *testing.T) {
	finalizersCount := rand.Intn(100) + 1
	finalizers := []string{}
	for i := 0; i < finalizersCount; i++ {
		finalizers = append(finalizers, fmt.Sprintf("finalizer%d.domain.k8s.org", i))
	}

	notebookJob := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:       randStr.String(10),
			Namespace:  "default",
			Finalizers: finalizers,
		}}

	g := gomega.NewGomegaWithT(t)

	g.Expect(len(notebookJob.GetFinalizers())).To(gomega.BeIdenticalTo(finalizersCount))
	for _, fn := range finalizers {
		g.Expect(containsString(notebookJob.GetFinalizers(), fn)).To(gomega.BeTrue())
		notebookJob.RemoveFinalizer(fn)
		g.Expect(containsString(notebookJob.GetFinalizers(), fn)).To(gomega.BeFalse())
	}
	g.Expect(len(notebookJob.GetFinalizers())).To(gomega.BeIdenticalTo(0))
}

func TestHasFinalizers(t *testing.T) {
	finalizer := "finalizer.domain.k8s.org"
	notebookJob := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      randStr.String(10),
			Namespace: "default",
		}}

	g := gomega.NewGomegaWithT(t)

	g.Expect(notebookJob.HasFinalizer(finalizer)).To(gomega.BeFalse())
	notebookJob.ObjectMeta.Finalizers = []string{finalizer}
	g.Expect(notebookJob.HasFinalizer(finalizer)).To(gomega.BeTrue())
}

func TestContainsString(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	slice1 := []string{"strA", "strB", "strC"}
	slice2 := []string{"strD", "strE", "strF"}

	for _, str := range slice1 {
		g.Expect(containsString(slice1, str)).To(gomega.BeTrue())
	}

	for _, str := range slice2 {
		g.Expect(containsString(slice1, str)).To(gomega.BeFalse())
	}
}

func TestRemoveString(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	slice := []string{"strA", "strB", "strC"}
	before := len(slice)

	for _, str := range slice {
		g.Expect(containsString(removeString(slice, str), str)).To(gomega.BeFalse())
	}

	for _, str := range slice {
		g.Expect(len(slice)).To(gomega.BeIdenticalTo(before))
		slice = removeString(slice, str)
		g.Expect(len(slice)).To(gomega.BeIdenticalTo(before - 1))
		before--
	}

	g.Expect(len(slice)).To(gomega.BeIdenticalTo(0))

}

func TestIsBeingDeleted(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	finalizer := "finalizer.domain.k8s.org"
	finalizers := []string{finalizer}

	key := types.NamespacedName{
		Name:      randStr.String(10),
		Namespace: "default",
	}
	notebookjob := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:       key.Name,
			Namespace:  key.Namespace,
			Finalizers: finalizers,
		}}

	defer func() {
		c.Get(context.TODO(), key, notebookjob)
		notebookjob.RemoveFinalizer(finalizer)
		c.Update(context.TODO(), notebookjob)
	}()

	c.Create(context.TODO(), notebookjob)
	c.Delete(context.TODO(), notebookjob)

	g.Eventually(func() bool {
		_ = c.Get(context.TODO(), key, notebookjob)
		return notebookjob.IsBeingDeleted()
	}, timeout,
	).Should(gomega.BeTrue())
}

func TestIsRunning(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	notebookJob := &NotebookJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      randStr.String(10),
			Namespace: "default",
		}}
	g.Expect(notebookJob.IsRunning()).To(gomega.BeFalse())

	notebookJob.Spec.NotebookTask.RunID = rand.Intn(100) + 1

	g.Expect(notebookJob.IsRunning()).To(gomega.BeTrue())
}
