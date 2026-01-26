package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"personalfinancedss/internal/module/chatbot/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// chatService implements domain.ChatService
type chatService struct {
	convRepo        domain.ConversationRepository
	msgRepo         domain.MessageRepository
	providerFactory domain.LLMProviderFactory
	logger          *zap.Logger
}

// NewChatService creates a new chat service
func NewChatService(
	convRepo domain.ConversationRepository,
	msgRepo domain.MessageRepository,
	providerFactory domain.LLMProviderFactory,
	logger *zap.Logger,
) domain.ChatService {
	return &chatService{
		convRepo:        convRepo,
		msgRepo:         msgRepo,
		providerFactory: providerFactory,
		logger:          logger,
	}
}

func (s *chatService) Chat(ctx context.Context, req domain.ChatServiceRequest) (*domain.ChatServiceResponse, error) {
	startTime := time.Now()

	// Get or create conversation
	conv, err := s.getOrCreateConversation(ctx, req)
	if err != nil {
		return nil, err
	}

	// Save user message
	userMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           domain.RoleUser,
		Content:        req.Message,
		CreatedAt:      time.Now(),
	}
	if err := s.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation history
	history, err := s.msgRepo.GetByConversationID(ctx, conv.ID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// Get provider
	llmProvider, err := s.providerFactory.GetProvider(req.Provider)
	if err != nil {
		llmProvider = s.providerFactory.GetDefaultProvider()
	}

	// Build chat request
	chatReq := domain.ChatRequest{
		Messages:     history,
		Tools:        domain.ToolDefinitions,
		SystemPrompt: s.getSystemPrompt(),
	}

	// Call LLM
	chatResp, err := llmProvider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	// Save assistant message
	assistantMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           domain.RoleAssistant,
		Content:        chatResp.Message.Content,
		ToolCalls:      chatResp.Message.ToolCalls,
		TokenCount:     chatResp.Usage.TotalTokens,
		LatencyMs:      time.Since(startTime).Milliseconds(),
		Model:          llmProvider.Name(),
		FinishReason:   chatResp.FinishReason,
		CreatedAt:      time.Now(),
	}
	if err := s.msgRepo.Create(ctx, assistantMsg); err != nil {
		s.logger.Error("failed to save assistant message", zap.Error(err))
	}

	// Update conversation
	conv.UpdatedAt = time.Now()
	if err := s.convRepo.Update(ctx, conv); err != nil {
		s.logger.Error("failed to update conversation", zap.Error(err))
	}

	return &domain.ChatServiceResponse{
		ConversationID: conv.ID.String(),
		Message:        chatResp.Message.Content,
		Usage: &domain.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
		ToolsUsed: s.extractToolNames(chatResp.Message.ToolCalls),
	}, nil
}

