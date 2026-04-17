package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/prometheus/client_golang/prometheus/testutil"

	_ "modernc.org/sqlite"
)

func setupMetricsTestDB(t *testing.T) {
	t.Helper()

	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close db: %v", err)
		}
	})

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			force_password_reset BOOLEAN DEFAULT 0
		);

		CREATE TABLE pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL UNIQUE,
			language TEXT NOT NULL DEFAULT 'en',
			content TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE pages_fts USING fts5(
			title,
			content,
			language UNINDEXED,
			url UNINDEXED,
			content='pages',
			content_rowid='id'
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
}

func setupMetricsTestRuntime(t *testing.T) {
	t.Helper()
	initMetrics()
	resetMetricsForTests()
	store = sessions.NewCookieStore([]byte("test-secret-key-for-metrics"))
	if err := os.Setenv("CSRF_RELAXED", "true"); err != nil {
		t.Fatalf("failed setting CSRF_RELAXED: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("CSRF_RELAXED")
	})
}

func TestMetricsEndpointExposesPrometheusFormat(t *testing.T) {
	setupMetricsTestRuntime(t)
	httpRequestsTotal.WithLabelValues(http.MethodGet, "/metrics", "200").Inc()

	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	t.Cleanup(func() {
		http.DefaultServeMux = oldMux
	})

	registerMetricsRoute()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for /metrics, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "whoknows_http_requests_total") {
		t.Fatalf("expected whoknows_http_requests_total in metrics output")
	}
	if !strings.Contains(body, "process_start_time_seconds") {
		t.Fatalf("expected process_start_time_seconds in metrics output")
	}
}

func TestHTTPMetricsMiddlewareTracksCounterAndHistogramBuckets(t *testing.T) {
	setupMetricsTestRuntime(t)

	handler := metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/123e4567-e89b-12d3-a456-426614174000", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	counter := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(http.MethodGet, "/users/{id}", "201"))
	if counter != 1 {
		t.Fatalf("expected request counter to be 1, got %f", counter)
	}

	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	t.Cleanup(func() {
		http.DefaultServeMux = oldMux
	})
	registerMetricsRoute()

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResp := httptest.NewRecorder()
	mux.ServeHTTP(metricsResp, metricsReq)

	if !strings.Contains(metricsResp.Body.String(), "whoknows_http_request_duration_seconds_bucket") {
		t.Fatalf("expected histogram bucket metric in /metrics output")
	}
}

func TestLoginMetricsSuccessAndFailureIncrementOncePerAttempt(t *testing.T) {
	setupMetricsTestRuntime(t)
	setupMetricsTestDB(t)

	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("failed hashing password: %v", err)
	}
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", "loginuser", "login@example.com", hash)
	if err != nil {
		t.Fatalf("failed inserting login user: %v", err)
	}

	failureForm := url.Values{}
	failureForm.Set("username", "loginuser")
	failureForm.Set("password", "wrong-password")
	failureReq := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(failureForm.Encode()))
	failureReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	failureResp := httptest.NewRecorder()
	apiLoginHandler(failureResp, failureReq)

	successForm := url.Values{}
	successForm.Set("username", "loginuser")
	successForm.Set("password", "password123")
	successReq := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(successForm.Encode()))
	successReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	successResp := httptest.NewRecorder()
	apiLoginHandler(successResp, successReq)

	failureCount := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues(loginOutcomeFailure))
	if failureCount != 1 {
		t.Fatalf("expected login failure count 1, got %f", failureCount)
	}

	successCount := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues(loginOutcomeSuccess))
	if successCount != 1 {
		t.Fatalf("expected login success count 1, got %f", successCount)
	}
}

func TestRegistrationMetricsOutcomeLabels(t *testing.T) {
	t.Run("validation_error", func(t *testing.T) {
		setupMetricsTestRuntime(t)
		setupMetricsTestDB(t)

		validationForm := url.Values{}
		validationForm.Set("username", "newuser")
		validationForm.Set("email", "invalid-email")
		validationForm.Set("password", "password123")
		validationForm.Set("password2", "password123")
		validationReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(validationForm.Encode()))
		validationReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		validationResp := httptest.NewRecorder()
		apiRegisterHandler(validationResp, validationReq)

		validationCount := testutil.ToFloat64(registrationsTotal.WithLabelValues(registrationOutcomeValidationError))
		if validationCount != 1 {
			t.Fatalf("expected validation_error registration count 1, got %f", validationCount)
		}
	})

	t.Run("success", func(t *testing.T) {
		setupMetricsTestRuntime(t)
		setupMetricsTestDB(t)

		successForm := url.Values{}
		successForm.Set("username", "okuser")
		successForm.Set("email", "okuser@example.com")
		successForm.Set("password", "password123")
		successForm.Set("password2", "password123")
		successReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(successForm.Encode()))
		successReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		successResp := httptest.NewRecorder()
		apiRegisterHandler(successResp, successReq)

		successCount := testutil.ToFloat64(registrationsTotal.WithLabelValues(registrationOutcomeSuccess))
		if successCount != 1 {
			t.Fatalf("expected success registration count 1, got %f", successCount)
		}
	})

	t.Run("failure", func(t *testing.T) {
		setupMetricsTestRuntime(t)
		setupMetricsTestDB(t)

		_, err := db.Exec("DROP TABLE users")
		if err != nil {
			t.Fatalf("failed dropping users table for failure test: %v", err)
		}

		failureForm := url.Values{}
		failureForm.Set("username", "failureuser")
		failureForm.Set("email", "failure@example.com")
		failureForm.Set("password", "password123")
		failureForm.Set("password2", "password123")
		failureReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(failureForm.Encode()))
		failureReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		failureResp := httptest.NewRecorder()
		apiRegisterHandler(failureResp, failureReq)

		failureCount := testutil.ToFloat64(registrationsTotal.WithLabelValues(registrationOutcomeFailure))
		if failureCount != 1 {
			t.Fatalf("expected failure registration count 1, got %f", failureCount)
		}
	})
}

func TestSearchMetricsNormalizeLabels(t *testing.T) {
	setupMetricsTestRuntime(t)

	recordSearch("HTML", "EN", "   Fortran   Basics   ", false)
	recordSearch("api", "DA", "fortran", true)

	successCount := testutil.ToFloat64(searchesTotal.WithLabelValues("web", "en", "fortran basics", searchOutcomeSuccess))
	if successCount != 1 {
		t.Fatalf("expected normalized web/en success search count 1, got %f", successCount)
	}

	failureCount := testutil.ToFloat64(searchesTotal.WithLabelValues("api", "da", "fortran", searchOutcomeFailure))
	if failureCount != 1 {
		t.Fatalf("expected api/da failure search count 1, got %f", failureCount)
	}
}
