package prom

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CustomPromMetrics struct {
	HttpCounter   prometheus.Counter
	HttpHistogram *prometheus.HistogramVec
	register      prometheus.Registerer
}

func NewCustomPromMetrics() *CustomPromMetrics {
	m := &CustomPromMetrics{
		HttpCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "total_http_req",
			Help: "Total http requests",
		}),
		HttpHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "response_time_seconds_http",
			Help:    "Http reqursts hystogram",
			Buckets: []float64{0.1 * 0.001, 1 * 0.001, 10 * 0.001, 100 * 0.001, 500 * 0.001, 1000 * 0.001},
		}, []string{"method", "path", "status"}),
	}

	m.register = prometheus.WrapRegistererWithPrefix("ypmetricssrv_http_", prometheus.DefaultRegisterer)

	m.register.MustRegister(
		m.HttpCounter,
		m.HttpHistogram,
	)

	return m
}

func (m *CustomPromMetrics) IncHttpRequest() {
	m.HttpCounter.Inc()
}

func (m *CustomPromMetrics) IncHttpHistogram(method, path string, status int, duration float64) {
	m.HttpHistogram.With(prometheus.Labels{"method": method, "path": path, "status": strconv.Itoa(status)}).Observe(duration)
}

func (m *CustomPromMetrics) Register() prometheus.Registerer {
	return m.register
}
