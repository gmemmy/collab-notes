package notes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

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
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database: %v", err)
	}

	handler := NewHandler(db)
	app := fiber.New()

	// Mock user ID in context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user-id", "user123")
		return c.Next()
	})

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
	// NOTE: Don't close the database connection here as sqlmock
	// automatically closes it after expectations are met
}

// setupRoute sets up a route for testing
func (h *testHelper) setupRoute(method, path string, handler fiber.Handler) {
	switch method {
	case "GET":
		h.app.Get(path, handler)
	case "POST":
		h.app.Post(path, handler)
	case "PUT":
		h.app.Put(path, handler)
	case "DELETE":
		h.app.Delete(path, handler)
	}
}

func TestGetNotes(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("GET", "/notes", helper.handler.GetNotes)

	now := time.Now()
	// Test cases
	testCases := []struct {
		name           string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedStatus int
		expectedNotes  int
		expectedError  string
	}{
		{
			name: "Success",
			mockRows: sqlmock.NewRows([]string{"id", "user_id", "title", "content", "created_at", "updated_at"}).
				AddRow("note1", "user123", "Test Note 1", "Content 1", now, now).
				AddRow("note2", "user123", "Test Note 2", "Content 2", now, now),
			expectedStatus: fiber.StatusOK,
			expectedNotes:  2,
		},
		{
			name:           "Database Error",
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
			expectedNotes:  0,
		},
		{
			name:           "No Notes",
			mockRows:       sqlmock.NewRows([]string{"id", "user_id", "title", "content", "created_at", "updated_at"}),
			expectedStatus: fiber.StatusOK,
			expectedNotes:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := regexp.QuoteMeta("SELECT id, user_id, title, content, created_at, updated_at FROM notes WHERE user_id = ?")
			if tc.mockError != nil {
				helper.mockDB.ExpectQuery(query).WithArgs("user123").WillReturnError(tc.mockError)
			} else {
				helper.mockDB.ExpectQuery(query).WithArgs("user123").WillReturnRows(tc.mockRows)
			}

			req := httptest.NewRequest("GET", "/notes", nil)
			resp, err := helper.app.Test(req)
			if err != nil {
				t.Fatalf("error performing request: %v", err)
			}

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedStatus == fiber.StatusOK {
				var notes []Note
				err = json.NewDecoder(resp.Body).Decode(&notes)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.Len(t, notes, tc.expectedNotes)
			} else if tc.expectedError != "" {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.Equal(t, tc.expectedError, response["error"])
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestCreateNote(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("POST", "/notes", helper.handler.CreateNote)

	testCases := []struct {
		name           string
		payload        map[string]string
		mockError      error
		expectedStatus int
		expectedError  string
		expectQuery    bool
	}{
		{
			name:           "Empty Title",
			payload:        map[string]string{"title": "", "content": "Some content"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Title cannot be empty",
			expectQuery:    false,
		},
		{
			name:           "Whitespace Title",
			payload:        map[string]string{"title": "   ", "content": "Some content"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Title cannot be empty",
			expectQuery:    false,
		},
		{
			name:           "Valid Note",
			payload:        map[string]string{"title": "Valid Title", "content": "Some content"},
			expectedStatus: fiber.StatusCreated,
			expectQuery:    true,
		},
		{
			name:           "Database Error",
			payload:        map[string]string{"title": "Valid Title", "content": "Some content"},
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
			expectQuery:    true,
		},
		{
			name:           "Missing Content",
			payload:        map[string]string{"title": "Valid Title"},
			expectedStatus: fiber.StatusCreated,
			expectQuery:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("error marshaling payload: %v", err)
			}

			if tc.expectQuery {
				query := regexp.QuoteMeta("INSERT INTO notes (id, user_id, title, content) VALUES (?, ?, ?, ?)")
				if tc.mockError != nil {
					helper.mockDB.ExpectExec(query).
						WithArgs(sqlmock.AnyArg(), "user123", tc.payload["title"], tc.payload["content"]).
						WillReturnError(tc.mockError)
				} else {
					helper.mockDB.ExpectExec(query).
						WithArgs(sqlmock.AnyArg(), "user123", tc.payload["title"], tc.payload["content"]).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			req := httptest.NewRequest("POST", "/notes", bytes.NewBuffer(jsonPayload))
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
			} else if tc.expectedStatus == fiber.StatusCreated {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
				assert.NotEmpty(t, response["id"])
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUpdateNote(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("PUT", "/notes/:id", helper.handler.UpdateNote)

	testCases := []struct {
		name           string
		noteID         string
		payload        map[string]string
		mockError      error
		expectedStatus int
		expectedError  string
		expectQuery    bool
		rowsAffected   int64
	}{
		{
			name:           "Successful Update",
			noteID:         "note1",
			payload:        map[string]string{"title": "Updated Title", "content": "Updated content"},
			expectedStatus: fiber.StatusNoContent,
			expectQuery:    true,
			rowsAffected:   1,
		},
		{
			name:           "Empty Title",
			noteID:         "note1",
			payload:        map[string]string{"title": "", "content": "Some content"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Title cannot be empty",
			expectQuery:    false,
		},
		{
			name:           "Whitespace Title",
			noteID:         "note1",
			payload:        map[string]string{"title": "   ", "content": "Some content"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Title cannot be empty",
			expectQuery:    false,
		},
		{
			name:           "Note Not Found",
			noteID:         "nonexistent",
			payload:        map[string]string{"title": "Valid Title", "content": "Some content"},
			expectedStatus: fiber.StatusNotFound,
			expectedError:  "Note not found or unauthorized",
			expectQuery:    true,
			rowsAffected:   0,
		},
		{
			name:           "Database Error",
			noteID:         "note1",
			payload:        map[string]string{"title": "Valid Title", "content": "Some content"},
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
			expectQuery:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("error marshaling payload: %v", err)
			}

			if tc.expectQuery {
				query := regexp.QuoteMeta("UPDATE notes SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?")
				if tc.mockError != nil {
					helper.mockDB.ExpectExec(query).
						WithArgs(tc.payload["title"], tc.payload["content"], tc.noteID, "user123").
						WillReturnError(tc.mockError)
				} else {
					helper.mockDB.ExpectExec(query).
						WithArgs(tc.payload["title"], tc.payload["content"], tc.noteID, "user123").
						WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))
				}
			}

			req := httptest.NewRequest("PUT", "/notes/"+tc.noteID, bytes.NewBuffer(jsonPayload))
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
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDeleteNote(t *testing.T) {
	helper := newTestHelper(t)
	defer helper.cleanup()

	helper.setupRoute("DELETE", "/notes/:id", helper.handler.DeleteNote)

	testCases := []struct {
		name           string
		noteID         string
		mockError      error
		expectedStatus int
		expectedError  string
		rowsAffected   int64
	}{
		{
			name:           "Successful Deletion",
			noteID:         "note1",
			expectedStatus: fiber.StatusNoContent,
			rowsAffected:   1,
		},
		{
			name:           "Note Not Found",
			noteID:         "nonexistent",
			expectedStatus: fiber.StatusNotFound,
			expectedError:  "Note not found or unauthorized",
			rowsAffected:   0,
		},
		{
			name:           "Database Error",
			noteID:         "note1",
			mockError:      errors.New("database error"),
			expectedStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := regexp.QuoteMeta("DELETE FROM notes WHERE id = ? AND user_id = ?")
			if tc.mockError != nil {
				helper.mockDB.ExpectExec(query).
					WithArgs(tc.noteID, "user123").
					WillReturnError(tc.mockError)
			} else {
				helper.mockDB.ExpectExec(query).
					WithArgs(tc.noteID, "user123").
					WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))
			}

			req := httptest.NewRequest("DELETE", "/notes/"+tc.noteID, nil)
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
			}
		})
	}

	if err := helper.mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
