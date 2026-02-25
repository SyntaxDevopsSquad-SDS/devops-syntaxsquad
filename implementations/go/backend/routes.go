package main

import (
    "html/template"
    "net/http"
)

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
    tmpl.ExecuteTemplate(w, "layout", BaseData{})
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
    tmpl.ExecuteTemplate(w, "layout", BaseData{})
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
    tmpl.ExecuteTemplate(w, "layout", BaseData{})
}