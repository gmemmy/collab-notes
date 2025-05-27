// Package realtime provides real-time functionality for the application
// It handles WebSocket connections for note collaboration
package realtime

import (
	"encoding/json"
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

// PresenceMessage represents a presence update message (join/leave)
type PresenceMessage struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	UserID string `json:"user-id"`
}

// IncomingMessage represents a message from a client
type IncomingMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
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

		userIDInterface := c.Locals("user-id")
		userID, ok := userIDInterface.(string)
		if !ok {
			if err := c.WriteJSON(fiber.Map{
				"error": "User ID not found in context",
			}); err != nil {
				log.Printf("Error sending user ID not found message: %v", err)
			}
			return
		}

		joinPayload, _ := json.Marshal(PresenceMessage{
			Type:   "presence",
			Action: "join",
			UserID: userID,
		})
		manager.JoinRoom(noteID, c)
		manager.BroadcastToRoom(noteID, c, websocket.TextMessage, joinPayload)
		log.Println("User joined note room:", noteID)

		// Ensure user is removed from room when connection closes
		defer func() {
			leavePayload, _ := json.Marshal(PresenceMessage{
				Type:   "presence",
				Action: "leave",
				UserID: userID,
			})
			manager.LeaveRoom(noteID, c)
			manager.BroadcastToRoom(noteID, c, websocket.TextMessage, leavePayload)
			log.Println("User left note room:", noteID)
		}()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			var m IncomingMessage
			if err := json.Unmarshal(message, &m); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			if m.Type == "" || m.Content == "" {
				log.Printf("Invalid message received: missing type or content")
				continue
			}

			outgoing := map[string]string{
				"type":    m.Type,
				"content": m.Content,
				"user-id": userID,
			}
			outgoingJSON, err := json.Marshal(outgoing)
			if err != nil {
				log.Printf("Error marshalling outgoing message: %v", err)
				continue
			}

			// Broadcast message to all users in the room
			manager.BroadcastToRoom(noteID, c, mt, outgoingJSON)
		}
	})(c)
}
