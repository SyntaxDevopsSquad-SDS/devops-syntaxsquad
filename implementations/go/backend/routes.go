package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

/*
################################################################################
# Session Management and Utility Functions
################################################################################
*/

// getSecretKey reads the secret key from SECRET_KEY env variable. Exits if not set.
func getSecretKey() []byte {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		fmt.Println("Critical Error: SECRET_KEY environment variable is not set")
		os.Exit(1)
	}
	return []byte(key)
}

var store *sessions.CookieStore

// BaseData contains data shared across all pages
type BaseData struct {
	User      string
	Flash     string
	Error     string
	CSRFToken string
}

type SearchPageData struct {
	BaseData
	SearchResults []Page
	Query         string
}

type Page struct {
	Title    string
	Content  string
	Language string
	URL      string
}

// getSessionUser retrieves the logged-in user from the session cookie.
func getSessionUser(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}
	user, ok := session.Values["user"].(string)
	if !ok {
		return ""
	}
	return user
}

// setFlash stores a one-time flash message in the session.
func setFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := store.Get(r, "session")
	session.Values["flash"] = message
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// getFlash reads and clears the flash message from the session.
func getFlash(w http.ResponseWriter, r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}
	flash, ok := session.Values["flash"].(string)
	if !ok || flash == "" {
		return ""
	}
	delete(session.Values, "flash")
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return ""
	}
	return flash
}

// generateAndStoreCSRFToken creates a new one-time CSRF token and stores it in the session.
func generateAndStoreCSRFToken(w http.ResponseWriter, r *http.Request) string {
	token, err := generateCSRFToken()
	if err != nil {
		return ""
	}
	session, _ := store.Get(r, "session")
	session.Values["csrf_token"] = token
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return ""
	}
	return token
}

// validateCSRFToken checks the submitted token against the session-stored token (one-time use).
func validateCSRFToken(w http.ResponseWriter, r *http.Request) bool {
	session, err := store.Get(r, "session")
	if err != nil {
		return false
	}
	storedToken, ok := session.Values["csrf_token"].(string)
	if !ok || storedToken == "" {
		return false
	}
	delete(session.Values, "csrf_token")
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return false
	}
	submittedToken := r.FormValue("csrf_token")
	return submittedToken != "" && submittedToken == storedToken
}

