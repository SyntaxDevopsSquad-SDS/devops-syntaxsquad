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

// checkDBExists verifies if the database file is physically present on the disk.
func checkDBExists() bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ConnectDB initiates a connection, checks for file existence, and pings the database.
func ConnectDB() (*sql.DB, error) {
	if !checkDBExists() {
		fmt.Printf("Critical Error: Database file not found at %s\n", dbPath)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// Ping confirms that the connection is active and the file is readable.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("Connection Status: Successfully connected to whoknows.db")
	return db, nil
}

func main() {
	db, err := ConnectDB()
	if err != nil {
		fmt.Printf("Initialization failed: %v\n", err)
		return
	}
	defer db.Close()
}