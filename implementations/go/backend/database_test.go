package main

import (
	"testing"
)

func TestQueryDB(t *testing.T) {
	// Connect to database
	connectDB()

	// Test 1: Query all users
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

	// Test 2: Query single user
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

	// Test 3: Query with no results
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