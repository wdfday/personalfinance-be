package service

import (
	"encoding/json"
	"log"
	"personalfinancedss/internal/module/notification/domain"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a WebSocket client connection
type Client struct {
	Hub    *WebSocketHub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID uuid.UUID
}

// WebSocketHub maintains active WebSocket connections and broadcasts messages
type WebSocketHub struct {
	// Registered clients mapped by user ID
	clients map[uuid.UUID]map[*Client]bool

	// Inbound messages from clients
	broadcast chan *BroadcastMessage

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	UserID  uuid.UUID
	Message []byte
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered for user %s (total: %d)", client.UserID, len(h.clients[client.UserID]))

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered for user %s", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.clients[message.UserID]
			h.mu.RUnlock()

			if ok {
				for client := range clients {
					select {
					case client.Send <- message.Message:
					default:
						// Client's send buffer is full, close connection
						h.mu.Lock()
						close(client.Send)
						delete(h.clients[message.UserID], client)
						if len(h.clients[message.UserID]) == 0 {
							delete(h.clients, message.UserID)
						}
						h.mu.Unlock()
					}
				}
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *WebSocketHub) SendToUser(userID uuid.UUID, messageType string, payload interface{}) {
	message := domain.WebSocketMessage{
		Type:    messageType,
		Payload: map[string]interface{}{},
	}

	// Convert payload to map
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		message.Payload = payloadMap
	} else {
		// Try to marshal and unmarshal to convert to map
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload: %v", err)
			return
		}
		if err := json.Unmarshal(jsonBytes, &message.Payload); err != nil {
			log.Printf("Failed to unmarshal payload: %v", err)
			return
		}
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: messageBytes,
	}
}

// GetConnectedUsers returns list of connected user IDs
func (h *WebSocketHub) GetConnectedUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserConnected checks if a user has active WebSocket connections
func (h *WebSocketHub) IsUserConnected(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., acknowledgments, client requests)
		log.Printf("Received message from user %s: %s", c.UserID, string(message))
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
