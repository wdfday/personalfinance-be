package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/chatbot/domain"
	"personalfinancedss/internal/module/chatbot/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles chat HTTP requests
type Handler struct {
	chatService domain.ChatService
	logger      *zap.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService domain.ChatService, logger *zap.Logger) *Handler {
	return &Handler{
		chatService: chatService,
		logger:      logger,
	}
}

// Chat handles POST /chat
// @Summary Send a chat message
// @Description Send a message and get AI response
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChatRequest true "Chat request"
// @Success 200 {object} dto.ChatResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /chat [post]
func (h *Handler) Chat(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build service request
	serviceReq := domain.ChatServiceRequest{
		UserID:  userID,
		Message: req.Message,
	}

	if req.ConversationID != nil {
		convID, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation_id"})
			return
		}
		serviceReq.ConversationID = &convID
	}

	if req.Provider != nil {
		serviceReq.Provider = *req.Provider
	}

	// Call service
	resp, err := h.chatService.Chat(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("chat failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ChatResponse{
		ConversationID: resp.ConversationID,
		Message:        resp.Message,
		Usage:          dto.ToTokenUsage(resp.Usage),
		ToolsUsed:      resp.ToolsUsed,
	})
}

// ChatStream handles POST /chat/stream
// @Summary Send a chat message with streaming response
// @Description Send a message and receive streaming AI response via SSE
// @Tags Chatbot
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param request body dto.ChatRequest true "Chat request"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /chat/stream [post]
func (h *Handler) ChatStream(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build service request
	serviceReq := domain.ChatServiceRequest{
		UserID:  userID,
		Message: req.Message,
	}

	if req.ConversationID != nil {
		convID, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation_id"})
			return
		}
		serviceReq.ConversationID = &convID
	}

	if req.Provider != nil {
		serviceReq.Provider = *req.Provider
	}

	// Start streaming
	eventCh, err := h.chatService.ChatStream(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("chat stream failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Stream events
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return false
			}

			data := map[string]any{}
			switch event.Type {
			case domain.EventStart:
				data["conversation_id"] = serviceReq.ConversationID
			case domain.EventDelta:
				data["content"] = event.Delta
			case domain.EventThinking:
				data["thinking"] = event.Delta
			case domain.EventToolCall:
				data["tool"] = "executing"
				if event.Message != nil {
					data["content"] = event.Message.Content
				}
			case domain.EventToolResult:
				data["tool_result"] = true
				if event.Message != nil {
					data["content"] = event.Message.Content
				}
			case domain.EventEnd:
				if event.Usage != nil {
					data["usage"] = event.Usage
				}
			case domain.EventError:
				if event.Error != nil {
					data["error"] = event.Error.Error()
				}
			}

			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, jsonData)
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// ListConversations handles GET /conversations
// @Summary List conversations
// @Description Get user's conversation history
// @Tags Chatbot
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Status filter" Enums(active, archived)
// @Success 200 {object} dto.ConversationListResponse
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /conversations [get]
func (h *Handler) ListConversations(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var query dto.ListConversationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	result, err := h.chatService.ListConversations(c.Request.Context(), userID, query.Limit, query.Offset, query.Status)
	if err != nil {
		h.logger.Error("list conversations failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ConversationListResponse{
		Conversations: dto.ToConversationResponses(result.Conversations),
		Total:         result.Total,
		HasMore:       result.HasMore,
	})
}

// GetConversation handles GET /conversations/:id
// @Summary Get conversation
// @Description Get a specific conversation with messages
// @Tags Chatbot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param message_limit query int false "Message limit" default(50)
// @Success 200 {object} dto.ConversationDetailResponse
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /conversations/{id} [get]
func (h *Handler) GetConversation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var query dto.GetConversationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.MessageLimit == 0 {
		query.MessageLimit = 50
	}

	result, err := h.chatService.GetConversation(c.Request.Context(), userID, convID, query.MessageLimit)
	if err != nil {
		h.logger.Error("get conversation failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ConversationDetailResponse{
		Conversation: dto.ToConversationResponse(&result.Conversation),
		Messages:     dto.ToMessageResponses(result.Messages),
	})
}

// UpdateConversation handles PATCH /conversations/:id
// @Summary Update conversation
// @Description Update conversation title or status
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param request body dto.UpdateConversationRequest true "Update request"
// @Success 200 {object} dto.ConversationResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /conversations/{id} [patch]
func (h *Handler) UpdateConversation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var req dto.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.chatService.UpdateConversation(c.Request.Context(), userID, convID, req.Title, req.Status)
	if err != nil {
		h.logger.Error("update conversation failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToConversationResponse(conv))
}

// DeleteConversation handles DELETE /conversations/:id
// @Summary Delete conversation
// @Description Soft delete a conversation
// @Tags Chatbot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /conversations/{id} [delete]
func (h *Handler) DeleteConversation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	if err := h.chatService.DeleteConversation(c.Request.Context(), userID, convID); err != nil {
		h.logger.Error("delete conversation failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted successfully"})
}

// Helper to get user ID from context
func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, error) {
	// Try to get from context (set by auth middleware)
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			return id, nil
		}
		if idStr, ok := userIDStr.(string); ok {
			return uuid.Parse(idStr)
		}
	}

	// For development: use a mock user ID
	// TODO: Remove this in production
	return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
}

// RegisterRoutes registers chat routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	// Chat endpoints
	chat := r.Group("/api/v1")
	chat.Use(authMiddleware.AuthMiddleware())
	{
		chat.POST("/chat", h.Chat)
		chat.POST("/chat/stream", h.ChatStream)
	}

	// Conversation endpoints
	conversations := r.Group("/api/v1/conversations")
	conversations.Use(authMiddleware.AuthMiddleware())
	{
		conversations.GET("", h.ListConversations)
		conversations.GET("/:id", h.GetConversation)
		conversations.PATCH("/:id", h.UpdateConversation)
		conversations.DELETE("/:id", h.DeleteConversation)
	}
}
