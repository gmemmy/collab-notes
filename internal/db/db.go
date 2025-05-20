// Package db provides database connection and operations
// for the application.
package db

import (
	"database/sql"
	"log"
	"os"

	// Import MySQL driver for database connection.
	// This blank import is needed to register the MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// DB is the global database connection instance used throughout the application
var DB *sql.DB

// Connect establishes a connection to the MySQL database using environment
// variables and initializes the global DB instance
func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	DB = db
	log.Println("Connected to MySQL database 🎉")
}
