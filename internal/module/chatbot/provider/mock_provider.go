package provider

import (
	"context"
	"strings"
	"time"

	"personalfinancedss/internal/module/chatbot/domain"

	"github.com/google/uuid"
)

// mockProvider implements domain.LLMProvider for development
type mockProvider struct{}

// NewMockProvider creates a new mock provider
func NewMockProvider() domain.LLMProvider {
	return &mockProvider{}
}

func (p *mockProvider) Name() string {
	return "mock"
}

func (p *mockProvider) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	// Get the last user message
	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == domain.RoleUser {
			userMessage = req.Messages[i].Content
			break
		}
	}

	response := generateMockResponse(userMessage)

	return domain.ChatResponse{
		Message: domain.Message{
			ID:           uuid.New(),
			Role:         domain.RoleAssistant,
			Content:      response,
			Model:        "mock-model",
			FinishReason: "stop",
			CreatedAt:    time.Now(),
		},
		Usage: domain.TokenUsage{
			PromptTokens:     len(userMessage) / 4,
			CompletionTokens: len(response) / 4,
			TotalTokens:      (len(userMessage) + len(response)) / 4,
		},
		FinishReason: "stop",
	}, nil
}

func (p *mockProvider) ChatStream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamEvent, error) {
	ch := make(chan domain.StreamEvent)

	go func() {
		defer close(ch)

		// Get the last user message
		var userMessage string
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == domain.RoleUser {
				userMessage = req.Messages[i].Content
				break
			}
		}

		response := generateMockResponse(userMessage)

		// Send start event
		ch <- domain.StreamEvent{Type: domain.EventStart}

		// Stream response word by word
		words := strings.Split(response, " ")
		for i, word := range words {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(30 * time.Millisecond)
				delta := word
				if i < len(words)-1 {
					delta += " "
				}
				ch <- domain.StreamEvent{
					Type:  domain.EventDelta,
					Delta: delta,
				}
			}
		}

		// Send end event
		ch <- domain.StreamEvent{
			Type: domain.EventEnd,
			Usage: &domain.TokenUsage{
				PromptTokens:     len(userMessage) / 4,
				CompletionTokens: len(response) / 4,
				TotalTokens:      (len(userMessage) + len(response)) / 4,
			},
		}
	}()

	return ch, nil
}

func (p *mockProvider) CountTokens(ctx context.Context, messages []domain.Message) (int, error) {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4
	}
	return total, nil
}

func (p *mockProvider) GetModels() []domain.ModelInfo {
	return []domain.ModelInfo{
		{
			ID:            "mock-model",
			Name:          "Mock Model",
			MaxTokens:     4096,
			SupportsTools: true,
		},
	}
}

func (p *mockProvider) SupportsTools() bool {
	return true
}

func generateMockResponse(input string) string {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "ngân sách") || strings.Contains(lower, "budget") || strings.Contains(lower, "phân bổ") {
		return `Để phân bổ ngân sách hiệu quả, tôi khuyên bạn áp dụng quy tắc 50/30/20:

• **50%** cho nhu cầu thiết yếu (nhà ở, ăn uống, đi lại)
• **30%** cho mong muốn (giải trí, mua sắm)
• **20%** cho tiết kiệm và trả nợ

Bạn có muốn tôi tạo kế hoạch chi tiết dựa trên thu nhập của bạn không?`
	}

	if strings.Contains(lower, "nợ") || strings.Contains(lower, "debt") || strings.Contains(lower, "trả nợ") {
		return `Về chiến lược trả nợ, có 2 phương pháp phổ biến:

**1. Avalanche (Tuyết lở):**
• Ưu tiên trả nợ lãi suất cao nhất
• Tiết kiệm tiền lãi nhiều nhất

**2. Snowball (Quả cầu tuyết):**
• Ưu tiên trả nợ nhỏ nhất trước
• Tạo động lực tâm lý

Bạn có bao nhiêu khoản nợ? Tôi sẽ phân tích và đề xuất chiến lược phù hợp.`
	}

	if strings.Contains(lower, "mục tiêu") || strings.Contains(lower, "goal") || strings.Contains(lower, "kế hoạch") {
		return `Để đặt mục tiêu tài chính hiệu quả, hãy áp dụng nguyên tắc SMART:

• **S**pecific - Cụ thể
• **M**easurable - Đo lường được
• **A**chievable - Khả thi
• **R**elevant - Phù hợp
• **T**ime-bound - Có thời hạn

Bạn có mục tiêu tài chính nào muốn tôi giúp lập kế hoạch?`
	}

	if strings.Contains(lower, "đầu tư") || strings.Contains(lower, "investment") {
		return `Về đầu tư, tôi có thể giúp bạn:

• Đánh giá mức độ chấp nhận rủi ro
• Phân bổ danh mục đầu tư
• Tính toán lợi nhuận kỳ vọng
• So sánh các kênh đầu tư

Bạn đã có kinh nghiệm đầu tư chưa? Mức vốn bạn dự định đầu tư là bao nhiêu?`
	}

	if strings.Contains(lower, "khẩn cấp") || strings.Contains(lower, "emergency") || strings.Contains(lower, "dự phòng") {
		return `Quỹ khẩn cấp là nền tảng tài chính quan trọng!

**Khuyến nghị:**
• Tối thiểu: 3 tháng chi tiêu
• Lý tưởng: 6 tháng chi tiêu
• Gia đình có trẻ nhỏ: 9-12 tháng

Quỹ này nên để ở tài khoản tiết kiệm có tính thanh khoản cao. Bạn muốn tôi tính toán số tiền cần thiết không?`
	}

	return `Cảm ơn bạn đã hỏi! Tôi là AI Financial Advisor, tôi có thể giúp bạn về:

• Phân bổ ngân sách
• Chiến lược trả nợ
• Mục tiêu tài chính
• Quỹ khẩn cấp
• Quyết định chi tiêu

Bạn muốn tôi giúp gì?`
}
