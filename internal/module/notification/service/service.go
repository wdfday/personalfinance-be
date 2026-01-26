package service

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"time"
)

// EmailService defines email sending operations
type EmailService interface {
	// SendEmailFromTemplate sends email using HTML template file
	SendEmailFromTemplate(to, subject, templatePath string, data map[string]interface{}) error

	// SendVerificationEmail sends email verification email
	SendVerificationEmail(to, name, token string) error

	// SendPasswordResetEmail sends password reset email
	SendPasswordResetEmail(to, name, token string) error

	// SendWelcomeEmail sends welcome email to new user
	SendWelcomeEmail(to, name, email string) error

	// SendBudgetAlert sends budget alert notification
	SendBudgetAlert(to, name, category string, budgetAmount, spentAmount, threshold float64) error

	// SendGoalAchievedEmail sends goal achievement notification
	SendGoalAchievedEmail(to, name, goalName string, targetAmount, currentAmount float64, targetDate, achievedDate time.Time) error

	// SendMonthlySummary sends monthly financial summary
	SendMonthlySummary(to, name, month string, year int, summary map[string]interface{}) error

	// SendCustomEmail sends a custom email with template
	SendCustomEmail(to, subject, templateFileName string, data map[string]interface{}) error
}

// NotificationService defines notification orchestration operations
type NotificationService interface {
	// SendNotification sends a notification based on type
	SendNotification(notification domain.NotificationData) error

	// Convenience methods for cashflow notifications
	NotifyUserRegistration(userEmail, userName, username string)
	NotifyBudgetAlert(userEmail, userName, category string, budgetAmount, spentAmount, threshold float64)
	NotifyGoalAchieved(userEmail, userName, goalName string, targetAmount, currentAmount float64, targetDate, achievedDate time.Time)
	NotifyMonthlySummary(userEmail, userName, month string, year int, summary map[string]interface{})
}

// ScheduledReportService defines operations for generating scheduled financial reports
type ScheduledReportService interface {
	// GenerateDailyReport generates and sends daily summary report
	GenerateDailyReport(ctx context.Context, userID string) error

	// GenerateWeeklyReport generates and sends weekly summary report
	GenerateWeeklyReport(ctx context.Context, userID string) error

	// GenerateMonthlySummary generates and sends monthly summary report
	GenerateMonthlySummary(ctx context.Context, userID string) error

	// GenerateCustomReport generates a custom report for specific date range
	GenerateCustomReport(ctx context.Context, userID string, startDate, endDate time.Time) error
}

// SchedulerService defines operations for managing scheduled tasks
type SchedulerService interface {
	// Start starts the scheduler
	Start()

	// Stop stops the scheduler
	Stop()

	// IsRunning returns whether the scheduler is currently running
	IsRunning() bool
}

// NotificationPreferenceService defines operations for managing notification preferences
type NotificationPreferenceService interface {
	// CreatePreference creates a new notification preference
	CreatePreference(ctx context.Context, pref *domain.NotificationPreference) error

	// GetPreference retrieves a preference by ID
	GetPreference(ctx context.Context, prefID string) (*domain.NotificationPreference, error)

	// GetUserPreferenceForType retrieves user's preference for a specific notification type
	GetUserPreferenceForType(ctx context.Context, userID string, notifType domain.NotificationType) (*domain.NotificationPreference, error)

	// UpdatePreference updates an existing preference
	UpdatePreference(ctx context.Context, pref *domain.NotificationPreference) error

	// DeletePreference deletes a preference
	DeletePreference(ctx context.Context, prefID string) error

	// ListUserPreferences lists all preferences for a user
	ListUserPreferences(ctx context.Context, userID string) ([]domain.NotificationPreference, error)

	// UpsertPreference creates or updates a preference
	UpsertPreference(ctx context.Context, pref *domain.NotificationPreference) error

	// GetEffectiveChannels returns the effective channels for a user and notification type
	GetEffectiveChannels(ctx context.Context, userID string, notifType domain.NotificationType) ([]domain.NotificationChannel, error)

	// ShouldSendNotification checks if notification should be sent based on preferences
	ShouldSendNotification(ctx context.Context, userID string, notifType domain.NotificationType) (bool, error)
}
