// Package main is the entry point for the Collab Notes application.
// It initializes the server, database connection, and sets up routes.
package main

import (
	"log"
	"os"

	"collab-notes/internal/db"
	"collab-notes/internal/handlers/auth"
	"collab-notes/internal/handlers/notes"
	"collab-notes/internal/middleware"
	"collab-notes/internal/realtime"

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

	authHandler := auth.NewHandler(db.DB, &auth.JWTService{})
	notesHandler := notes.NewHandler(db.DB)

	app.Post("/signup", authHandler.SignUp)
	app.Post("/login", authHandler.Login)

	note := app.Group("/notes", middleware.Protected())
	note.Get("/", notesHandler.GetNotes)
	note.Post("/", notesHandler.CreateNote)
	note.Put("/:id", notesHandler.UpdateNote)
	note.Delete("/:id", notesHandler.DeleteNote)

	// WebSocket routes with authentication
	ws := app.Group("/ws", middleware.Protected())
	ws.Get("/notes/:id", realtime.HandleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
