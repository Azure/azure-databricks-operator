package router

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusHTTPMetric struct {
	Prefix                string
	ClientConnected       prometheus.Gauge
	ResponseTimeHistogram *prometheus.HistogramVec
	Buckets               []float64
}

func initPrometheusHTTPMetric(prefix string, buckets []float64) *prometheusHTTPMetric {
	phm := prometheusHTTPMetric{
		Prefix: prefix,
		ClientConnected: promauto.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_client_connected",
			Help: "Number of active client connections",
		}),
		ResponseTimeHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    prefix + "_response_time",
			Help:    "Histogram of response time for handler",
			Buckets: buckets,
		}, []string{"code", "type", "action", "method"}),
	}

	return &phm
}

func (phm *prometheusHTTPMetric) createHandlerWrapper(typeLabel string, actionLabel string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		wrappedHandler := promhttp.InstrumentHandlerInFlight(phm.ClientConnected,
			promhttp.InstrumentHandlerDuration(phm.ResponseTimeHistogram.MustCurryWith(prometheus.Labels{"type": typeLabel, "action": actionLabel}),
				handler),
		)
		return wrappedHandler
	}
}
