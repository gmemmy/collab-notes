// Package pkg provides utility functions that are used
// across different parts of the application
package pkg

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword securely hashes a plaintext password using bcrypt
// with the default cost factor for security
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
