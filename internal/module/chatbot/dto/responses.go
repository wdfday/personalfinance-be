package dto

import (
	"time"

	"personalfinancedss/internal/module/chatbot/domain"

	"github.com/google/uuid"
)

// ============================================
// REQUEST DTOs
// ============================================

// ChatRequest represents a chat API request
type ChatRequest struct {
	Message        string  `json:"message" binding:"required"`
	ConversationID *string `json:"conversation_id,omitempty"`
	Provider       *string `json:"provider,omitempty"`
}

// UpdateConversationRequest represents an update conversation request
type UpdateConversationRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ListConversationsQuery represents query params for listing conversations
type ListConversationsQuery struct {
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
	Status string `form:"status"`
}

// GetConversationQuery represents query params for getting a conversation
type GetConversationQuery struct {
	MessageLimit int `form:"message_limit"`
}

// ============================================
// RESPONSE DTOs
// ============================================

// ChatResponse represents a chat API response
type ChatResponse struct {
	ConversationID string      `json:"conversation_id"`
	Message        string      `json:"message"`
	Usage          *TokenUsage `json:"usage,omitempty"`
	ToolsUsed      []string    `json:"tools_used,omitempty"`
}

// TokenUsage represents token usage in response
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ConversationResponse represents a conversation in API response
type ConversationResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Provider     string    `json:"provider,omitempty"`
	Model        string    `json:"model,omitempty"`
	LastMessage  string    `json:"last_message,omitempty"`
	MessageCount int       `json:"message_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MessageResponse represents a message in API response
type MessageResponse struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	TokenCount     int       `json:"token_count,omitempty"`
	LatencyMs      int64     `json:"latency_ms,omitempty"`
	Model          string    `json:"model,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationListResponse represents a list of conversations
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int                    `json:"total"`
	HasMore       bool                   `json:"has_more"`
}

// ConversationDetailResponse represents a conversation with messages
type ConversationDetailResponse struct {
	Conversation ConversationResponse `json:"conversation"`
	Messages     []MessageResponse    `json:"messages"`
}

// ============================================
// CONVERSION HELPERS
// ============================================

// ToTokenUsage converts domain.TokenUsage to dto.TokenUsage
func ToTokenUsage(usage *domain.TokenUsage) *TokenUsage {
	if usage == nil {
		return nil
	}
	return &TokenUsage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

// ToConversationResponse converts domain.Conversation to dto.ConversationResponse
func ToConversationResponse(conv *domain.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:        conv.ID,
		UserID:    conv.UserID,
		Title:     conv.Title,
		Status:    string(conv.Status),
		Provider:  conv.Provider,
		Model:     conv.Model,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}
}

// ToConversationResponses converts a slice of domain.Conversation to dto.ConversationResponse
func ToConversationResponses(convs []domain.Conversation) []ConversationResponse {
	responses := make([]ConversationResponse, len(convs))
	for i, conv := range convs {
		responses[i] = ToConversationResponse(&conv)
	}
	return responses
}

// ToMessageResponse converts domain.Message to dto.MessageResponse
func ToMessageResponse(msg *domain.Message) MessageResponse {
	return MessageResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		Role:           string(msg.Role),
		Content:        msg.Content,
		TokenCount:     msg.TokenCount,
		LatencyMs:      msg.LatencyMs,
		Model:          msg.Model,
		CreatedAt:      msg.CreatedAt,
	}
}

// ToMessageResponses converts a slice of domain.Message to dto.MessageResponse
func ToMessageResponses(msgs []domain.Message) []MessageResponse {
	responses := make([]MessageResponse, len(msgs))
	for i, msg := range msgs {
		responses[i] = ToMessageResponse(&msg)
	}
	return responses
}
