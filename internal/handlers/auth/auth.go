// Package auth contains the handlers for the authentication endpoints
// TODO: Implement idempotent signup by verifying existing password and returning token if valid
package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"

	"quanta/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DBInterface defines the methods for database operations
type DBInterface interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Handler is a struct that contains the database and JWT interfaces
type Handler struct {
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

// NewHandler creates a new Handler
func NewHandler(db DBInterface, jwt JWTInterface) *Handler {
	return &Handler{
		db:  db,
		jwt: jwt,
	}
}

// SignUp handles user registration by creating a new user account
// and returning a JWT token for authenticated access.
func (h *Handler) SignUp(c *fiber.Ctx) error {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Input"})
	}

	payload.Email = strings.TrimSpace(payload.Email)
	payload.Password = strings.TrimSpace(payload.Password)

	// Validate email format
	_, err := mail.ParseAddress(payload.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	if len(payload.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 8 characters long"})
	}

	// Check for duplicate email
	var existingUserID string
	err = h.db.QueryRow("SELECT id FROM users WHERE email = ?", payload.Email).Scan(&existingUserID)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already in use"})
	} else if !errors.Is(err, sql.ErrNoRows) {
		// Some other DB error
		log.Println("Error checking for duplicate email:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	hashedPw, err := pkg.HashPassword(payload.Password)
	if err != nil {
		log.Println("Error hashing password", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	userID := uuid.New().String()
	_, err = h.db.Exec(
		"INSERT INTO users (id, email, password) VALUES (?, ?, ?)",
		userID, payload.Email, hashedPw,
	)
	if err != nil {
		log.Println("Error inserting user:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user-id": userID,
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

// Login handles user authentication and returns a JWT token upon successful login.
func (h *Handler) Login(c *fiber.Ctx) error {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	payload.Email = strings.TrimSpace(payload.Email)
	payload.Password = strings.TrimSpace(payload.Password)
	payload.Email = strings.ToLower(payload.Email)

	if payload.Email == "" || payload.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password cannot be empty"})
	}

	var userID string
	var hashedPw string

	err := h.db.QueryRow(
		"SELECT id, password FROM users WHERE email = ?",
		payload.Email,
	).Scan(&userID, &hashedPw)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		log.Println("DB error during login:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if err := pkg.CheckPasswordHash(payload.Password, hashedPw); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user-id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := h.jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := h.jwt.SignedString(token, []byte(secret))
	if err != nil {
		log.Println("JWT signing error:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token": signedToken,
	})
}
