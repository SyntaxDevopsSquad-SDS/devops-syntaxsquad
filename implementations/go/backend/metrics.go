package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "whoknows_http_requests_total",
			Help: "Total number of HTTP requests received by the Go backend.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "whoknows_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	loginAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "whoknows_login_attempts_total",
			Help: "Number of login attempts partitioned by outcome.",
		},
		[]string{"outcome"},
	)

	registrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "whoknows_registrations_total",
			Help: "Number of registration attempts partitioned by outcome.",
		},
		[]string{"outcome"},
	)

	searchesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "whoknows_searches_total",
			Help: "Number of searches by source, language, query, and outcome.",
		},
		[]string{"source", "language", "query", "outcome"},
	)
)

func initMetrics() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDurationSeconds,
		loginAttemptsTotal,
		registrationsTotal,
		searchesTotal,
	)
}

func registerMetricsRoute() {
	http.Handle("/metrics", promhttp.Handler())
}

func normalizeMetricPath(path string) string {
	if strings.HasPrefix(path, "/static/") {
		return "/static/*"
	}

	switch path {
	case "/", "/about", "/login", "/logout", "/register", "/reset-password":
		return path
	case "/api/search", "/api/login", "/api/logout", "/api/register", "/api/reset-password":
		return path
	case "/metrics":
		return "/metrics"
	default:
		return "/other"
	}
}

func normalizeQueryLabel(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return "_empty"
	}
	normalized = strings.Join(strings.Fields(normalized), " ")
	if len(normalized) > 40 {
		return normalized[:40] + "..."
	}
	return normalized
}

func recordSearch(source, language, query string, resultCount int, hadError bool) {
	if strings.TrimSpace(query) == "" {
		return
	}

	outcome := "empty"
	if hadError {
		outcome = "error"
	} else if resultCount > 0 {
		outcome = "found"
	}

	searchesTotal.WithLabelValues(source, language, normalizeQueryLabel(query), outcome).Inc()
}

type instrumentedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *instrumentedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := normalizeMetricPath(r.URL.Path)

		recorder := &instrumentedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		statusLabel := strconv.Itoa(recorder.statusCode)
		httpRequestsTotal.WithLabelValues(r.Method, path, statusLabel).Inc()
		httpRequestDurationSeconds.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}
