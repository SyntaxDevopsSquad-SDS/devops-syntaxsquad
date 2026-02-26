package main

import (
    "fmt"
    "html/template"
    "net/http"

    "github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

// BaseData indeholder data som alle sider bruger
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

// Helper funktion - henter den loggede bruger fra session
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

    tmpl, err := template.ParseFiles(
        "../templates/layout.html",
        "../templates/search.html",
    )
    if err != nil {
        http.Error(w, "Template error", http.StatusInternalServerError)
        return
    }

    data := SearchPageData{
	BaseData:      BaseData{User: getUserFromContext(r)},
	SearchResults: searchResults,
	Query:         query,
    }

    tmpl.ExecuteTemplate(w, "layout", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles(
        "../templates/layout.html",
        "../templates/about.html",
    )
    if err != nil {
        http.Error(w, "Template error", http.StatusInternalServerError)
        return
    }
    tmpl.ExecuteTemplate(w, "layout", BaseData{User: getUserFromContext(r)})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles(
        "../templates/layout.html",
        "../templates/login.html",
    )
    if err != nil {
        http.Error(w, "Template error", http.StatusInternalServerError)
        return
    }
    tmpl.ExecuteTemplate(w, "layout", BaseData{User: getUserFromContext(r)})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles(
        "../templates/layout.html",
        "../templates/register.html",
    )
    if err != nil {
        http.Error(w, "Template error", http.StatusInternalServerError)
        return
    }
    tmpl.ExecuteTemplate(w, "layout", BaseData{User: getUserFromContext(r)})
}

/*
################################################################################
# API Endpoints
################################################################################*/

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    var storedHash string
    err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedHash)
    if err != nil {
        fmt.Println("Login fejl:", err)
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/login.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: err.Error()})
        return
    }

    if !verifyPassword(storedHash, password) {
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/login.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Invalid username or password"})
        return
    }

    session, _ := store.Get(r, "session")
    session.Values["user"] = username
    session.Save(r, w)

    http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    delete(session.Values, "user")
    session.Save(r, w)
    http.Redirect(w, r, "/", http.StatusFound)
}

func apiRegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/register", http.StatusFound)
        return
    }

    username := r.FormValue("username")
    email := r.FormValue("email")
    password := r.FormValue("password")
    password2 := r.FormValue("password2")

    if username == "" || email == "" || password == "" {
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/register.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "All fields are required"})
        return
    }

    if password != password2 {
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/register.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Passwords do not match"})
        return
    }

    var exists int
    db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
    if exists > 0 {
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/register.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: "Username already taken"})
        return
    }

    hashedPassword := hashPassword(password)
    _, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
        username, email, hashedPassword)
    if err != nil {
        fmt.Println("Register fejl:", err)
        tmpl, _ := template.ParseFiles("../templates/layout.html", "../templates/register.html")
        tmpl.ExecuteTemplate(w, "layout", BaseData{Error: err.Error()})
        return
    }

    http.Redirect(w, r, "/login", http.StatusFound)
}