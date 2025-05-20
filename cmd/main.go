package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	app := fiber.New()

	// Auth routes
	// app.Post("/signup", handlers.SignUp)
	// app.Post("/login", handlers.Login)

	// // Notes routes (protected)
	// note := app.Group("/notes", middleware.Protected())
	// note.Post("/", handlers.CreateNote)
	// note.Get("/", handlers.GetNotes)
	// note.Put("/:id", handlers.UpdateNote)
	// note.Delete("/:id", handlers.DeleteNote)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}