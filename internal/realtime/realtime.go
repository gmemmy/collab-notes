// Package realtime provides real-time functionality for the application
// It handles WebSocket connections for note collaboration
package realtime

import (
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// RoomManager handles WebSocket room management with thread safety
type RoomManager struct {
	// Mutex to protect concurrent access to rooms
	mu sync.RWMutex
	rooms map[string]map[*websocket.Conn]bool
}

// NewRoomManager creates a new RoomManager instance
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

// Global singleton room manager
var manager = NewRoomManager()

// JoinRoom adds a connection to a specific note room
func (rm *RoomManager) JoinRoom(noteID string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Create the room if it doesn't exist
	if _, exists := rm.rooms[noteID]; !exists {
		rm.rooms[noteID] = make(map[*websocket.Conn]bool)
		log.Printf("Created new note room: %s", noteID)
	}

	// Add the connection to the room
	rm.rooms[noteID][conn] = true
}

// LeaveRoom removes a connection from a specific note room
// Returns true if the room is now empty and was removed
func (rm *RoomManager) LeaveRoom(noteID string, conn *websocket.Conn) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[noteID]
	if !exists {
		return false
	}

	// Remove connection from the room
	delete(room, conn)

	// If room is empty, remove it
	if len(room) == 0 {
		delete(rm.rooms, noteID)
		log.Printf("Removed empty note room: %s", noteID)
		return true
	}

	return false
}

// BroadcastToRoom sends a message to all connections in a room except the sender
func (rm *RoomManager) BroadcastToRoom(noteID string, sender *websocket.Conn, messageType int, message []byte) {
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

// HandleWebSocket upgrades the connection and manages note room
func HandleWebSocket(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	noteID := c.Params("id")
	if noteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing note ID",
		})
	}


	userID, ok := c.Locals("user-id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}

	// Upgrade the connection to WebSocket
	return websocket.New(func(conn *websocket.Conn) {
		// Recover from any panics to prevent the server from crashing
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in WebSocket handler: %v", r)
			}
		}()

		// Add user to the note room
		manager.JoinRoom(noteID, conn)
		log.Printf("User %s joined room: %s", userID, noteID)

		// Ensure connection is closed and user is removed from room
		defer func() {
			roomRemoved := manager.LeaveRoom(noteID, conn)
			conn.Close()
			
			if !roomRemoved {
				log.Printf("User %s left room: %s", userID, noteID)
			} else {
				log.Printf("User %s left and room %s was removed", userID, noteID)
			}
		}()

		welcomeMsg := fmt.Sprintf("Connected to note room: %s", noteID)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg)); err != nil {
			log.Printf("Error sending welcome message: %v", err)
			return
		}

		// Main message loop
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Unexpected close error: %v", err)
				}
				break
			}

			manager.BroadcastToRoom(noteID, conn, mt, msg)
		}
	})(c)
}
