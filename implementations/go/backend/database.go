package main

// Download needed dependencies: Standard libraries for SQL, formatting, and OS interaction
import (
	"database/sql"
	"fmt"
	"os"

	// SQLite driver required for the database/sql package to communicate with the file.
	_ "modernc.org/sqlite"
)

// CONFIGURATION: The path to the physical database file used by the application.
const dbPath = "whoknows.db"

// Global db variabel - kan bruges i alle filer
var db *sql.DB

// checkDBExists verifies if the database file is physically present on the disk.
func checkDBExists() bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ConnectDB initiates a connection, checks for file existence, and pings the database.
func connectDB() {
	if !checkDBExists() {
		fmt.Printf("Critical Error: Database file not found at %s\n", dbPath)
		os.Exit(1)
	}

	var err error
	// ⚠️ Vigtigt: brug den globale db variabel, IKKE en lokal!
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("could not open database: %v\n", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("database ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connection Status: Successfully connected to whoknows.db")
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
