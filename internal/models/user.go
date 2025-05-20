// Package models contains data structures that represent
// the application's domain objects and database schema
package models

// User represents a user account in the system with
// identification, authentication and profile information
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
