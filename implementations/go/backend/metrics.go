package main

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	loginOutcomeSuccess = "success"
	loginOutcomeFailure = "failure"

	registrationOutcomeSuccess         = "success"
	registrationOutcomeValidationError = "validation_error"
	registrationOutcomeFailure         = "failure"

	searchOutcomeSuccess = "success"
	searchOutcomeFailure = "failure"
)

var (
	metricsInitOnce sync.Once

	numericSegmentPattern = regexp.MustCompile(`^\d+$`)
	uuidSegmentPattern    = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	hexSegmentPattern     = regexp.MustCompile(`(?i)^[0-9a-f]{16,}$`)
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
	metricsInitOnce.Do(func() {
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDurationSeconds,
			loginAttemptsTotal,
			registrationsTotal,
			searchesTotal,
		)
	})
}

func registerMetricsRoute() {
	http.Handle("/metrics", promhttp.Handler())
}

func normalizeMetricPath(path string) string {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" || trimmedPath == "/" {
		return "/"
	}

	if strings.HasPrefix(path, "/static/") {
		return "/static/*"
	}

	switch trimmedPath {
	case "/", "/about", "/login", "/logout", "/register", "/reset-password":
		return trimmedPath
	case "/api/search", "/api/login", "/api/logout", "/api/register", "/api/reset-password":
		return trimmedPath
	case "/metrics":
		return "/metrics"
	default:
		segments := strings.Split(strings.Trim(trimmedPath, "/"), "/")
		for i, segment := range segments {
			if numericSegmentPattern.MatchString(segment) || uuidSegmentPattern.MatchString(segment) || hexSegmentPattern.MatchString(segment) {
				segments[i] = "{id}"
			}
		}
		return "/" + strings.Join(segments, "/")
	}
}

func normalizeLanguageLabel(language string) string {
	normalized := strings.ToLower(strings.TrimSpace(language))
	if normalized == "" {
		return "unknown"
	}
	return normalized
}

func normalizeSourceLabel(source string) string {
	normalized := strings.ToLower(strings.TrimSpace(source))
	switch normalized {
	case "html", "web":
		return "web"
	case "api":
		return "api"
	default:
		return "unknown"
	}
}

func normalizeQueryLabel(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return "_empty"
	}
	normalized = strings.Join(strings.Fields(normalized), " ")
	if len(normalized) > 64 {
		return normalized[:64] + "..."
	}
	return normalized
}

func recordLoginAttempt(outcome string) {
	if outcome != loginOutcomeSuccess && outcome != loginOutcomeFailure {
		outcome = loginOutcomeFailure
	}
	loginAttemptsTotal.WithLabelValues(outcome).Inc()
}

func recordRegistrationAttempt(outcome string) {
	switch outcome {
	case registrationOutcomeSuccess, registrationOutcomeValidationError, registrationOutcomeFailure:
		registrationsTotal.WithLabelValues(outcome).Inc()
	default:
		registrationsTotal.WithLabelValues(registrationOutcomeFailure).Inc()
	}
}

func recordSearch(source, language, query string, hadError bool) {
	if strings.TrimSpace(query) == "" {
		return
	}

	outcome := searchOutcomeSuccess
	if hadError {
		outcome = searchOutcomeFailure
	}

	searchesTotal.WithLabelValues(normalizeSourceLabel(source), normalizeLanguageLabel(language), normalizeQueryLabel(query), outcome).Inc()
}

func resetMetricsForTests() {
	httpRequestsTotal.Reset()
	httpRequestDurationSeconds.Reset()
	loginAttemptsTotal.Reset()
	registrationsTotal.Reset()
	searchesTotal.Reset()
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
