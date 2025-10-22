package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// --- Pomoćna struktura za hvatanje status koda ---

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{w, http.StatusOK}
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

// --- Definicija HTTP metrika ---

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "requests_served_total",
			Help:      "Total number of HTTP requests processed, labelled by method, route, and status code.",
		},
		[]string{"method", "route", "code"},
	)

	httpRequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "request_latency_seconds",
			Help:      "Duration of HTTP requests in seconds, labelled by method and route.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestLatency)
}

// --- Glavni Metrics Middleware ---

// HTTPMetricsMiddleware beleži latenciju, metodu, rutu i status kod za svaki HTTP zahtev.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := NewStatusRecorder(w)

		// Rute iz Handlera su /configurations, /configgroups, itd.
		// Koristimo r.URL.Path za rutu (može se poboljšati korišćenjem rutera kao što je Chi da bi dobili šablon rute)
		route := r.URL.Path

		next.ServeHTTP(recorder, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(recorder.Status)
		method := r.Method

		// Beleženje Countera: Uključuje code, route, i method labele
		httpRequestsTotal.With(prometheus.Labels{
			"method": method,
			"route":  route,
			"code":   statusCode,
		}).Inc()

		// Beleženje Histogram Latencije: Uključuje route i method labele
		httpRequestLatency.With(prometheus.Labels{
			"method": method,
			"route":  route,
		}).Observe(duration)
	})
}

// MetricsHandler izlaže Prometheus metrike na /metrics ruti
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
