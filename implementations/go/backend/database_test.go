package main

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with test data
func setupTestDB(t *testing.T) {
	t.Helper()

	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email    TEXT NOT NULL,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (username, email, password)
		VALUES ('admin', 'admin@test.com', 'hashedpassword123')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
}

func TestQueryDB(t *testing.T) {
	setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing test database: %v", err)
		}
	}()

	t.Run("Query all users", func(t *testing.T) {
		results, err := QueryDB("SELECT * FROM users", []interface{}{}, false)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		userList := results.([]map[string]interface{})
		if len(userList) == 0 {
			t.Error("Expected at least one user")
		}
		t.Logf("Found %d users", len(userList))
	})

	t.Run("Query single user", func(t *testing.T) {
		result, err := QueryDB("SELECT * FROM users WHERE username = ?", []interface{}{"admin"}, true)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result == nil {
			t.Error("Expected admin user, got nil")
		}
		user := result.(map[string]interface{})
		if user["username"] != "admin" {
			t.Errorf("Expected username 'admin', got '%v'", user["username"])
		}
		t.Logf("Found user: %v", user)
	})

	t.Run("Query non-existent user", func(t *testing.T) {
		result, err := QueryDB("SELECT * FROM users WHERE username = ?", []interface{}{"nonexistent"}, true)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result != nil {
			t.Error("Expected nil for non-existent user")
		}
	})
}
