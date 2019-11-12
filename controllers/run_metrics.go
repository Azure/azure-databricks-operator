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
	runNowSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_now_success_total",
			Help: "Number of run now success",
		},
	)
	runNowFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_now_failures_total",
			Help: "Number of run now failures",
		},
	)

	runSubmitSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_submit_success_total",
			Help: "Number of run submit success",
		},
	)
	runSubmitFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_submit_failures_total",
			Help: "Number of run submit failures",
		},
	)

	runGetSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_get_success_total",
			Help: "Number of get run success",
		},
	)
	runGetFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_get_failures_total",
			Help: "Number of get run failures",
		},
	)

	runGetOutputSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_getoutput_success_total",
			Help: "Number of get run success",
		},
	)

	runGetOutputFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "run_getoutput_failures_total",
			Help: "Number of get run failures",
		},
	)

	runSubmitDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "run_submit_duration",
		Help:    "Duration of DB api run submit calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	runNowDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "run_now_duration",
		Help:    "Duration of DB api run now calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	runGetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "run_get_duration",
		Help:    "Duration of DB api run get calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	runGetOutputDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "run_get_output_duration",
		Help:    "Duration of DB api run get output calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	runDeleteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "run_delete_duration",
		Help:    "Duration of DB api run delete calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(runNowSuccess, runNowFailure, runSubmitSuccess, runSubmitFailure,
		runGetSuccess, runGetFailure, runGetOutputSuccess, runGetOutputFailure, runSubmitDuration,
		runNowDuration, runGetDuration, runGetOutputDuration, runDeleteDuration)
}
