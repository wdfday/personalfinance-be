package chatbot

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/chatbot/domain"
	"personalfinancedss/internal/module/chatbot/handler"
	"personalfinancedss/internal/module/chatbot/provider"
	"personalfinancedss/internal/module/chatbot/repository"
	"personalfinancedss/internal/module/chatbot/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Module provides chatbot dependencies
var Module = fx.Module("chatbot",
	fx.Provide(
		NewProviderFactory,
		NewConversationRepository,
		NewMessageRepository,
		NewChatService,
		NewChatHandler,
	),
)

// NewProviderFactory creates a new provider factory with configured providers
func NewProviderFactory(cfg *config.Config, logger *zap.Logger) domain.LLMProviderFactory {
	factory := provider.NewProviderFactory()

	// Register Gemini provider if API key is configured
	if cfg.ExternalAPIs.GeminiAPIKey != "" {
		geminiProvider, err := provider.NewGenAIProvider(&provider.GenAIConfig{
			APIKey: cfg.ExternalAPIs.GeminiAPIKey,
			Model:  "gemini-2.0-flash",
		})
		if err != nil {
			logger.Warn("Failed to create Gemini provider", zap.Error(err))
		} else {
			factory.RegisterProvider(string(domain.ProviderGemini), geminiProvider)
			// Set Gemini as default if available
			if err := factory.SetDefaultProvider(string(domain.ProviderGemini)); err != nil {
				logger.Warn("Failed to set Gemini as default provider", zap.Error(err))
			} else {
				logger.Info("Gemini provider registered and set as default")
			}
		}
	} else {
		logger.Info("Gemini API key not configured, using mock provider")
	}

	return factory
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) domain.ConversationRepository {
	return repository.NewGormConversationRepository(db)
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return repository.NewGormMessageRepository(db)
}

// NewChatService creates a new chat service
func NewChatService(
	convRepo domain.ConversationRepository,
	msgRepo domain.MessageRepository,
	providerFactory domain.LLMProviderFactory,
	logger *zap.Logger,
) domain.ChatService {
	return service.NewChatService(convRepo, msgRepo, providerFactory, logger)
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService domain.ChatService, logger *zap.Logger) *handler.Handler {
	return handler.NewChatHandler(chatService, logger)
}