func (s *chatService) ChatStream(ctx context.Context, req domain.ChatServiceRequest) (<-chan domain.StreamEvent, error) {
	// Get or create conversation
	conv, err := s.getOrCreateConversation(ctx, req)
	if err != nil {
		return nil, err
	}

	// Save user message
	userMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		Role:           domain.RoleUser,
		Content:        req.Message,
		CreatedAt:      time.Now(),
	}
	if err := s.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get history
	history, err := s.msgRepo.GetByConversationID(ctx, conv.ID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// Get provider
	llmProvider, err := s.providerFactory.GetProvider(req.Provider)
	if err != nil {
		llmProvider = s.providerFactory.GetDefaultProvider()
	}

	// Build request
	chatReq := domain.ChatRequest{
		Messages:     history,
		Tools:        domain.ToolDefinitions,
		SystemPrompt: s.getSystemPrompt(),
	}

	// Start streaming
	eventCh, err := llmProvider.ChatStream(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	// Wrap to save message at the end
	outputCh := make(chan domain.StreamEvent)
	go func() {
		defer close(outputCh)
		var contentBuilder strings.Builder
		var usage *domain.TokenUsage

		for event := range eventCh {
			outputCh <- event

			if event.Type == domain.EventDelta {
				contentBuilder.WriteString(event.Delta)
			}
			if event.Type == domain.EventEnd && event.Usage != nil {
				usage = event.Usage
			}
		}

		// Save assistant message
		assistantMsg := &domain.Message{
			ID:             uuid.New(),
			ConversationID: conv.ID,
			Role:           domain.RoleAssistant,
			Content:        contentBuilder.String(),
			Model:          llmProvider.Name(),
			CreatedAt:      time.Now(),
		}
		if usage != nil {
			assistantMsg.TokenCount = usage.TotalTokens
		}
		if err := s.msgRepo.Create(context.Background(), assistantMsg); err != nil {
			s.logger.Error("failed to save streamed message", zap.Error(err))
		}
	}()

	return outputCh, nil
}

func (s *chatService) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int, status string) (*domain.ConversationListResult, error) {
	convs, total, err := s.convRepo.GetByUserID(ctx, userID, limit, offset, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	hasMore := int64(offset+len(convs)) < total

	return &domain.ConversationListResult{
		Conversations: convs,
		Total:         int(total),
		HasMore:       hasMore,
	}, nil
}

func (s *chatService) GetConversation(ctx context.Context, userID, convID uuid.UUID, messageLimit int) (*domain.ConversationDetailResult, error) {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil || conv.UserID != userID {
		return nil, fmt.Errorf("conversation not found")
	}

	messages, err := s.msgRepo.GetByConversationID(ctx, convID, messageLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &domain.ConversationDetailResult{
		Conversation: *conv,
		Messages:     messages,
	}, nil
}

func (s *chatService) UpdateConversation(ctx context.Context, userID, convID uuid.UUID, title *string, status *string) (*domain.Conversation, error) {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil || conv.UserID != userID {
		return nil, fmt.Errorf("conversation not found")
	}

	if title != nil {
		conv.Title = *title
	}
	if status != nil {
		conv.Status = domain.ConversationStatus(*status)
	}

	if err := s.convRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return conv, nil
}

func (s *chatService) DeleteConversation(ctx context.Context, userID, convID uuid.UUID) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil || conv.UserID != userID {
		return fmt.Errorf("conversation not found")
	}

	return s.convRepo.Delete(ctx, convID)
}

// Helper methods

func (s *chatService) getOrCreateConversation(ctx context.Context, req domain.ChatServiceRequest) (*domain.Conversation, error) {
	if req.ConversationID != nil {
		conv, err := s.convRepo.GetByID(ctx, *req.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
		if conv == nil {
			return nil, fmt.Errorf("conversation not found")
		}
		if conv.UserID != req.UserID {
			return nil, fmt.Errorf("conversation not found")
		}
		return conv, nil
	}

	// Create new conversation
	conv := &domain.Conversation{
		ID:       uuid.New(),
		UserID:   req.UserID,
		Title:    s.generateTitle(req.Message),
		Status:   domain.ConversationStatusActive,
		Provider: req.Provider,
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	return conv, nil
}

func (s *chatService) generateTitle(message string) string {
	if len(message) > 50 {
		return message[:50] + "..."
	}
	return message
}

func (s *chatService) getSystemPrompt() string {
	return `Bạn là AI Financial Advisor - trợ lý tài chính thông minh.

Nhiệm vụ của bạn:
1. Tư vấn tài chính cá nhân bằng tiếng Việt
2. Phân tích và đề xuất phân bổ ngân sách
3. Đề xuất chiến lược trả nợ
4. Giúp đặt và theo dõi mục tiêu tài chính
5. Tính toán quỹ khẩn cấp
6. Hỗ trợ quyết định chi tiêu

Nguyên tắc:
- Luôn trả lời bằng tiếng Việt
- Đưa ra lời khuyên cụ thể, có thể hành động được
- Sử dụng các công cụ phân tích khi cần thiết
- Giải thích rõ ràng, dễ hiểu
- Tôn trọng quyết định của người dùng`
}

func (s *chatService) extractToolNames(toolCalls domain.ToolCalls) []string {
	if len(toolCalls) == 0 {
		return nil
	}
	names := make([]string, len(toolCalls))
	for i, tc := range toolCalls {
		names[i] = tc.Name
	}
	return names
}
