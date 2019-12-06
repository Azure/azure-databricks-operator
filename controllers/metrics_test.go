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
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

const metadata = `
		# HELP databricks_request_total Counter of upstream calls to Databricks REST service endpoints
		# TYPE databricks_request_total counter
`

func TestExecutionWithNilError(t *testing.T) {
	databricksRequestCounter.Reset()
	execution := NewExecution("wibble", "wobble")
	execution.Finish(nil)

	expected := `
		databricks_request_total{ action = "wobble", object_type = "wibble", outcome="success" } 1
	`

	if err := testutil.CollectAndCompare(databricksRequestCounter, strings.NewReader(metadata+expected)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestExecutionWithError(t *testing.T) {
	databricksRequestCounter.Reset()
	execution := NewExecution("wibble", "wobble")
	execution.Finish(errors.New("test"))

	expected := `
		databricks_request_total{ action = "wobble", object_type = "wibble", outcome="failure" } 1
	`

	if err := testutil.CollectAndCompare(databricksRequestCounter, strings.NewReader(metadata+expected)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
