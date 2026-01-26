package provider

import (
	"context"
	"fmt"
	"strings"

	"personalfinancedss/internal/module/chatbot/domain"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// GenAIConfig holds configuration for Google GenAI provider
type GenAIConfig struct {
	APIKey       string
	Model        string
	SystemPrompt string
}

// genaiProvider implements domain.LLMProvider using Google GenAI SDK
type genaiProvider struct {
	client       *genai.Client
	model        string
	systemPrompt string
}

// DefaultSystemPrompt is the default system prompt for financial advisor
const DefaultSystemPrompt = `Bạn là AI Financial Advisor - trợ lý tài chính thông minh cho hệ thống Personal Finance DSS.

**Vai trò của bạn:**
- Tư vấn tài chính cá nhân bằng tiếng Việt
- Phân tích và đưa ra khuyến nghị dựa trên dữ liệu người dùng
- Sử dụng các công cụ phân tích khi cần thiết

**Nguyên tắc:**
1. Luôn trả lời bằng tiếng Việt
2. Đưa ra lời khuyên cụ thể, có thể hành động được
3. Giải thích rõ ràng các khái niệm tài chính
4. Sử dụng emoji để làm nội dung dễ đọc hơn
5. Khi cần dữ liệu người dùng, hãy gọi tool get_user_financial_summary

**Các chủ đề bạn có thể tư vấn:**
- Phân bổ ngân sách (quy tắc 50/30/20)
- Chiến lược trả nợ (Avalanche, Snowball)
- Mục tiêu tài chính (SMART goals)
- Quỹ khẩn cấp
- Đầu tư cơ bản
- Cân bằng tiết kiệm và trả nợ`

// NewGenAIProvider creates a new Google GenAI provider
func NewGenAIProvider(cfg *GenAIConfig) (domain.LLMProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("genai API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}

	return &genaiProvider{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
	}, nil
}

func (p *genaiProvider) Name() string {
	return "gemini"
}

func (p *genaiProvider) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	// Build contents from messages
	contents := p.buildContents(req.Messages)

	// Create config
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: p.getSystemPrompt(req.SystemPrompt)}},
		},
	}

	if req.Temperature > 0 {
		temp := float32(req.Temperature)
		config.Temperature = &temp
	}

	if req.MaxTokens > 0 {
		maxTokens := int32(req.MaxTokens)
		config.MaxOutputTokens = maxTokens
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		config.Tools = p.convertTools(req.Tools)
	}

	// Generate content
	resp, err := p.client.Models.GenerateContent(ctx, p.model, contents, config)
	if err != nil {
		return domain.ChatResponse{}, fmt.Errorf("genai chat failed: %w", err)
	}

	// Extract response
	content := p.extractContent(resp)
	usage := p.extractUsage(resp)

	return domain.ChatResponse{
		Message: domain.Message{
			ID:           uuid.New(),
			Role:         domain.RoleAssistant,
			Content:      content,
			Model:        p.model,
			FinishReason: "stop",
		},
		Usage:        usage,
		FinishReason: "stop",
	}, nil
}

func (p *genaiProvider) ChatStream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamEvent, error) {
	ch := make(chan domain.StreamEvent)

	go func() {
		defer close(ch)

		// Build contents
		contents := p.buildContents(req.Messages)

		// Create config
		config := &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: p.getSystemPrompt(req.SystemPrompt)}},
			},
		}

		if req.Temperature > 0 {
			temp := float32(req.Temperature)
			config.Temperature = &temp
		}

		if len(req.Tools) > 0 {
			config.Tools = p.convertTools(req.Tools)
		}

		// Send start event
		ch <- domain.StreamEvent{Type: domain.EventStart}

		// Stream response
		for resp := range p.client.Models.GenerateContentStream(ctx, p.model, contents, config) {
			// Process candidates
			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						if part.Text != "" {
							ch <- domain.StreamEvent{
								Type:  domain.EventDelta,
								Delta: part.Text,
							}
						}
						if part.FunctionCall != nil {
							ch <- domain.StreamEvent{
								Type: domain.EventToolCall,
								Message: &domain.Message{
									Role:    domain.RoleTool,
									Content: fmt.Sprintf("Calling %s...", part.FunctionCall.Name),
								},
							}
						}
					}
				}
			}

			// Check for usage
			if resp.UsageMetadata != nil {
				ch <- domain.StreamEvent{
					Type: domain.EventEnd,
					Usage: &domain.TokenUsage{
						PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
						CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
						TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
					},
				}
			}
		}
	}()

	return ch, nil
}

