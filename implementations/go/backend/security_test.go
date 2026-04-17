package main

import (
	"testing"
)

// TestPasswordHashing verifies the bcrypt logic for hashing and verification.
// Pure unit test as it doesn't rely on external services.
func TestPasswordHashing(t *testing.T) {
	password := "Secret123!"

	// 1: Generate a hash from a plaintext password
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	// 2: Verify that the correct password matches the hash
	if !verifyPassword(hash, password) {
		t.Error("Verification failed for the correct password")
	}

	// 3: Ensure that an incorrect password is rejected
	if verifyPassword(hash, "WrongPassword") {
		t.Error("System accepted an incorrect password!")
	}
}

// TestMD5Logic checks if  MD5 generator and Regex validator work in sync.
func TestMD5Logic(t *testing.T) {
	input := "hello-world"
	hash := md5Hash(input)

	// 1: Does the generated hash match the expected MD5 format?
	if !isMD5Hash(hash) {
		t.Errorf("md5Hash generated a format that isMD5Hash does not recognize: %s", hash)
	}

	// 2: Does the regex reject invalid strings?
	invalidHashes := []string{
		"not-a-hash",
		"a-f0-9",                          // Too short
		"g1234567890123456789012345678901", // Invalid character 'g'
	}

	for _, h := range invalidHashes {
		if isMD5Hash(h) {
			t.Errorf("isMD5Hash should have rejected: %s", h)
		}
	}
}

// TestCSRFTokenGeneration verifies that tokens are generated with the correct length.
func TestCSRFTokenGeneration(t *testing.T) {
	token, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("Could not generate CSRF token: %v", err)
	}

	// 32 bytes hex-encoded should result in exactly 64 characters
	expectedLength := 64
	if len(token) != expectedLength {
		t.Errorf("Expected length %d, but got %d", expectedLength, len(token))
	}
}

// TestResetTokenFlow tests the end-to-end lifecycle of a password reset token.
// This requires a database connection (side-effect testing).
func TestResetTokenFlow(t *testing.T) {
	if db == nil {
		t.Skip("Database not initialized - skipping database-dependent test")
	}

	// SETUP: Create the temporary table in the :memory: database.
	// Since :memory: starts empty, we must define the schema for the test.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		used_at DATETIME
	);`
	
	_, err := db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create temporary test table: %v", err)
	}

	testUserID := 42

	// Step 1: Generate a new token
	token, err := generateResetToken(testUserID)
	if err != nil {
		t.Fatalf("Failed to generate reset token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Step 2: Validate the generated token
	userID, err := validateResetToken(token)
	if err != nil {
		t.Fatalf("Validation failed for a valid token: %v", err)
	}

	if userID != testUserID {
		t.Errorf("Expected userID %d, but got %d", testUserID, userID)
	}

	// Step 3: Ensure the token cannot be used twice (Security check)
	_, err = validateResetToken(token)
	if err == nil {
		t.Error("Security risk: Token was allowed to be reused!")
	}
}

// Negative path
func TestExpiredResetToken(t *testing.T) {
    // We reuse the database setup
    if db == nil { t.Skip() }

    // 1. Manually insert an EXPIRED token (1 hour ago)
    expiredToken := "i-am-old-and-expired"
    _, err := db.Exec(`
        INSERT INTO password_reset_tokens (user_id, token, expires_at) 
        VALUES (?, ?, datetime('now', '-1 hour'))`, 1, expiredToken)
    
    if err != nil {
        t.Fatalf("Failed to insert expired token: %v", err)
    }

    // 2. Try to validate it
    _, err = validateResetToken(expiredToken)
    
    // 3. We EXPECT an error here
    if err == nil {
        t.Error("Security flaw: System accepted an expired token!")
    } else {
        t.Logf("Correctly rejected expired token with error: %v", err)
    }
}