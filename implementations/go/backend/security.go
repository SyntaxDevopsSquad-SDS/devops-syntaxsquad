package main

import (
	"crypto/md5"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

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