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
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricPrefix  = "databricks_"
	successMetric = "success"
	failureMetric = "failure"
)

func trackExecutionTime(histogram prometheus.Histogram, f func() error) error {
	timer := prometheus.NewTimer(histogram)
	defer timer.ObserveDuration()
	return f()
}

func trackSuccessFailure(err error, counterVec *prometheus.CounterVec, method string) {
	if err == nil {
		counterVec.With(prometheus.Labels{"status": successMetric, "method": method}).Inc()
	} else {
		counterVec.With(prometheus.Labels{"status": failureMetric, "method": method}).Inc()
	}
}
