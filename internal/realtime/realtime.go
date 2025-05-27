// Package realtime provides real-time functionality for the application
// It handles WebSocket connections for note collaboration
package realtime

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocketConn defines the interface for WebSocket connections
type WebSocketConn interface {
	WriteMessage(messageType int, message []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// RoomManager handles WebSocket room management with thread safety
type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]map[WebSocketConn]bool
}

// NewRoomManager creates a new RoomManager instance
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]map[WebSocketConn]bool),
	}
}

// Global singleton room manager
var manager = NewRoomManager()

// JoinRoom adds a connection to a specific note room
func (rm *RoomManager) JoinRoom(noteID string, conn WebSocketConn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[noteID]; !exists {
		rm.rooms[noteID] = make(map[WebSocketConn]bool)
		log.Printf("Created new note room: %s", noteID)
	}

	rm.rooms[noteID][conn] = true
}

// LeaveRoom removes a connection from a specific note room
// Returns true if the room is now empty and was removed
func (rm *RoomManager) LeaveRoom(noteID string, conn WebSocketConn) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[noteID]
	if !exists {
		return false
	}

	delete(room, conn)
	if len(room) == 0 {
		delete(rm.rooms, noteID)
		log.Printf("Removed empty note room: %s", noteID)
		return true
	}

	return false
}

// BroadcastToRoom sends a message to all connections in a room except the sender
func (rm *RoomManager) BroadcastToRoom(noteID string, sender WebSocketConn, messageType int, message []byte) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[noteID]
	if !exists {
		return
	}

	for conn := range room {
		if conn != sender {
			if err := conn.WriteMessage(messageType, message); err != nil {
				log.Printf("Broadcast error to a client in room %s: %v", noteID, err)
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections for note collaboration
func HandleWebSocket(c *fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		noteID := c.Params("id")
		if noteID == "" {
			if err := c.WriteJSON(fiber.Map{
				"error": "Missing note ID",
			}); err != nil {
				log.Printf("Error sending missing note ID message: %v", err)
			}
			return
		}

		if _, ok := c.Locals("user-id").(string); !ok {
			if err := c.WriteJSON(fiber.Map{
				"error": "User ID not found in context",
			}); err != nil {
				log.Printf("Error sending user ID not found message: %v", err)
			}
			return
		}

		manager.JoinRoom(noteID, c)
		log.Println("User joined note room:", noteID)

		// Ensure user is removed from room when connection closes
		defer func() {
			manager.LeaveRoom(noteID, c)
		}()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			// Broadcast message to all users in the room
			manager.BroadcastToRoom(noteID, c, mt, message)
		}
	})(c)
}
