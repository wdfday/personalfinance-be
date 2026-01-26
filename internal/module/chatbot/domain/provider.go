package domain

// ============================================
// CHAT REQUEST/RESPONSE TYPES
// ============================================

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages     []Message `json:"messages"`
	Tools        []Tool    `json:"tools,omitempty"`
	Model        string    `json:"model,omitempty"`
	Temperature  float64   `json:"temperature,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Message      Message    `json:"message"`
	Usage        TokenUsage `json:"usage"`
	FinishReason string     `json:"finish_reason"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ============================================
// STREAMING TYPES
// ============================================

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type    StreamEventType `json:"type"`
	Delta   string          `json:"delta,omitempty"`
	Message *Message        `json:"message,omitempty"`
	Usage   *TokenUsage     `json:"usage,omitempty"`
	Error   error           `json:"-"`
}

// StreamEventType represents the type of stream event
type StreamEventType string

const (
	EventStart      StreamEventType = "start"
	EventDelta      StreamEventType = "delta"
	EventThinking   StreamEventType = "thinking"
	EventToolCall   StreamEventType = "tool_call"
	EventToolResult StreamEventType = "tool_result"
	EventEnd        StreamEventType = "end"
	EventError      StreamEventType = "error"
)

// ============================================
// MODEL INFO
// ============================================

// ModelInfo contains information about an LLM model
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	MaxTokens     int    `json:"max_tokens"`
	SupportsTools bool   `json:"supports_tools"`
}

// ============================================
// PROVIDER TYPES
// ============================================

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderMock   ProviderType = "mock"
	ProviderGemini ProviderType = "gemini"
	ProviderClaude ProviderType = "claude"
)
