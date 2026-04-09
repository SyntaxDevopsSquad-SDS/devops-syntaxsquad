package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// generateCSRFToken returns a cryptographically secure random 32-byte hex token.
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashPassword takes a plaintext password and returns a bcrypt hash.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// verifyPassword compares a plaintext password with a stored bcrypt hash.
func verifyPassword(storedHash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
}

// isMD5Hash returns true if the string looks like an MD5 hash (32 hex chars).
func isMD5Hash(s string) bool {
	matched, _ := regexp.MatchString(`^[a-f0-9]{32}$`, s)
	return matched
}

// md5Hash returns the MD5 hex digest of a string.
func md5Hash(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

// generateResetToken creates a one-time password reset token and stores it in DB
func generateResetToken(userID int) (string, error) {
	token, err := generateCSRFToken() // Reuse secure random generation
	if err != nil {
		return "", err
	}

	query := `INSERT INTO password_reset_tokens (user_id, token, expires_at) 
	          VALUES (?, ?, datetime('now', '+15 minutes'))`
	_, err = db.Exec(query, userID, token)
	if err != nil {
		return "", err
	}
	return token, nil
}

// validateResetToken checks if token is valid and marks it as used
func validateResetToken(token string) (int, error) {
	query := `SELECT user_id FROM password_reset_tokens 
	          WHERE token = ? AND expires_at > datetime('now') AND used_at IS NULL`
	var userID int
	err := db.QueryRow(query, token).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token")
	}

	// Mark token as used
	_, err = db.Exec(`UPDATE password_reset_tokens SET used_at = datetime('now') WHERE token = ?`, token)
	return userID, err
}
