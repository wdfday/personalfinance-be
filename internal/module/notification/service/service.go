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

// SecurityLogger defines security event logging operations
type SecurityLogger interface {
	// LogEvent logs a security event
	LogEvent(ctx context.Context, event domain.SecurityEvent)

	// Convenience methods for cashflow security events
	LogRegistration(ctx context.Context, userID, email, ipAddress string)
	LogLoginSuccess(ctx context.Context, userID, email, ipAddress string)
	LogLogout(ctx context.Context, userID, email, ipAddress string)
	LogLoginFailed(ctx context.Context, email, ipAddress, reason string)
	LogAccountLocked(ctx context.Context, userID, email, ipAddress string, lockedUntil time.Time)
	LogPasswordChanged(ctx context.Context, userID, email string)
	LogPasswordResetRequest(ctx context.Context, email, ipAddress string)
	LogPasswordReset(ctx context.Context, userID, email string)
	LogEmailVerified(ctx context.Context, userID, email string)
	LogGoogleOAuthLogin(ctx context.Context, userID, email, ipAddress string, isNewUser bool)
	LogTokenRefreshed(ctx context.Context, userID, email string)
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

// AlertRuleService defines operations for managing and evaluating alert rules
type AlertRuleService interface {
	// CreateRule creates a new alert rule
	CreateRule(ctx context.Context, rule *domain.AlertRule) error

	// GetRule retrieves an alert rule by ID
	GetRule(ctx context.Context, ruleID string) (*domain.AlertRule, error)

	// UpdateRule updates an existing alert rule
	UpdateRule(ctx context.Context, rule *domain.AlertRule) error

	// DeleteRule deletes an alert rule
	DeleteRule(ctx context.Context, ruleID string) error

	// ListUserRules lists all alert rules for a user
	ListUserRules(ctx context.Context, userID string, limit, offset int) ([]domain.AlertRule, error)

	// ListEnabledRules lists all enabled rules for a user
	ListEnabledRules(ctx context.Context, userID string) ([]domain.AlertRule, error)

	// EvaluateRule evaluates a rule and triggers notification if condition is met
	EvaluateRule(ctx context.Context, rule *domain.AlertRule) (bool, error)

	// EvaluateAllUserRules evaluates all enabled rules for a user
	EvaluateAllUserRules(ctx context.Context, userID string) error
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

// NotificationAnalyticsService defines operations for tracking notification analytics
type NotificationAnalyticsService interface {
	// TrackNotification creates an analytics record for a new notification
	TrackNotification(ctx context.Context, notification *domain.Notification) error

	// MarkDelivered marks a notification as delivered
	MarkDelivered(ctx context.Context, notificationID string) error

	// MarkRead marks a notification as read
	MarkRead(ctx context.Context, notificationID string) error

	// MarkClicked marks a notification as clicked
	MarkClicked(ctx context.Context, notificationID string) error

	// MarkFailed marks a notification as failed
	MarkFailed(ctx context.Context, notificationID string, reason string) error

	// TrackEmailOpen tracks email open via tracking pixel
	TrackEmailOpen(ctx context.Context, notificationID string) error

	// GetUserAnalytics retrieves analytics for a user
	GetUserAnalytics(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error)

	// GetAnalyticsByType retrieves analytics grouped by type
	GetAnalyticsByType(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error)

	// GetOverviewStats retrieves overall statistics
	GetOverviewStats(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error)

	// GetFailedNotifications retrieves failed notifications
	GetFailedNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error)
}
