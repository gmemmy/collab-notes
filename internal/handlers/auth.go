// Package handlers contains HTTP request handlers for the application's endpoints
package handlers

import (
	"log"
	"os"
	"time"

	"collab-notes/internal/db"
	"collab-notes/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SignUp handles user registration by creating a new user account
// and returning a JWT token for authenticated access.
func SignUp(c *fiber.Ctx) error {
	// Parse request body
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Input"})
	}

	// Basic validation
	hashedPw, err := pkg.HashPassword(payload.Password)
	if err != nil {
		log.Println("Error hashing password", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	// Insert user
	userID := uuid.New().String()
	result, err := db.DB.Exec(
		"INSERT INTO users (id, email, password) VALUES (?, ?, ?)",
		userID, payload.Email, hashedPw,
	)
	if err != nil {
		log.Println("Error inserting user:", err)
	}

	// Get user ID
	id, err := result.LastInsertId()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Generate JWT token
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user-id": id,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token": signedToken,
	})
}
