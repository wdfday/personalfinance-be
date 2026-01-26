package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID   `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	ConversationID uuid.UUID   `gorm:"type:uuid;not null;index" json:"conversation_id"`
	Role           MessageRole `gorm:"type:varchar(20);not null" json:"role"`
	Content        string      `gorm:"type:text" json:"content"`
	ToolCalls      ToolCalls   `gorm:"type:jsonb" json:"tool_calls,omitempty"`
	ToolResults    ToolResults `gorm:"type:jsonb" json:"tool_results,omitempty"`
	TokenCount     int         `gorm:"type:int;default:0" json:"token_count"`
	LatencyMs      int64       `gorm:"type:bigint;default:0" json:"latency_ms"`
	Model          string      `gorm:"type:varchar(100)" json:"model,omitempty"`
	FinishReason   string      `gorm:"type:varchar(50)" json:"finish_reason,omitempty"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for GORM
func (Message) TableName() string {
	return "chatbot_messages"
}

// ToolCall represents a function call made by the LLM
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolCalls is a slice of ToolCall for GORM JSON handling
type ToolCalls []ToolCall

func (tc ToolCalls) Value() (driver.Value, error) {
	if tc == nil {
		return nil, nil
	}
	return json.Marshal(tc)
}

func (tc *ToolCalls) Scan(value interface{}) error {
	if value == nil {
		*tc = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, tc)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// ToolResults is a slice of ToolResult for GORM JSON handling
type ToolResults []ToolResult

func (tr ToolResults) Value() (driver.Value, error) {
	if tr == nil {
		return nil, nil
	}
	return json.Marshal(tr)
}

func (tr *ToolResults) Scan(value interface{}) error {
	if value == nil {
		*tr = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, tr)
}
