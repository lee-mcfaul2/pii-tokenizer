package obs

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "tokenizer_requests_total"},
		[]string{"endpoint", "status"},
	)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tokenizer_request_duration_seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "tokenizer_errors_total"},
		[]string{"type"},
	)
	RedisErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "tokenizer_redis_errors_total"},
		[]string{"op"},
	)
	KMasterErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "tokenizer_kmaster_errors_total"},
		[]string{"backend", "operation"},
	)
	KMasterVersionInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "tokenizer_kmaster_version_in_use"},
		[]string{"version"},
	)
	KMasterVersionsLoaded = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "tokenizer_kmaster_versions_loaded"},
	)
)

func init() {
	prometheus.MustRegister(
		RequestsTotal, RequestDuration, ErrorsTotal,
		RedisErrors, KMasterErrors,
		KMasterVersionInUse, KMasterVersionsLoaded,
	)
}

func Handler() http.Handler { return promhttp.Handler() }
