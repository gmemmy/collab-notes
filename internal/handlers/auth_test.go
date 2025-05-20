package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

type MockJWT struct {
	mock.Mock
}

func (m *MockDB) Exec(query string, args ...any) (sql.Result, error) {
	callArgs := m.Called(query, args)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockJWT) NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token {
	callArgs := m.Called(method, claims)
	return callArgs.Get(0).(*jwt.Token)
}

func (m *MockJWT) SignedString(token *jwt.Token, key []byte) (string, error) {
	callArgs := m.Called(token, key)
	return callArgs.String(0), callArgs.Error(1)
}

type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	callArgs := m.Called()
	return callArgs.Get(0).(int64), callArgs.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	callArgs := m.Called()
	return callArgs.Get(0).(int64), callArgs.Error(1)
}

func TestSignUp(t *testing.T) {
	// Set up the test
	app := fiber.New()
	mockDB := new(MockDB)
	mockJWT := new(MockJWT)
	mockResult := new(MockResult)

	// Create a new handler with mock dependencies
	h := NewAuthHandler(mockDB, mockJWT)

	app.Post("/signup", h.SignUp)

	mockResult.On("LastInsertId").Return(int64(1), nil)
	mockDB.On("Exec", "INSERT INTO users (id, email, password) VALUES (?, ?, ?)", mock.MatchedBy(func(args []any) bool {
		return len(args) == 3 && args[1] == "test@example.com"
	})).Return(mockResult, nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	mockJWT.On("NewWithClaims", jwt.SigningMethodHS256, mock.Anything).Return(token)
	mockJWT.On("SignedString", token, mock.Anything).Return("test-token", nil)

	payload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	// Call the handler
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockDB.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
	mockResult.AssertExpectations(t)
}
