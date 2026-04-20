package main

import (
	"database/sql"
	"os"
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

// TestGetDBPath verifies that we can change the DB path via environment variables
func TestGetDBPath(t *testing.T) {
	// Test 1: Default value when environment variable is not set
	os.Unsetenv("DB_PATH") // Ensure it's empty
	expectedDefault := "whoknows.db"
	if path := getDBPath(); path != expectedDefault {
		t.Errorf("Expected default path %s, but got %s", expectedDefault, path)
	}

	// Test 2: Custom value via environment variable
	customPath := "/tmp/test_database.db"
	os.Setenv("DB_PATH", customPath)
	defer os.Unsetenv("DB_PATH") // Clean up after the test

	if path := getDBPath(); path != customPath {
		t.Errorf("Expected custom path %s, but got %s", customPath, path)
	}
}

// TestCheckDBExists verifies the file detection logic
func TestCheckDBExists(t *testing.T) {
	// Test 1: A file that definitely doesn't exist
	os.Setenv("DB_PATH", "non_existent_file_999.db")
	if checkDBExists() {
		t.Error("checkDBExists returned true for a file that does not exist")
	}

	// Test 2: A file that exists (we create a temporary one)
	tempFile := "temp_test.db"
	f, _ := os.Create(tempFile)
	f.Close()
	defer os.Remove(tempFile) // Always clean up files in tests!

	os.Setenv("DB_PATH", tempFile)
	if !checkDBExists() {
		t.Error("checkDBExists returned false for a file that actually exists")
	}
}

// TestQueryDBInvalidSQL verifies that the function returns an error for bad queries
func TestQueryDBInvalidSQL(t *testing.T) {
	setupTestDB(t)
	// We call a table that doesn't exist
	_, err := QueryDB("SELECT * FROM non_existent_table", []interface{}{}, false)
	
	if err == nil {
		t.Error("Expected an error when querying a non-existent table, but got nil")
	} else {
		t.Logf("Correctly caught error: %v", err)
	}
}

// TestDatabaseCRUD handles the Create, Update, and Delete logic
func TestDatabaseCRUD(t *testing.T) {
	setupTestDB(t)

	// --- CREATE ---
	t.Run("Create User", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", 
			"newuser", "new@test.com", "hash123")
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		// Verify it exists
		res, _ := QueryDB("SELECT * FROM users WHERE username = ?", []interface{}{"newuser"}, true)
		if res == nil {
			t.Error("User was not found after insertion")
		}
	})

	// --- UPDATE ---
	t.Run("Update User Email", func(t *testing.T) {
		_, err := db.Exec("UPDATE users SET email = ? WHERE username = ?", "updated@test.com", "admin")
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		res, _ := QueryDB("SELECT * FROM users WHERE username = ?", []interface{}{"admin"}, true)
		user := res.(map[string]interface{})
		if user["email"] != "updated@test.com" {
			t.Errorf("Expected email 'updated@test.com', got '%v'", user["email"])
		}
	})

	// --- DELETE ---
	t.Run("Delete User", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM users WHERE username = ?", "admin")
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		res, _ := QueryDB("SELECT * FROM users WHERE username = ?", []interface{}{"admin"}, true)
		if res != nil {
			t.Error("User still exists after deletion")
		}
	})
}