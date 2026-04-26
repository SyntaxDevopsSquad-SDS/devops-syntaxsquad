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

func getSecretKey() []byte {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		fmt.Println("Critical Error: SECRET_KEY environment variable is not set")
		os.Exit(1)
	}
	return []byte(key)
}

var store *sessions.CookieStore

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

func setFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := store.Get(r, "session")
	session.Values["flash"] = message
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

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

func isCSRFRelaxed() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("CSRF_RELAXED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

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
		searchQuery := strings.TrimSpace(query) + ":*"
		rows, err := db.Query(
			"SELECT title, content, language, url FROM pages WHERE language = $1 AND search_vector @@ to_tsquery('english', $2)",
			language, searchQuery,
		)
		if err != nil {
			recordSearch("web", language, query, true)
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
				recordSearch("web", language, query, true)
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}

		if err := rows.Err(); err != nil {
			recordSearch("web", language, query, true)
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}

		recordSearch("web", language, query, false)
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
		searchQuery := strings.TrimSpace(query) + ":*"
		rows, err := db.Query(
			"SELECT title, content, language, url FROM pages WHERE language = $1 AND search_vector @@ to_tsquery('english', $2)",
			language, searchQuery,
		)
		if err != nil {
			recordSearch("api", language, query, true)
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
				recordSearch("api", language, query, true)
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}

		if err := rows.Err(); err != nil {
			recordSearch("api", language, query, true)
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}

		recordSearch("api", language, query, false)
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

	outcome := loginOutcomeFailure
	defer func() {
		recordLoginAttempt(outcome)
	}()

	if !requireCSRF(w, r, "/api/login") {
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedHash string
	var userID int
	var forceReset bool
	err := db.QueryRow(
		"SELECT id, password, COALESCE(force_password_reset, false) FROM users WHERE username = $1",
		username,
	).Scan(&userID, &storedHash, &forceReset)
	if err != nil {
		tmpl, _ := parseTemplates("layout.html", "login.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if !verifyPassword(storedHash, password) {
		if isMD5Hash(storedHash) && md5Hash(password) == storedHash {
			newHash, err := hashPassword(password)
			if err == nil {
				if _, err := db.Exec("UPDATE users SET password = $1 WHERE username = $2", newHash, username); err != nil {
					log.Printf("error updating password: %v", err)
				}
			}
		} else {
			tmpl, _ := parseTemplates("layout.html", "login.html")
			if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"}); err != nil {
				log.Printf("error executing template: %v", err)
			}
			return
		}
	}

	if forceReset {
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
	outcome = loginOutcomeSuccess
	http.Redirect(w, r, "/", http.StatusFound)
}

func apiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	outcome := registrationOutcomeFailure
	defer func() {
		recordRegistrationAttempt(outcome)
	}()

	if !requireCSRF(w, r, "/api/register") {
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if username == "" {
		outcome = registrationOutcomeValidationError
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a username"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if email == "" || !strings.Contains(email, "@") {
		outcome = registrationOutcomeValidationError
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a valid email address"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if len(password) < 8 {
		outcome = registrationOutcomeValidationError
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Password must be at least 8 characters"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	if password != password2 {
		outcome = registrationOutcomeValidationError
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The two passwords do not match"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&exists); err != nil {
		log.Printf("error checking username existence: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		outcome = registrationOutcomeValidationError
		tmpl, _ := parseTemplates("layout.html", "register.html")
		if err := tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The username is already taken"}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		fmt.Println("Hash error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		username, email, hashedPassword,
	)
	if err != nil {
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
	outcome = registrationOutcomeSuccess
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

	_, err = db.Exec(
		"UPDATE users SET password = $1, force_password_reset = false WHERE id = $2",
		hashedPassword, userID,
	)
	if err != nil {
		log.Printf("error updating password: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	setFlash(w, r, "Password reset successful. Please log in.")
	http.Redirect(w, r, "/login", http.StatusFound)
}

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
