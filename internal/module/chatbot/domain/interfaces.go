package domain

import (
	"context"

	"github.com/google/uuid"
)

// ============================================
// LLM PROVIDER INTERFACES
// ============================================

// LLMProvider defines the interface for LLM providers (Gemini, Claude, etc.)
type LLMProvider interface {
	// Name returns provider identifier
	Name() string

	// Chat sends messages and returns response
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// ChatStream sends messages and streams response
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)

	// CountTokens estimates token count for messages
	CountTokens(ctx context.Context, messages []Message) (int, error)

	// GetModels returns available models
	GetModels() []ModelInfo

	// SupportsTools returns if provider supports function calling
	SupportsTools() bool
}

// LLMProviderFactory creates and manages LLM providers
type LLMProviderFactory interface {
	// GetProvider returns a provider by type
	GetProvider(providerType string) (LLMProvider, error)

	// GetDefaultProvider returns the default provider
	GetDefaultProvider() LLMProvider

	// RegisterProvider registers a provider
	RegisterProvider(providerType string, provider LLMProvider)

	// SetDefaultProvider sets the default provider
	SetDefaultProvider(providerType string) error

	// ListProviders returns all registered provider names
	ListProviders() []string
}

// ============================================
// TOOL INTERFACES
// ============================================

// ToolExecutor executes tools/functions called by the LLM
type ToolExecutor interface {
	// Execute runs a tool and returns the result
	Execute(ctx context.Context, toolName string, args map[string]any) (ToolExecutionResult, error)

	// GetAvailableTools returns all available tools
	GetAvailableTools() []ToolDefinition
}

// ToolDefinition describes a tool that can be called by the LLM
type ToolDefinition interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Parameters returns the JSON schema for parameters
	Parameters() map[string]any
}

// ToolExecutionResult represents the result of a tool execution
type ToolExecutionResult interface {
	// Content returns the result content
	Content() string

	// IsError returns true if execution failed
	IsError() bool

	// Error returns the error if any
	Error() error
}

// ============================================
// REPOSITORY INTERFACES
// ============================================

// ConversationRepository manages conversation persistence
type ConversationRepository interface {
	// Create creates a new conversation
	Create(ctx context.Context, conv *Conversation) error

	// GetByID retrieves a conversation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Conversation, error)

	// GetByUserID retrieves conversations for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, status string) ([]Conversation, int64, error)

	// Update updates a conversation
	Update(ctx context.Context, conv *Conversation) error

	// Delete soft deletes a conversation
	Delete(ctx context.Context, id uuid.UUID) error
}

// MessageRepository manages message persistence
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, msg *Message) error

	// GetByID retrieves a message by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Message, error)

	// GetByConversationID retrieves messages for a conversation
	GetByConversationID(ctx context.Context, convID uuid.UUID, limit int) ([]Message, error)

	// GetLastMessage retrieves the last message in a conversation
	GetLastMessage(ctx context.Context, convID uuid.UUID) (*Message, error)

	// CountByConversationID counts messages in a conversation
	CountByConversationID(ctx context.Context, convID uuid.UUID) (int64, error)
}

// ============================================
// SERVICE INTERFACES
// ============================================

// ChatService handles chat operations
type ChatService interface {
	// Chat processes a message and returns a response
	Chat(ctx context.Context, req ChatServiceRequest) (*ChatServiceResponse, error)

	// ChatStream processes a message and streams the response
	ChatStream(ctx context.Context, req ChatServiceRequest) (<-chan StreamEvent, error)

	// ListConversations returns user's conversations
	ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int, status string) (*ConversationListResult, error)

	// GetConversation returns a conversation with messages
	GetConversation(ctx context.Context, userID, convID uuid.UUID, messageLimit int) (*ConversationDetailResult, error)

	// UpdateConversation updates a conversation
	UpdateConversation(ctx context.Context, userID, convID uuid.UUID, title *string, status *string) (*Conversation, error)

	// DeleteConversation deletes a conversation
	DeleteConversation(ctx context.Context, userID, convID uuid.UUID) error
}

// ============================================
// SERVICE REQUEST/RESPONSE TYPES
// ============================================

// ChatServiceRequest represents a chat request to the service
type ChatServiceRequest struct {
	UserID         uuid.UUID
	Message        string
	ConversationID *uuid.UUID
	Provider       string
}

// ChatServiceResponse represents a chat response from the service
type ChatServiceResponse struct {
	ConversationID string
	Message        string
	Usage          *TokenUsage
	ToolsUsed      []string
}

// ConversationListResult represents a list of conversations
type ConversationListResult struct {
	Conversations []Conversation
	Total         int
	HasMore       bool
}

// ConversationDetailResult represents a conversation with messages
type ConversationDetailResult struct {
	Conversation Conversation
	Messages     []Message
}
