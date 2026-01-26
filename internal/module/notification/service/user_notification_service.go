package service

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"
)

// UserNotificationService defines operations for user-facing notification features
type UserNotificationService interface {
	// GetNotifications retrieves notifications for a user
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error)

	// GetUnreadCount gets the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID string) (int64, error)

	// MarkAsRead marks a notification as read (verifies ownership)
	MarkAsRead(ctx context.Context, userID, notificationID string) error

	// MarkAllAsRead marks all unread notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID string) error
}

type userNotificationService struct {
	repo repository.NotificationRepository
}

// NewUserNotificationService creates a new user notification service
func NewUserNotificationService(repo repository.NotificationRepository) UserNotificationService {
	return &userNotificationService{
		repo: repo,
	}
}

func (s *userNotificationService) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListByUserID(ctx, userID, limit, offset)
}

func (s *userNotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.repo.CountUnreadByUserID(ctx, userID)
}

func (s *userNotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	return s.repo.MarkAsRead(ctx, userID, notificationID)
}

func (s *userNotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsReadByUserID(ctx, userID)
}
