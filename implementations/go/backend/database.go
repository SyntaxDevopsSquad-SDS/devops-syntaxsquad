package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// dbPath reads the database path from the DB_PATH environment variable,
// falling back to "whoknows.db" if not set.
func getDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "whoknows.db"
}

// Global db variabel - kan bruges i alle filer
var db *sql.DB

func checkDBExists() bool {
	if _, err := os.Stat(getDBPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

// connectDB initiates a connection, checks for file existence, and pings the database.
func connectDB() {
	if !checkDBExists() {
		fmt.Printf("Critical Error: Database file not found at %s\n", getDBPath())
		os.Exit(1)
	}

	var err error
	db, err = sql.Open("sqlite", getDBPath())
	if err != nil {
		fmt.Printf("could not open database: %v\n", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("database ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connection Status: Successfully connected to", getDBPath())
}

// Fjern db *sql.DB parameter - brug den globale db
func getUserID(username string) (int, error) {
	// Prepare the SQL statement to prevent SQL INJECTION vulnerabilities.
	statement, err := db.Prepare("SELECT id FROM users WHERE username = ?")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer statement.Close()

	var id int
	err = statement.QueryRow(username).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to query user ID: %w", err)
	}
	return id, nil
}
