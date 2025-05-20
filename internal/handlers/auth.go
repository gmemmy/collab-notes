// Package handlers contains HTTP request handlers for the application's endpoints
package handlers

import (
	"database/sql"
	"log"
	"os"
	"time"

	"collab-notes/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DBInterface defines the methods for database operations
type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// AuthHandler is a struct that contains the database and JWT interfaces
type AuthHandler struct {
	db  DBInterface
	jwt JWTInterface
}

// JWTInterface defines the methods for JWT operations
type JWTInterface interface {
	NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token
	SignedString(token *jwt.Token, key []byte) (string, error)
}

// JWTService is a struct that contains the JWT interface
type JWTService struct{}

// NewWithClaims creates a new JWT token with given claims
func (j *JWTService) NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token {
	return jwt.NewWithClaims(method, claims)
}

// SignedString signs a JWT token with a given key
func (j *JWTService) SignedString(token *jwt.Token, key []byte) (string, error) {
	return token.SignedString(key)
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db DBInterface, jwt JWTInterface) *AuthHandler {
	return &AuthHandler{
		db:  db,
		jwt: jwt,
	}
}

// SignUp handles user registration by creating a new user account
// and returning a JWT token for authenticated access.
func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Input"})
	}

	hashedPw, err := pkg.HashPassword(payload.Password)
	if err != nil {
		log.Println("Error hashing password", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	userID := uuid.New().String()
	result, err := h.db.Exec(
		"INSERT INTO users (id, email, password) VALUES (?, ?, ?)",
		userID, payload.Email, hashedPw,
	)
	if err != nil {
		log.Println("Error inserting user:", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user-id": id,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := h.jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := h.jwt.SignedString(token, []byte(secret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token": signedToken,
	})
}
