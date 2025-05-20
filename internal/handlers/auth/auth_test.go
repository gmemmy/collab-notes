package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// testHelper contains common test setup and utilities
type testHelper struct {
	t       *testing.T
	db      *sql.DB
	mockDB  sqlmock.Sqlmock
	app     *fiber.App
	handler *Handler
}

// newTestHelper creates a new test helper with common setup
func newTestHelper(t *testing.T) *testHelper {
	if err := os.Setenv("JWT_SECRET", "test-secret"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}

	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database: %v", err)
	}

	jwtService := &JWTService{}
	handler := NewHandler(db, jwtService)
	app := fiber.New()

	return &testHelper{
		t:       t,
		db:      db,
		mockDB:  mockDB,
		app:     app,
		handler: handler,
	}
}

// cleanup performs cleanup after tests
func (h *testHelper) cleanup() {
	if err := os.Unsetenv("JWT_SECRET"); err != nil {
		h.t.Logf("failed to unset environment variable: %v", err)
	}
	// NOTE: Don't close the database connection here as sqlmock
	// automatically closes it after expectations are met
}

// setupRoute sets up a route for testing
func (h *testHelper) setupRoute(method, path string, handler fiber.Handler) {
	switch method {
	case "POST":
		h.app.Post(path, handler)
	}
}

func TestSignUp(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("POST", "/signup", helper.handler.SignUp)

	testCases := []struct {
		name           string
		payload        map[string]string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockRows:       sqlmock.NewRows([]string{"id"}),
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "Duplicate Email",
			payload: map[string]string{
				"email":    "existing@example.com",
				"password": "password123",
			},
			mockRows:       sqlmock.NewRows([]string{"id"}).AddRow("existing-user-id"),
			expectedStatus: fiber.StatusConflict,
			expectedError:  "Email already in use",
		},
		{
			name: "Database Error",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name: "Invalid Email",
			payload: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name: "Short Password",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "short",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters long",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip database expectations for cases that should fail at validation
			skipDbSetup := tc.name == "Invalid Email" || tc.name == "Short Password"

			if !skipDbSetup {
				// Setup mock expectations
				query := regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")
				if tc.mockError != nil {
					helper.mockDB.ExpectQuery(query).WithArgs(tc.payload["email"]).WillReturnError(tc.mockError)
				} else {
					helper.mockDB.ExpectQuery(query).WithArgs(tc.payload["email"]).WillReturnRows(tc.mockRows)
				}
			}

			if !skipDbSetup && tc.expectedStatus == fiber.StatusOK {
				helper.mockDB.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, email, password) VALUES (?, ?, ?)")).
					WithArgs(sqlmock.AnyArg(), tc.payload["email"], sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("error marshaling payload: %v", err)
			}

			req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := helper.app.Test(req)
			if err != nil {
				t.Fatalf("error performing request: %v", err)
			}

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedError != "" {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.Equal(t, tc.expectedError, response["error"])
			} else if tc.expectedStatus == fiber.StatusOK {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.NotEmpty(t, response["token"])
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestLogin(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("POST", "/login", helper.handler.Login)

	// Use a valid bcrypt hash for 'password123'
	validHash := "$2a$10$E6HhdnHa3eB0JwE2ZyieJuxCAYpWDYe403HM/LKSPi3FVetNZDk4i"

	testCases := []struct {
		name           string
		payload        map[string]string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockRows:       sqlmock.NewRows([]string{"id", "password"}).AddRow("user123", validHash),
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "Invalid Credentials",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockRows:       sqlmock.NewRows([]string{"id", "password"}).AddRow("user123", validHash),
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  "Invalid credentials",
		},
		{
			name: "User Not Found",
			payload: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			mockRows:       sqlmock.NewRows([]string{"id", "password"}),
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  "Invalid credentials",
		},
		{
			name: "Database Error",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name: "Empty Credentials",
			payload: map[string]string{
				"email":    "",
				"password": "",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Email and password cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip database expectations for cases that should fail at validation
			skipDbSetup := tc.name == "Empty Credentials"

			if !skipDbSetup {
				query := regexp.QuoteMeta("SELECT id, password FROM users WHERE email = ?")
				if tc.mockError != nil {
					helper.mockDB.ExpectQuery(query).WithArgs(tc.payload["email"]).WillReturnError(tc.mockError)
				} else {
					helper.mockDB.ExpectQuery(query).WithArgs(tc.payload["email"]).WillReturnRows(tc.mockRows)
				}
			}

			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("error marshaling payload: %v", err)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := helper.app.Test(req)
			if err != nil {
				t.Fatalf("error performing request: %v", err)
			}

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedError != "" {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.Equal(t, tc.expectedError, response["error"])
			} else if tc.expectedStatus == fiber.StatusOK {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.NotEmpty(t, response["token"])
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
