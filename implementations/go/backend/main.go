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
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

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

	// 2. Run migrations
	if err := runMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 2. Run migrations
	if err := runMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 3. Server static filer (CSS, billeder)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

	// 4. Page routes
	http.HandleFunc("/", searchHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/reset-password", resetPasswordHandler)

	// 5. API routes
	http.HandleFunc("/api/search", apiSearchHandler)
	http.HandleFunc("/api/login", apiLoginHandler)
	http.HandleFunc("/api/logout", apiLogoutHandler)
	http.HandleFunc("/api/register", apiRegisterHandler)
	http.HandleFunc("/api/reset-password", apiResetPasswordHandler)

	// 6. Start serveren
	fmt.Println("Server starter på port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
