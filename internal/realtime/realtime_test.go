package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketConn is a mock implementation of WebSocketConn
type MockWebSocketConn struct {
	mock.Mock
}

func (m *MockWebSocketConn) WriteMessage(messageType int, message []byte) error {
	args := m.Called(messageType, message)
	return args.Error(0)
}

func (m *MockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	args := m.Called()
	return args.Int(0), args.Get(1).([]byte), args.Error(2)
}

func (m *MockWebSocketConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRoomManager_JoinRoom(t *testing.T) {
	rm := NewRoomManager()
	mockConn := new(MockWebSocketConn)
	noteID := "test-note"

	// Test joining a new room
	rm.JoinRoom(noteID, mockConn)
	assert.Contains(t, rm.rooms, noteID)
	assert.Contains(t, rm.rooms[noteID], mockConn)

	// Test joining an existing room
	mockConn2 := new(MockWebSocketConn)
	rm.JoinRoom(noteID, mockConn2)
	assert.Contains(t, rm.rooms[noteID], mockConn2)
	assert.Equal(t, 2, len(rm.rooms[noteID]))
}

func TestRoomManager_LeaveRoom(t *testing.T) {
	rm := NewRoomManager()
	mockConn := new(MockWebSocketConn)
	noteID := "test-note"

	// Test leaving a non-existent room
	assert.False(t, rm.LeaveRoom(noteID, mockConn))

	// Test leaving an existing room
	rm.JoinRoom(noteID, mockConn)
	assert.True(t, rm.LeaveRoom(noteID, mockConn))
	assert.NotContains(t, rm.rooms, noteID)

	// Test leaving a room with multiple connections
	mockConn1 := new(MockWebSocketConn)
	mockConn2 := new(MockWebSocketConn)
	rm.JoinRoom(noteID, mockConn1)
	rm.JoinRoom(noteID, mockConn2)
	assert.False(t, rm.LeaveRoom(noteID, mockConn1))
	assert.Contains(t, rm.rooms, noteID)
	assert.Equal(t, 1, len(rm.rooms[noteID]))
}

func TestRoomManager_BroadcastToRoom(t *testing.T) {
	rm := NewRoomManager()
	mockConn1 := new(MockWebSocketConn)
	mockConn2 := new(MockWebSocketConn)
	noteID := "test-note"
	message := []byte("test message")

	mockConn1.On("WriteMessage", 1, message).Return(nil)
	mockConn2.On("WriteMessage", 1, message).Return(nil)

	// Test broadcasting to empty room
	rm.BroadcastToRoom(noteID, mockConn1, 1, message)

	// Test broadcasting to room with one connection
	rm.JoinRoom(noteID, mockConn1)
	rm.BroadcastToRoom(noteID, mockConn1, 1, message)

	// Test broadcasting to room with multiple connections
	rm.JoinRoom(noteID, mockConn2)
	rm.BroadcastToRoom(noteID, mockConn1, 1, message)

	// Verify that sender didn't receive the message
	mockConn1.AssertNotCalled(t, "WriteMessage", 1, message)
	// Verify that other connection received the message
	mockConn2.AssertCalled(t, "WriteMessage", 1, message)
}

func TestRoomManager_ConcurrentAccess(t *testing.T) {
	rm := NewRoomManager()
	noteID := "test-note"
	connections := make([]*MockWebSocketConn, 10)

	// Create multiple connections
	for i := range connections {
		connections[i] = new(MockWebSocketConn)
	}

	// Test concurrent joins
	done := make(chan bool)
	for _, conn := range connections {
		go func(c *MockWebSocketConn) {
			rm.JoinRoom(noteID, c)
			done <- true
		}(conn)
	}

	// Wait for all goroutines to complete
	for range connections {
		<-done
	}

	assert.Equal(t, len(connections), len(rm.rooms[noteID]))

	// Test concurrent leaves
	for _, conn := range connections {
		go func(c *MockWebSocketConn) {
			rm.LeaveRoom(noteID, c)
			done <- true
		}(conn)
	}

	for range connections {
		<-done
	}

	// Verify room is empty
	assert.NotContains(t, rm.rooms, noteID)
}
