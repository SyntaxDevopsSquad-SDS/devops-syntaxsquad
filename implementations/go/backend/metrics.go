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

	// Business metrics
	searchZeroResultsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "whoknows_search_zero_results_total",
			Help: "Number of searches that returned zero results, partitioned by language. Indicates content gaps.",
		},
		[]string{"language"},
	)

	activeSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "whoknows_active_sessions",
			Help: "Number of users currently logged in (server-side session tracking via PostgreSQL).",
		},
	)

	sessionDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "whoknows_session_duration_seconds",
			Help:    "Duration of user sessions in seconds, from login to explicit logout.",
			Buckets: []float64{30, 60, 120, 300, 600, 1200, 1800, 3600, 7200},
		},
	)

	registeredUsersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "whoknows_registered_users_total",
			Help: "Total number of registered users in the database. Polled every 30 seconds.",
		},
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
			searchZeroResultsTotal,
			activeSessions,
			sessionDurationSeconds,
			registeredUsersTotal,
		)
	})
}

func startDBMetricsPoller() {
	go func() {
		for {
			var count float64
			if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err == nil {
				registeredUsersTotal.Set(count)
			}

			var activeSessCount float64
			if err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE logout_at IS NULL").Scan(&activeSessCount); err == nil {
				activeSessions.Set(activeSessCount)
			}

			time.Sleep(30 * time.Second)
		}
	}()
}

func recordZeroResults(language string) {
	searchZeroResultsTotal.WithLabelValues(normalizeLanguageLabel(language)).Inc()
}

func recordSessionStart(sessionID, username string) {
	if _, err := db.Exec(
		"INSERT INTO sessions (id, username, login_at) VALUES ($1, $2, NOW()) ON CONFLICT (id) DO NOTHING",
		sessionID, username,
	); err != nil {
		return
	}
}

func recordSessionEnd(sessionID string) {
	var loginAt time.Time
	err := db.QueryRow("SELECT login_at FROM sessions WHERE id = $1 AND logout_at IS NULL", sessionID).Scan(&loginAt)
	if err != nil {
		return
	}
	duration := time.Since(loginAt).Seconds()
	if _, err := db.Exec("UPDATE sessions SET logout_at = NOW() WHERE id = $1", sessionID); err != nil {
		return
	}
	sessionDurationSeconds.Observe(duration)
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
	searchZeroResultsTotal.Reset()
	activeSessions.Set(0)
	registeredUsersTotal.Set(0)
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
