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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	successMetric = "success"
	failureMetric = "failure"
)

var databricksRequestHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "databricks_request_duration_seconds",
	Help: "Duration of upstream calls to Databricks REST service endpoints",
}, []string{"object_type", "action", "outcome"})

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(databricksRequestHistogram)
}

// NewExecution creates an Execution instance and starts the timer
func NewExecution(objectType string, action string) Execution {
	return Execution{
		begin:  time.Now(),
		labels: prometheus.Labels{"object_type": objectType, "action": action},
	}
}

// Execution tracks state for an API execution for emitting metrics
type Execution struct {
	begin  time.Time
	labels prometheus.Labels
}

// Finish is used to log duration and success/failure
func (e *Execution) Finish(err error) {
	if err == nil {
		e.labels["outcome"] = successMetric
	} else {
		e.labels["outcome"] = failureMetric
	}
	duration := time.Since(e.begin)
	databricksRequestHistogram.With(e.labels).Observe(duration.Seconds())
}
