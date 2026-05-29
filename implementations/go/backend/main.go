package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func main() {
	// Structured JSON logging to stdout
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load .env file if it exists (ignored in production where env vars are set via systemd)
	if err := godotenv.Load(); err != nil {
		slog.Warn("could not load .env file", "error", err)
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

	// Initialize and expose Prometheus metrics
	initMetrics()

	// 2. Run migrations
	if err := runMigrations(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	// Start polling DB for registered users and active sessions every 30s
	startDBMetricsPoller()

	// 3. Server static filer (CSS, billeder)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

	// 4. Page routes
	http.HandleFunc("/", searchHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/weather", weatherHandler)
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
	http.HandleFunc("/health", healthHandler)
	registerMetricsRoute()

	slog.Info("server starting", "port", 8080)
	if err := http.ListenAndServe(":8080", metricsMiddleware(http.DefaultServeMux)); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
