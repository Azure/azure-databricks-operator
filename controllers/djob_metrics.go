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
	djobCreateSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "djob_success_total",
			Help: "Number of create djob success",
		},
	)
	djobCreateFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "djob_failures_total",
			Help: "Number of create djob failures",
		},
	)

	djobGetSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "djob_success_total",
			Help: "Number of create djob success",
		},
	)
	djobGetFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "djob_failures_total",
			Help: "Number of create djob failures",
		},
	)

	djobCreateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "djob_creation_duration",
		Help:    "Duration of DB api create calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20), 
	})

	djobGetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "djob_get_duration",
		Help:    "Duration of DB api get calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),  // 5 buckets, each 5 centigrade wide.
	})

	djobDeleteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "djob_delete_duration",
		Help:    "Duration of DB api delete calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),  // 5 buckets, each 5 centigrade wide.
	})
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(djobCreateSuccess, djobCreateFailure, djobGetDuration)
}
