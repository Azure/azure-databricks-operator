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

package v1alpha1

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
	Context("String Operations", func() {
		It("should contain string", func() {
			slice1 := []string{"strA", "strB", "strC"}
			slice2 := []string{"strD", "strE", "strF"}

			for _, str := range slice1 {
				Expect(containsString(slice1, str)).To(BeTrue())
			}

			for _, str := range slice2 {
				Expect(containsString(slice1, str)).To(BeFalse())
			}
		})

		It("should remove string", func() {
			slice := []string{"strA", "strB", "strC"}
			before := len(slice)

			for _, str := range slice {
				Expect(containsString(removeString(slice, str), str)).To(BeFalse())
			}

			for _, str := range slice {
				Expect(len(slice)).To(BeIdenticalTo(before))
				slice = removeString(slice, str)
				Expect(len(slice)).To(BeIdenticalTo(before - 1))
				before--
			}

			Expect(len(slice)).To(BeIdenticalTo(0))
		})

		It("should create random string matches length", func() {
			a1 := RandomString(5)
			a2 := RandomString(5)
			b1 := RandomString(10)

			Expect(a1).ToNot(Equal(a2))
			Expect(len(a1)).To(Equal(len(a2)))
			Expect(len(b1)).To(Equal(10))
		})
	})
})
