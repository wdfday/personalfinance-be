package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"time"
)

// SecurityEventRepository defines data access methods for security events (audit logs)
type SecurityEventRepository interface {
	// Create a new security event
	Create(ctx context.Context, event *domain.SecurityEvent) error

	// GetByID retrieves a security event by ID
	GetByID(ctx context.Context, id string) (*domain.SecurityEvent, error)

	// ListByUserID retrieves security events for a specific user
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.SecurityEvent, error)

	// ListByType retrieves security events of a specific type
	ListByType(ctx context.Context, eventType domain.SecurityEventType, limit, offset int) ([]domain.SecurityEvent, error)

	// ListByEmail retrieves security events for a specific email
	ListByEmail(ctx context.Context, email string, limit, offset int) ([]domain.SecurityEvent, error)

	// ListFailedEvents retrieves failed security events
	ListFailedEvents(ctx context.Context, limit, offset int) ([]domain.SecurityEvent, error)
}

// NotificationRepository defines data access methods for notifications
type NotificationRepository interface {
	// Create a new notification record
	Create(ctx context.Context, notification *domain.Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id string) (*domain.Notification, error)

	// ListByUserID retrieves notifications for a specific user
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error)

	// CountUnreadByUserID counts unread notifications for a user
	CountUnreadByUserID(ctx context.Context, userID string) (int64, error)

	// Update notification status
	UpdateStatus(ctx context.Context, id string, status string) error

	// MarkAsSent marks notification as sent
	MarkAsSent(ctx context.Context, id string) error

	// MarkAsFailed marks notification as failed with error message
	MarkAsFailed(ctx context.Context, id string, errorMsg string) error

	// MarkAsRead marks a single notification as read
	MarkAsRead(ctx context.Context, id string) error

	// MarkAllAsReadByUserID marks all unread notifications as read for a user
	MarkAllAsReadByUserID(ctx context.Context, userID string) error
}

// AlertRuleRepository defines data access methods for alert rules
type AlertRuleRepository interface {
	// Create a new alert rule
	Create(ctx context.Context, rule *domain.AlertRule) error

	// GetByID retrieves an alert rule by ID
	GetByID(ctx context.Context, id string) (*domain.AlertRule, error)

	// Update an existing alert rule
	Update(ctx context.Context, rule *domain.AlertRule) error

	// Delete an alert rule (soft delete)
	Delete(ctx context.Context, id string) error

	// ListByUserID retrieves all alert rules for a specific user
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.AlertRule, error)

	// ListByUserAndType retrieves alert rules by user and type
	ListByUserAndType(ctx context.Context, userID string, ruleType domain.AlertRuleType, limit, offset int) ([]domain.AlertRule, error)

	// ListEnabled retrieves all enabled alert rules for a user
	ListEnabled(ctx context.Context, userID string) ([]domain.AlertRule, error)

	// ListScheduled retrieves all alert rules with schedules that need to be triggered
	ListScheduled(ctx context.Context) ([]domain.AlertRule, error)

	// UpdateLastTriggered updates the last and next trigger times
	UpdateLastTriggered(ctx context.Context, id string, lastTriggered, nextTrigger *time.Time) error
}

// NotificationPreferenceRepository defines data access methods for notification preferences
type NotificationPreferenceRepository interface {
	// Create a new notification preference
	Create(ctx context.Context, pref *domain.NotificationPreference) error

	// GetByID retrieves a notification preference by ID
	GetByID(ctx context.Context, id string) (*domain.NotificationPreference, error)

	// GetByUserAndType retrieves notification preference by user and notification type
	GetByUserAndType(ctx context.Context, userID string, notifType domain.NotificationType) (*domain.NotificationPreference, error)

	// Update an existing notification preference
	Update(ctx context.Context, pref *domain.NotificationPreference) error

	// Delete a notification preference
	Delete(ctx context.Context, id string) error

	// ListByUserID retrieves all notification preferences for a user
	ListByUserID(ctx context.Context, userID string) ([]domain.NotificationPreference, error)

	// Upsert creates or updates a notification preference
	Upsert(ctx context.Context, pref *domain.NotificationPreference) error
}

// NotificationAnalyticsRepository defines data access methods for notification analytics
type NotificationAnalyticsRepository interface {
	// Create a new analytics record
	Create(ctx context.Context, analytics *domain.NotificationAnalytics) error

	// GetByNotificationID retrieves analytics by notification ID
	GetByNotificationID(ctx context.Context, notificationID string) (*domain.NotificationAnalytics, error)

	// Update an existing analytics record
	Update(ctx context.Context, analytics *domain.NotificationAnalytics) error

	// MarkDelivered updates the delivered timestamp
	MarkDelivered(ctx context.Context, notificationID string) error

	// MarkRead updates the read timestamp
	MarkRead(ctx context.Context, notificationID string) error

	// MarkClicked updates the clicked timestamp and increments click count
	MarkClicked(ctx context.Context, notificationID string) error

	// MarkFailed marks the notification as failed with reason
	MarkFailed(ctx context.Context, notificationID string, reason string) error

	// IncrementOpenCount increments the open count (for email tracking pixels)
	IncrementOpenCount(ctx context.Context, notificationID string) error

	// GetAnalyticsByUserID retrieves analytics for a specific user
	GetAnalyticsByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error)

	// GetAnalyticsByType retrieves analytics by notification type with aggregation
	GetAnalyticsByType(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error)

	// GetOverviewStats retrieves overall analytics statistics
	GetOverviewStats(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error)

	// GetFailedNotifications retrieves list of failed notifications
	GetFailedNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error)
}
