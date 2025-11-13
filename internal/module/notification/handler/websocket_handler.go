package handler

import (
	"log"
	"net/http"
	"personalfinancedss/internal/middleware"
	notificationservice "personalfinancedss/internal/module/notification/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin (configure based on your needs)
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add proper origin checking in production
		return true
	},
}

// WebSocketHandler manages WebSocket connections
type WebSocketHandler struct {
	hub *notificationservice.WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *notificationservice.WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// RegisterRoutes registers WebSocket routes
func (h *WebSocketHandler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	ws := r.Group("/api/v1/ws")
	ws.Use(authMiddleware.AuthMiddleware())
	{
		ws.GET("/notifications", h.handleWebSocket)
	}

	// Optional: Status endpoint to check WebSocket health
	r.GET("/api/v1/ws/status", h.getWebSocketStatus)
}

// handleWebSocket godoc
// @Summary WebSocket endpoint for real-time notifications
// @Description Establish WebSocket connection to receive real-time notifications
// @Tags websocket
// @Security BearerAuth
// @Router /api/v1/ws/notifications [get]
func (h *WebSocketHandler) handleWebSocket(c *gin.Context) {
	// Get authenticated user from context
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create client and register with hub
	client := &notificationservice.Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: currentUser.ID,
	}

	h.hub.Register <- client

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()

	log.Printf("WebSocket connection established for user: %s", currentUser.ID)
}

// getWebSocketStatus godoc
// @Summary Get WebSocket service status
// @Description Get current WebSocket service statistics
// @Tags websocket
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ws/status [get]
func (h *WebSocketHandler) getWebSocketStatus(c *gin.Context) {
	connectedUsers := h.hub.GetConnectedUsers()

	shared.RespondWithSuccess(c, http.StatusOK, "WebSocket status retrieved successfully", gin.H{
		"status":           "online",
		"connected_users":  len(connectedUsers),
		"websocket_server": "running",
	})
}
