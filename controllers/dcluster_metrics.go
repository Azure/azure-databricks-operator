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
	dclusterCreateSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dcluster_create_success_total",
			Help: "Number of create dcluster success",
		},
	)
	dclusterCreateFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dcluster_create_failures_total",
			Help: "Number of create dcluster failures",
		},
	)

	dclusterGetSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dcluster_get_success_total",
			Help: "Number of create dcluster success",
		},
	)
	dclusterGetFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dcluster_get_failures_total",
			Help: "Number of create dcluster failures",
		},
	)

	dclusterCreateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dcluster_creation_duration",
		Help:    "Duration of DB api dcluster create calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	dclusterGetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dcluster_get_duration",
		Help:    "Duration of DB api dcluster get calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})

	dclusterDeleteDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dcluster_delete_duration",
		Help:    "Duration of DB api dcluster delete calls.",
		Buckets: prometheus.LinearBuckets(100, 10, 20),
	})
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(dclusterCreateSuccess, dclusterCreateFailure, dclusterGetSuccess, dclusterGetFailure,
		dclusterCreateDuration, dclusterGetDuration, dclusterDeleteDuration)
}
