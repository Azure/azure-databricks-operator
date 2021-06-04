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