func (p *genaiProvider) CountTokens(ctx context.Context, messages []domain.Message) (int, error) {
	contents := p.buildContents(messages)
	resp, err := p.client.Models.CountTokens(ctx, p.model, contents, nil)
	if err != nil {
		return 0, fmt.Errorf("count tokens failed: %w", err)
	}
	return int(resp.TotalTokens), nil
}

func (p *genaiProvider) GetModels() []domain.ModelInfo {
	return []domain.ModelInfo{
		{
			ID:            "gemini-2.0-flash",
			Name:          "Gemini 2.0 Flash",
			MaxTokens:     1048576,
			SupportsTools: true,
		},
		{
			ID:            "gemini-1.5-pro-latest",
			Name:          "Gemini 1.5 Pro",
			MaxTokens:     2097152,
			SupportsTools: true,
		},
		{
			ID:            "gemini-1.5-flash-latest",
			Name:          "Gemini 1.5 Flash",
			MaxTokens:     1048576,
			SupportsTools: true,
		},
	}
}

func (p *genaiProvider) SupportsTools() bool {
	return true
}

// Helper methods

func (p *genaiProvider) getSystemPrompt(override string) string {
	if override != "" {
		return override
	}
	return p.systemPrompt
}

func (p *genaiProvider) buildContents(messages []domain.Message) []*genai.Content {
	var contents []*genai.Content

	for _, msg := range messages {
		if msg.Role == domain.RoleSystem {
			continue // System messages handled separately
		}

		role := "user"
		if msg.Role == domain.RoleAssistant {
			role = "model"
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	return contents
}

func (p *genaiProvider) extractContent(resp *genai.GenerateContentResponse) string {
	var content strings.Builder

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					content.WriteString(part.Text)
				}
			}
		}
	}

	return content.String()
}

func (p *genaiProvider) extractUsage(resp *genai.GenerateContentResponse) domain.TokenUsage {
	if resp.UsageMetadata == nil {
		return domain.TokenUsage{}
	}

	return domain.TokenUsage{
		PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
		CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
		TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
	}
}

func (p *genaiProvider) convertTools(tools []domain.Tool) []*genai.Tool {
	var genaiTools []*genai.Tool

	for _, tool := range tools {
		fd := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  p.convertParameters(tool.Parameters),
		}

		genaiTools = append(genaiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{fd},
		})
	}

	return genaiTools
}

func (p *genaiProvider) convertParameters(params domain.ToolParameters) *genai.Schema {
	properties := make(map[string]*genai.Schema)

	for name, prop := range params.Properties {
		properties[name] = p.convertProperty(prop)
	}

	return &genai.Schema{
		Type:       genai.TypeObject,
		Properties: properties,
		Required:   params.Required,
	}
}

func (p *genaiProvider) convertProperty(prop domain.ToolProperty) *genai.Schema {
	schema := &genai.Schema{
		Description: prop.Description,
	}

	switch prop.Type {
	case "string":
		schema.Type = genai.TypeString
		if len(prop.Enum) > 0 {
			schema.Enum = prop.Enum
		}
	case "number":
		schema.Type = genai.TypeNumber
	case "integer":
		schema.Type = genai.TypeInteger
	case "boolean":
		schema.Type = genai.TypeBoolean
	case "array":
		schema.Type = genai.TypeArray
		if prop.Items != nil {
			schema.Items = p.convertProperty(*prop.Items)
		}
	case "object":
		schema.Type = genai.TypeObject
	}

	return schema
}
