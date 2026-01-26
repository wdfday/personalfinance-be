package domain

import (
	"time"

	"github.com/google/uuid"
)

// ConversationStatus represents the status of a conversation
type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusArchived ConversationStatus = "archived"
	ConversationStatusDeleted  ConversationStatus = "deleted"
)

// Conversation represents a chat session
type Conversation struct {
	ID        uuid.UUID          `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID    uuid.UUID          `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string             `gorm:"type:varchar(255)" json:"title"`
	Status    ConversationStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	Provider  string             `gorm:"type:varchar(50)" json:"provider"` // gemini, claude
	Model     string             `gorm:"type:varchar(100)" json:"model"`   // gemini-pro, claude-3
	CreatedAt time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time          `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// TableName returns the table name for GORM
func (Conversation) TableName() string {
	return "chatbot_conversations"
}

// ConversationMeta contains metadata about a conversation
type ConversationMeta struct {
	TotalTokens int            `json:"total_tokens"`
	TotalCost   float64        `json:"total_cost"`
	Tags        []string       `json:"tags,omitempty"`
	CustomData  map[string]any `json:"custom_data,omitempty"`
}
