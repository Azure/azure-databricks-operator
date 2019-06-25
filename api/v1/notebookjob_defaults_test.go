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
)

var _ = Describe("Helpers", func() {

	const timeout = time.Second * 5

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
	Context("NotebookJob Default", func() {
		It("should load correct default", func() {
			var nbj NotebookJob
			nbj.LoadDefaultConfig()
			Expect(nbj.Spec.ClusterSpec.NodeTypeId).To(Equal("Standard_DS3_v2"))
			Expect(nbj.Spec.ClusterSpec.SparkVersion).To(Equal("5.2.x-scala2.11"))
			Expect(nbj.Spec.ClusterSpec.NumWorkers).To(Equal(3))
		})
	})
})
