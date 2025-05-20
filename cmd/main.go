// Package main is the entry point for the Collab Notes application.
// It initializes the server, database connection, and sets up routes.
package main

import (
	"log"
	"os"

	"collab-notes/internal/db"
	"collab-notes/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	// Connect to DB
	db.Connect()

	app := fiber.New()

	// Auth routes
	app.Post("/signup", handlers.SignUp)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
