package main

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	if db != nil {
		_ = db.Close()
	}

	if err := os.Setenv("DATABASE_URL", "postgres://whoknows:whoknows@localhost:5432/whoknows_test?sslmode=disable"); err != nil {
		t.Fatalf("Failed to set DATABASE_URL: %v", err)
	}
	connectDB()

	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
			db = nil
		}
	})

	_, err := db.Exec(`
		DROP TABLE IF EXISTS users CASCADE;
		CREATE TABLE IF NOT EXISTS users (
			id       SERIAL PRIMARY KEY,
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
		result, err := QueryDB("SELECT * FROM users WHERE username = $1", []interface{}{"admin"}, true)
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
		result, err := QueryDB("SELECT * FROM users WHERE username = $1", []interface{}{"nonexistent"}, true)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result != nil {
			t.Error("Expected nil for non-existent user")
		}
	})
}

func TestGetDatabaseURL(t *testing.T) {
	if err := os.Unsetenv("DATABASE_URL"); err != nil {
		t.Fatalf("Failed to unset DATABASE_URL: %v", err)
	}
	expected := "postgres://whoknows:whoknows@localhost:5432/whoknows?sslmode=disable"
	if url := getDatabaseURL(); url != expected {
		t.Errorf("Expected default URL %s, got %s", expected, url)
	}

	customURL := "postgres://user:pass@myhost:5432/mydb?sslmode=disable"
	if err := os.Setenv("DATABASE_URL", customURL); err != nil {
		t.Fatalf("Failed to set DATABASE_URL: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("DATABASE_URL")
	}()

	if url := getDatabaseURL(); url != customURL {
		t.Errorf("Expected custom URL %s, got %s", customURL, url)
	}
}

func TestQueryDBInvalidSQL(t *testing.T) {
	setupTestDB(t)
	_, err := QueryDB("SELECT * FROM non_existent_table", []interface{}{}, false)
	if err == nil {
		t.Error("Expected an error when querying a non-existent table, but got nil")
	} else {
		t.Logf("Correctly caught error: %v", err)
	}
}

func TestDatabaseCRUD(t *testing.T) {
	setupTestDB(t)

	t.Run("Create User", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
			"newuser", "new@test.com", "hash123")
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		res, _ := QueryDB("SELECT * FROM users WHERE username = $1", []interface{}{"newuser"}, true)
		if res == nil {
			t.Error("User was not found after insertion")
		}
	})

	t.Run("Update User Email", func(t *testing.T) {
		_, err := db.Exec("UPDATE users SET email = $1 WHERE username = $2", "updated@test.com", "admin")
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		res, _ := QueryDB("SELECT * FROM users WHERE username = $1", []interface{}{"admin"}, true)
		user := res.(map[string]interface{})
		if user["email"] != "updated@test.com" {
			t.Errorf("Expected email 'updated@test.com', got '%v'", user["email"])
		}
	})

	t.Run("Delete User", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM users WHERE username = $1", "admin")
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		res, _ := QueryDB("SELECT * FROM users WHERE username = $1", []interface{}{"admin"}, true)
		if res != nil {
			t.Error("User still exists after deletion")
		}
	})
}