// isCSRFRelaxed returns true if CSRF checks are explicitly relaxed via env var.
// Intended for controlled simulation/test environments only.
func isCSRFRelaxed() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("CSRF_RELAXED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// requireCSRF validates CSRF token, or allows a controlled bypass when CSRF_RELAXED is enabled.
func requireCSRF(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	if validateCSRFToken(w, r) {
		return true
	}

	if isCSRFRelaxed() {
		log.Printf("warning: CSRF check bypassed for %s from %s (CSRF_RELAXED enabled)", endpoint, r.RemoteAddr)
		return true
	}

	http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
	return false
}

/*
################################################################################
# HTML Page Handlers
################################################################################
*/

func parseTemplates(files ...string) (*template.Template, error) {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = "../templates/" + f
	}
	return template.ParseFiles(paths...)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	language := r.URL.Query().Get("language")

	if language == "" {
		language = "en"
	}

	var searchResults []Page

	if query != "" {
		// Add wildcard to allow partial matching (e.g., "f" matches "Fortran")
		searchQuery := strings.TrimSpace(query) + "*"
		rows, err := db.Query(
			"SELECT p.title, p.content, p.language, p.url FROM pages_fts f JOIN pages p ON f.rowid = p.id WHERE f.language = ? AND pages_fts MATCH ?",
			language, searchQuery,
		)
		if err != nil {
			recordSearch("html", language, query, 0, true)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := rows.Close(); err != nil {
				log.Printf("error closing rows: %v", err)
			}
		}()

		for rows.Next() {
			var page Page
			if err := rows.Scan(&page.Title, &page.Content, &page.Language, &page.URL); err != nil {
				recordSearch("html", language, query, 0, true)
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}

		recordSearch("html", language, query, len(searchResults), false)
	}

	tmpl, err := parseTemplates("layout.html", "search.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", SearchPageData{
		BaseData:      BaseData{User: getSessionUser(r), Flash: getFlash(w, r)},
		SearchResults: searchResults,
		Query:         query,
	}); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := parseTemplates("layout.html", "about.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", BaseData{User: getSessionUser(r)}); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if getSessionUser(r) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	token := generateAndStoreCSRFToken(w, r)
	tmpl, err := parseTemplates("layout.html", "login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", BaseData{
		User:      getSessionUser(r),
		Flash:     getFlash(w, r),
		CSRFToken: token,
	}); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// logoutHandler handles /logout (HTML route) and delegates to apiLogoutHandler
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	apiLogoutHandler(w, r)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if getSessionUser(r) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	token := generateAndStoreCSRFToken(w, r)
	tmpl, err := parseTemplates("layout.html", "register.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", BaseData{
		User:      getSessionUser(r),
		CSRFToken: token,
	}); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

/*
################################################################################
# API Endpoints
################################################################################
*/

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	language := r.URL.Query().Get("language")

	if language == "" {
		language = "en"
	}

	var searchResults []Page

	if query != "" {
		// Add wildcard to allow partial matching (e.g., "f" matches "Fortran")
		searchQuery := strings.TrimSpace(query) + "*"
		rows, err := db.Query(
			"SELECT p.title, p.content, p.language, p.url FROM pages_fts f JOIN pages p ON f.rowid = p.id WHERE f.language = ? AND pages_fts MATCH ?",
			language, searchQuery,
		)
		if err != nil {
			recordSearch("api", language, query, 0, true)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := rows.Close(); err != nil {
				log.Printf("error closing rows: %v", err)
			}
		}()

		for rows.Next() {
			var page Page
			if err := rows.Scan(&page.Title, &page.Content, &page.Language, &page.URL); err != nil {
				recordSearch("api", language, query, 0, true)
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}

		recordSearch("api", language, query, len(searchResults), false)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"search_results": searchResults,
	}); err != nil {
		log.Printf("error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if !requireCSRF(w, r, "/api/login") {
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedHash string
	var userID int
	var forceReset bool
	err := db.QueryRow("SELECT id, password, COALESCE(force_password_reset, 0) FROM users WHERE username = ?", username).Scan(&userID, &storedHash, &forceReset)
	if err != nil {
		loginAttemptsTotal.WithLabelValues("failure").Inc()
		tmpl, _ := parseTemplates("layout.html", "login.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if !verifyPassword(storedHash, password) {
		// Fallback: check if stored hash is legacy MD5
		if isMD5Hash(storedHash) && md5Hash(password) == storedHash {
			// Upgrade to bcrypt
			newHash, err := hashPassword(password)
			if err == nil {
				if _, err := db.Exec("UPDATE users SET password = ? WHERE username = ?", newHash, username); err != nil {
					log.Printf("error updating password: %v", err)
				}
			}
		} else {
			loginAttemptsTotal.WithLabelValues("failure").Inc()
			tmpl, _ := parseTemplates("layout.html", "login.html")
			if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"}); err != nil {
				log.Printf("error executing template: %v", err)
			}
			return
		}
	}

	// If forced password reset, generate token and redirect
	if forceReset {
		loginAttemptsTotal.WithLabelValues("reset_required").Inc()
		token, err := generateResetToken(userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		setFlash(w, r, "Your account requires a password reset")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusFound)
		return
	}

	session, _ := store.Get(r, "session")
	session.Values["user"] = username
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	setFlash(w, r, "You were logged in")
	loginAttemptsTotal.WithLabelValues("success").Inc()
	http.Redirect(w, r, "/", http.StatusFound)
}

func apiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	if !requireCSRF(w, r, "/api/register") {
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if username == "" {
		registrationsTotal.WithLabelValues("validation_error").Inc()
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a username"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if email == "" || !strings.Contains(email, "@") {
		registrationsTotal.WithLabelValues("validation_error").Inc()
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a valid email address"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if len(password) < 8 {
		registrationsTotal.WithLabelValues("validation_error").Inc()
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Password must be at least 8 characters"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if password != password2 {
		registrationsTotal.WithLabelValues("validation_error").Inc()
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The two passwords do not match"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists); err != nil {
		registrationsTotal.WithLabelValues("error").Inc()
		log.Printf("error checking username existence: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		registrationsTotal.WithLabelValues("already_exists").Inc()
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The username is already taken"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		registrationsTotal.WithLabelValues("error").Inc()
		fmt.Println("Hash error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, hashedPassword)
	if err != nil {
		registrationsTotal.WithLabelValues("error").Inc()
		fmt.Println("Register error:", err)
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Could not create user"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}
	session, _ := store.Get(r, "session")
	session.Values["user"] = username
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	setFlash(w, r, "You were successfully registered and logged in")
	registrationsTotal.WithLabelValues("success").Inc()
	http.Redirect(w, r, "/", http.StatusFound)
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing reset token", http.StatusBadRequest)
		return
	}

	tmpl, err := parseTemplates("layout.html", "reset-password.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", BaseData{
		CSRFToken: token,
		Flash:     getFlash(w, r),
	}); err != nil {
		log.Printf("error executing template: %v", err)
	}
}

func apiResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if token == "" {
		http.Error(w, "Missing reset token", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		setFlash(w, r, "Password must be at least 8 characters")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusFound)
		return
	}

	if password != password2 {
		setFlash(w, r, "Passwords do not match")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusFound)
		return
	}

	userID, err := validateResetToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET password = ?, force_password_reset = 0 WHERE id = ?", hashedPassword, userID)
	if err != nil {
		log.Printf("error updating password: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	setFlash(w, r, "Password reset successful. Please log in.")
	http.Redirect(w, r, "/login", http.StatusFound)
}

// apiLogoutHandler handles /api/logout (API endpoint)
func apiLogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user")
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	setFlash(w, r, "You were logged out")
	http.Redirect(w, r, "/", http.StatusFound)
}
