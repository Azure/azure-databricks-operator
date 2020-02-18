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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helpers", func() {

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
