package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
)

// setupRoutesTest - Defined only once to avoid "redeclared" error
func setupRoutesTest() {
	store = sessions.NewCookieStore([]byte("test-secret-key"))
}

/* ################################################################################
# PART 1: UTILITY TESTS
################################################################################
*/

func TestGetSecretKey(t *testing.T) {
	// 1. Vi bruger '_ =' for at sige: "Jeg ved den returnerer en fejl, men jeg ignorerer den med vilje"
	_ = os.Setenv("SECRET_KEY", "my-super-secret")

	// 2. 'defer' skal pakkes ind i en lille anonym funktion for at vi kan ignorere fejlen indeni den
	defer func() {
		_ = os.Unsetenv("SECRET_KEY")
	}()

	key := getSecretKey()
	if string(key) != "my-super-secret" {
		t.Errorf("Expected secret key 'my-super-secret', got '%s'", string(key))
	}
}

func TestSessionAndFlash(t *testing.T) {
	setupRoutesTest()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	testMessage := "Success!"
	setFlash(w, r, testMessage)
	r.AddCookie(w.Result().Cookies()[0])
	flash := getFlash(w, r)
	if flash != testMessage {
		t.Errorf("Expected flash '%s', got '%s'", testMessage, flash)
	}
}

func TestCSRFLogic(t *testing.T) {
	setupRoutesTest()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	token := generateAndStoreCSRFToken(w, r)
	rPost := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	rPost.AddCookie(w.Result().Cookies()[0])
	rPost.Form = make(map[string][]string)
	rPost.Form.Set("csrf_token", token)
	if !validateCSRFToken(httptest.NewRecorder(), rPost) {
		t.Error("CSRF validation failed")
	}
}

/* ################################################################################
# PART 2: API HANDLER TESTS
################################################################################
*/

func TestAPILoginHandler(t *testing.T) {
	setupRoutesTest()
	setupTestDB(t)
	// Do NOT defer db.Close() here to keep DB alive for other tests

	// Ensure the temporary table matches what the handler expects
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN force_password_reset INTEGER DEFAULT 0")

	// Pre-create a user for the tests
	password := "testpass123"
	hash, _ := hashPassword(password)
	_, _ = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		"testuser", "test@test.com", hash)

	t.Run("Successful Login", func(t *testing.T) {
		w := httptest.NewRecorder()
		rGet := httptest.NewRequest(http.MethodGet, "/login", nil)
		token := generateAndStoreCSRFToken(w, rGet)
		cookie := w.Result().Cookies()[0]

		data := url.Values{}
		data.Set("username", "testuser")
		data.Set("password", password)
		data.Set("csrf_token", token)

		rPost := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		rPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rPost.AddCookie(cookie)

		apiLoginHandler(w, rPost)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302 (Redirect), got %d", w.Code)
		}
	})

	t.Run("Failed Login - Wrong Password", func(t *testing.T) {
		w := httptest.NewRecorder()
		rGet := httptest.NewRequest(http.MethodGet, "/login", nil)
		token := generateAndStoreCSRFToken(w, rGet)

		data := url.Values{}
		data.Set("username", "testuser")
		data.Set("password", "wrong-password")
		data.Set("csrf_token", token)

		rPost := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		rPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rPost.AddCookie(w.Result().Cookies()[0])

		apiLoginHandler(w, rPost)

		// On failure, it should return 200 OK (re-rendering the login page) instead of a redirect
		if w.Code == http.StatusFound {
			t.Error("Expected login to fail for wrong password, but got a redirect")
		}
	})
}

/* ################################################################################
# PART 3: HTML PAGE HANDLERS
################################################################################
*/

func TestHTMLHandlers(t *testing.T) {
	setupRoutesTest()
	setupTestDB(t)

	t.Run("About Page Status 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/about", nil)

		aboutHandler(w, r)

		// Vi tjekker om siden indlæses korrekt.
		// Hvis denne fejler med 500, er det ofte fordi den ikke kan finde 'layout.html'
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for About page, got %d", w.Code)
		}
	})

	t.Run("Register Page contains CSRF", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/register", nil)

		registerHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for Register page, got %d", w.Code)
		}

		// Vi tjekker om der rent faktisk er genereret et CSRF token til formularen
		body := w.Body.String()
		if !strings.Contains(body, "name=\"csrf_token\"") {
			t.Errorf("Expected Register page to contain CSRF hidden input name=\"csrf_token\", body was: %s", body)
		}
	})

	t.Run("Search Page with empty query", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/search?q=", nil)

		searchHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for Search page, got %d", w.Code)
		}
	})
}
