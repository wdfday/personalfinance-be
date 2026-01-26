package notification

import (
	"context"
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/handler"
	"personalfinancedss/internal/module/notification/repository"
	"personalfinancedss/internal/module/notification/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideEmailConfig creates email configuration from app config
func ProvideEmailConfig(cfg *config.Config) domain.EmailConfig {
	return domain.EmailConfig{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		FromEmail:    cfg.Email.FromEmail,
		FromName:     cfg.Email.FromName,
		FrontendURL:  cfg.Server.FrontendURL, // Use centralized config
	}
}

// ProvideEmailService creates an email service
func ProvideEmailService(emailConfig domain.EmailConfig, logger *zap.Logger) service.EmailService {
	return service.NewEmailService(emailConfig, logger)
}

// ProvideNotificationRepository creates a notification repository
func ProvideNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return repository.NewNotificationRepository(db)
}

// ProvideWebSocketHub creates a WebSocket hub
func ProvideWebSocketHub(lc fx.Lifecycle, logger *zap.Logger) *service.WebSocketHub {
	hub := service.NewWebSocketHub()

	// Start the hub in a goroutine when app starts
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go hub.Run()
			logger.Info("WebSocket hub started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("WebSocket hub stopped")
			return nil
		},
	})

	return hub
}

// ProvideNotificationService creates an enhanced notification service (supports both in-app and email)
func ProvideNotificationService(
	emailService service.EmailService,
	notifRepo repository.NotificationRepository,
	wsHub *service.WebSocketHub,
) service.NotificationService {
	return service.NewEnhancedNotificationService(emailService, notifRepo, wsHub)
}

// ProvideUserNotificationService creates a user notification service
func ProvideUserNotificationService(repo repository.NotificationRepository) service.UserNotificationService {
	return service.NewUserNotificationService(repo)
}

// ProvideNotificationPreferenceRepository creates a notification preference repository
func ProvideNotificationPreferenceRepository(db *gorm.DB) repository.NotificationPreferenceRepository {
	return repository.NewNotificationPreferenceRepository(db)
}

// ProvideNotificationPreferenceService creates a notification preference service
func ProvideNotificationPreferenceService(repo repository.NotificationPreferenceRepository, logger *zap.Logger) service.NotificationPreferenceService {
	return service.NewNotificationPreferenceService(repo, logger)
}

// ProvideScheduledReportService creates a scheduled report service
// Note: Requires transaction and user repositories from other modules
func ProvideScheduledReportService(
	transactionRepo interface{}, // Will be typed properly when wiring
	userRepo interface{}, // Will be typed properly when wiring
	emailService service.EmailService,
	logger *zap.Logger,
) service.ScheduledReportService {
	// Type assertion will be handled in actual wiring
	// For now, return nil to avoid compilation errors
	// This should be properly wired in internal/fx/application.go
	return nil
}

// Module provides notification module dependencies
var Module = fx.Module("notification",
	fx.Provide(
		// Config
		ProvideEmailConfig,

		// WebSocket Hub
		ProvideWebSocketHub,

		// Repositories
		ProvideNotificationRepository,
		ProvideNotificationPreferenceRepository,

		// Services
		ProvideEmailService,
		ProvideNotificationService,
		ProvideUserNotificationService,
		ProvideNotificationPreferenceService,
		// ProvideScheduledReportService, // TODO: Wire properly with transaction/user repos
		// ProvideSchedulerService,        // TODO: Enable after ScheduledReportService is wired

		// Handlers
		handler.NewHandler,
		handler.NewWebSocketHandler,
		handler.NewPreferenceHandler,
	),
)
