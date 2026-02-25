package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    // 1. Forbind til databasen
		connectDB()
    // 2. Server static filer (CSS, billeder)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

    // 3. Page routes
    http.HandleFunc("/", searchHandler)
    http.HandleFunc("/about", aboutHandler)
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/register", registerHandler)

    // 4. API routes
    http.HandleFunc("/api/login", apiLoginHandler)
		http.HandleFunc("/api/logout", logoutHandler)
    http.HandleFunc("/api/register", apiRegisterHandler)

    // 4. Start serveren
    fmt.Println("Server starter på port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}