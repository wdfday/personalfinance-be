package service

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type notificationAnalyticsService struct {
	repo   repository.NotificationAnalyticsRepository
	logger *zap.Logger
}

// NewNotificationAnalyticsService creates a new notification analytics service
func NewNotificationAnalyticsService(
	repo repository.NotificationAnalyticsRepository,
	logger *zap.Logger,
) NotificationAnalyticsService {
	return &notificationAnalyticsService{
		repo:   repo,
		logger: logger,
	}
}

func (s *notificationAnalyticsService) TrackNotification(ctx context.Context, notification *domain.Notification) error {
	analytics := &domain.NotificationAnalytics{
		ID:             uuid.New(),
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Channel:        notification.Channel,
		QueuedAt:       time.Now(),
		Status:         "queued",
	}

	s.logger.Debug("Tracking notification",
		zap.String("notification_id", notification.ID.String()),
		zap.String("type", string(notification.Type)),
		zap.String("channel", string(notification.Channel)),
	)

	return s.repo.Create(ctx, analytics)
}

func (s *notificationAnalyticsService) MarkDelivered(ctx context.Context, notificationID string) error {
	s.logger.Debug("Marking notification as delivered", zap.String("notification_id", notificationID))
	return s.repo.MarkDelivered(ctx, notificationID)
}

func (s *notificationAnalyticsService) MarkRead(ctx context.Context, notificationID string) error {
	s.logger.Debug("Marking notification as read", zap.String("notification_id", notificationID))
	return s.repo.MarkRead(ctx, notificationID)
}

func (s *notificationAnalyticsService) MarkClicked(ctx context.Context, notificationID string) error {
	s.logger.Debug("Marking notification as clicked", zap.String("notification_id", notificationID))
	return s.repo.MarkClicked(ctx, notificationID)
}

func (s *notificationAnalyticsService) MarkFailed(ctx context.Context, notificationID string, reason string) error {
	s.logger.Warn("Marking notification as failed",
		zap.String("notification_id", notificationID),
		zap.String("reason", reason),
	)
	return s.repo.MarkFailed(ctx, notificationID, reason)
}

func (s *notificationAnalyticsService) TrackEmailOpen(ctx context.Context, notificationID string) error {
	s.logger.Debug("Tracking email open", zap.String("notification_id", notificationID))
	return s.repo.IncrementOpenCount(ctx, notificationID)
}

func (s *notificationAnalyticsService) GetUserAnalytics(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error) {
	return s.repo.GetAnalyticsByUserID(ctx, userID, limit, offset)
}

func (s *notificationAnalyticsService) GetAnalyticsByType(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.repo.GetAnalyticsByType(ctx, userID, startDate, endDate)
}

func (s *notificationAnalyticsService) GetOverviewStats(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return s.repo.GetOverviewStats(ctx, userID, startDate, endDate)
}

func (s *notificationAnalyticsService) GetFailedNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error) {
	return s.repo.GetFailedNotifications(ctx, userID, limit, offset)
}
