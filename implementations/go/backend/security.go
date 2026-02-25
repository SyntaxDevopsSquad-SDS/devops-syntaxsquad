package main

import (
	"crypto/md5"
	"fmt"
)

// hashPassword takes a plaintext password and returns its MD5 hash as a hexadecimal string.
func hashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}
