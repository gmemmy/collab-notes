package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestSignUp uses sqlmock to simulate database interactions and tests the SignUp handler.
func TestSignUp(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Create sqlmock database connection
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database: %v", err)
	}
	defer db.Close()

	// Use the real JWT service defined in auth.go
	jwtService := &JWTService{}

	handler := NewAuthHandler(db, jwtService)

	app := fiber.New()
	app.Post("/signup", handler.SignUp)

	rows := sqlmock.NewRows([]string{"id"})
	mockDB.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	mockDB.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, email, password) VALUES (?, ?, ?)")).
		WithArgs(sqlmock.AnyArg(), "test@example.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error performing request: %v", err)
	}

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Ensure all sqlmock expectations were met
	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestLogin uses sqlmock to simulate database interactions and tests the Login handler.
func TestLogin(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Create sqlmock database connection
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database: %v", err)
	}
	defer db.Close()

	jwtService := &JWTService{}

	handler := NewAuthHandler(db, jwtService)

	app := fiber.New()
	app.Post("/login", handler.Login)

	// Use a valid bcrypt hash for 'password123'
	validHash := "$2a$10$E6HhdnHa3eB0JwE2ZyieJuxCAYpWDYe403HM/LKSPi3FVetNZDk4i"
	rows := sqlmock.NewRows([]string{"id", "password"}).AddRow("user123", validHash)
	mockDB.ExpectQuery(regexp.QuoteMeta("SELECT id, password FROM users WHERE email = ?")).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error performing request: %v", err)
	}

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	if err := mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
