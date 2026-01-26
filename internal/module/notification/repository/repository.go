package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
)

// NotificationRepository defines data access methods for notifications
type NotificationRepository interface {
	// Create a new notification record
	Create(ctx context.Context, notification *domain.Notification) error

	// ListByUserID retrieves notifications for a specific user
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error)

	// CountUnreadByUserID counts unread notifications for a user
	CountUnreadByUserID(ctx context.Context, userID string) (int64, error)

	// MarkAsRead marks a single notification as read (verifies ownership)
	MarkAsRead(ctx context.Context, userID, id string) error

	// MarkAllAsReadByUserID marks all unread notifications as read for a user
	MarkAllAsReadByUserID(ctx context.Context, userID string) error
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
