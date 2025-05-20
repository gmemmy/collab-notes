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
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	db.Connect()

	app := fiber.New()

	authHandler := handlers.NewAuthHandler(db.DB, &handlers.JWTService{})

	app.Post("/signup", authHandler.SignUp)
	app.Post("/login", authHandler.Login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
