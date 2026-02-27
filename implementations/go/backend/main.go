package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (ignored in production where env vars are set via systemd)
	godotenv.Load()

	// Initialize session store (must happen after godotenv.Load)
	store = sessions.NewCookieStore(getSecretKey())
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

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
	http.HandleFunc("/api/search", apiSearchHandler)
	http.HandleFunc("/api/login", apiLoginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/register", apiRegisterHandler)

	// 5. Start serveren
	fmt.Println("Server starter på port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}