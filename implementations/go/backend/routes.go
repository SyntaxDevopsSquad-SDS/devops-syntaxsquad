package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
)

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
	User  string
	Flash string
	Error string
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
	session.Save(r, w)
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
	session.Save(r, w)
	return flash
}

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
		rows, err := db.Query(
			"SELECT title, content, language, url FROM pages WHERE language = ? AND content LIKE ?",
			language, "%"+query+"%",
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var page Page
			if err := rows.Scan(&page.Title, &page.Content, &page.Language, &page.URL); err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}
	}

	tmpl, err := parseTemplates("layout.html", "search.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", SearchPageData{
		BaseData:      BaseData{User: getSessionUser(r), Flash: getFlash(w, r)},
		SearchResults: searchResults,
		Query:         query,
	})
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := parseTemplates("layout.html", "about.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", BaseData{User: getSessionUser(r)})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if getSessionUser(r) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tmpl, err := parseTemplates("layout.html", "login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", BaseData{User: getSessionUser(r), Flash: getFlash(w, r)})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if getSessionUser(r) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tmpl, err := parseTemplates("layout.html", "register.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", BaseData{User: getSessionUser(r)})
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
		rows, err := db.Query(
			"SELECT title, content, language, url FROM pages WHERE language = ? AND content LIKE ?",
			language, "%"+query+"%",
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var page Page
			if err := rows.Scan(&page.Title, &page.Content, &page.Language, &page.URL); err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			searchResults = append(searchResults, page)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"search_results": searchResults,
	})
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedHash string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedHash)
	if err != nil {
		tmpl, _ := parseTemplates("layout.html", "login.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"})
		return
	}

	if !verifyPassword(storedHash, password) {
		// Fallback: check if stored hash is legacy MD5
		if isMD5Hash(storedHash) && md5Hash(password) == storedHash {
			// Upgrade to bcrypt
			newHash, err := hashPassword(password)
			if err == nil {
				db.Exec("UPDATE users SET password = ? WHERE username = ?", newHash, username)
			}
		} else {
			tmpl, _ := parseTemplates("layout.html", "login.html")
			tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"})
			return
		}
	}

	session, _ := store.Get(r, "session")
	session.Values["user"] = username
	session.Save(r, w)

	setFlash(w, r, "You were logged in")
	http.Redirect(w, r, "/", http.StatusFound)
}

func apiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if username == "" {
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a username"})
		return
	}

	if email == "" || !strings.Contains(email, "@") {
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "You have to enter a valid email address"})
		return
	}

	if len(password) < 8 {
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Password must be at least 8 characters"})
		return
	}

	if password != password2 {
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The two passwords do not match"})
		return
	}

	var exists int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if exists > 0 {
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "The username is already taken"})
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		fmt.Println("Hash error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, hashedPassword)
	if err != nil {
		fmt.Println("Register error:", err)
		tmpl, _ := parseTemplates("layout.html", "register.html")
		tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Could not create user"})
		return
	}

	setFlash(w, r, "You were successfully registered and can login now")
	http.Redirect(w, r, "/login", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user")
	session.Save(r, w)
	setFlash(w, r, "You were logged out")
	http.Redirect(w, r, "/", http.StatusFound)
}