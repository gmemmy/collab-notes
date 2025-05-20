// Package notes provides handlers for managing user notes
// including creating and retrieving notes
package notes

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DBInterface defines the methods for database operations
type DBInterface interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

// Note represents a user's note with metadata
type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Handler handles HTTP requests related to notes operations
type Handler struct {
	db DBInterface
}

// NewHandler creates a new Handler with the provided database interface
func NewHandler(db DBInterface) *Handler {
	return &Handler{db: db}
}

// GetNotes retrieves all notes for a user
func (h *Handler) GetNotes(c *fiber.Ctx) error {
	userID := c.Locals("user-id").(string)

	rows, err := h.db.Query("SELECT id, user_id, title, content, created_at, updated_at FROM notes WHERE user_id = ?", userID)
	if err != nil {
		log.Println("Error fetching notes:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("Error closing rows:", err)
		}
	}()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			log.Println("Error scanning note:", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		notes = append(notes, n)
	}

	return c.JSON(notes)
}

// CreateNote creates a new note for the user
func (h *Handler) CreateNote(c *fiber.Ctx) error {
	userID := c.Locals("user-id").(string)

	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Input validation
	payload.Title = strings.TrimSpace(payload.Title)
	payload.Content = strings.TrimSpace(payload.Content)

	if payload.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title cannot be empty"})
	}

	id := uuid.New().String()
	_, err := h.db.Exec("INSERT INTO notes (id, user_id, title, content) VALUES (?, ?, ?, ?)",
		id, userID, payload.Title, payload.Content)
	if err != nil {
		log.Println("Error creating note:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}
