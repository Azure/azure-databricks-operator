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
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	runCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricPrefix + "run_total",
			Help: "Counter related to the Run CRD partitioned by status and method invoked. Status = success/fail and method indicates REST endpoint",
		},
		[]string{"status", "method"},
	)

	runSubmitDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: metricPrefix + "run_submit_request_duration_seconds",
		Help: "Duration of DB api run submit calls.",
	})

	runNowDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: metricPrefix + "run_now_request_duration_seconds",
		Help: "Duration of DB api run now calls.",
	})

	runGetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: metricPrefix + "run_get_request_duration_seconds",
		Help: "Duration of DB api run get calls.",
	})

	runGetOutputDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: metricPrefix + "run_get_output_request_duration_seconds",
		Help: "Duration of DB api run get output calls.",
	})

	runDeleteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: metricPrefix + "run_delete_request_duration_seconds",
		Help: "Duration of DB api run delete calls.",
	})
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(runCounterVec, runSubmitDuration,
		runNowDuration, runGetDuration, runGetOutputDuration, runDeleteDuration)
}
