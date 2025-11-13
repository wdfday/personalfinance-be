package service

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"

	"go.uber.org/zap"
)

type notificationPreferenceService struct {
	repo   repository.NotificationPreferenceRepository
	logger *zap.Logger
}

// NewNotificationPreferenceService creates a new notification preference service
func NewNotificationPreferenceService(
	repo repository.NotificationPreferenceRepository,
	logger *zap.Logger,
) NotificationPreferenceService {
	return &notificationPreferenceService{
		repo:   repo,
		logger: logger,
	}
}

func (s *notificationPreferenceService) CreatePreference(ctx context.Context, pref *domain.NotificationPreference) error {
	s.logger.Info("Creating notification preference",
		zap.String("user_id", pref.UserID.String()),
		zap.String("type", string(pref.Type)),
	)
	return s.repo.Create(ctx, pref)
}

func (s *notificationPreferenceService) GetPreference(ctx context.Context, prefID string) (*domain.NotificationPreference, error) {
	return s.repo.GetByID(ctx, prefID)
}

func (s *notificationPreferenceService) GetUserPreferenceForType(ctx context.Context, userID string, notifType domain.NotificationType) (*domain.NotificationPreference, error) {
	return s.repo.GetByUserAndType(ctx, userID, notifType)
}

func (s *notificationPreferenceService) UpdatePreference(ctx context.Context, pref *domain.NotificationPreference) error {
	s.logger.Info("Updating notification preference",
		zap.String("id", pref.ID.String()),
		zap.String("user_id", pref.UserID.String()),
		zap.String("type", string(pref.Type)),
	)
	return s.repo.Update(ctx, pref)
}

func (s *notificationPreferenceService) DeletePreference(ctx context.Context, prefID string) error {
	s.logger.Info("Deleting notification preference", zap.String("id", prefID))
	return s.repo.Delete(ctx, prefID)
}

func (s *notificationPreferenceService) ListUserPreferences(ctx context.Context, userID string) ([]domain.NotificationPreference, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *notificationPreferenceService) UpsertPreference(ctx context.Context, pref *domain.NotificationPreference) error {
	s.logger.Info("Upserting notification preference",
		zap.String("user_id", pref.UserID.String()),
		zap.String("type", string(pref.Type)),
	)
	return s.repo.Upsert(ctx, pref)
}

func (s *notificationPreferenceService) GetEffectiveChannels(ctx context.Context, userID string, notifType domain.NotificationType) ([]domain.NotificationChannel, error) {
	// Try to get user's preference for this notification type
	pref, err := s.repo.GetByUserAndType(ctx, userID, notifType)
	if err != nil {
		// If no preference exists, return default channels
		s.logger.Debug("No preference found, using defaults",
			zap.String("user_id", userID),
			zap.String("type", string(notifType)),
		)
		return []domain.NotificationChannel{
			domain.ChannelEmail,
			domain.ChannelInApp,
		}, nil
	}

	// If preference exists but is disabled, return empty
	if !pref.Enabled {
		return []domain.NotificationChannel{}, nil
	}

	// Return preferred channels, or defaults if none specified
	if len(pref.PreferredChannels) > 0 {
		return pref.PreferredChannels, nil
	}

	return []domain.NotificationChannel{
		domain.ChannelEmail,
		domain.ChannelInApp,
	}, nil
}

func (s *notificationPreferenceService) ShouldSendNotification(ctx context.Context, userID string, notifType domain.NotificationType) (bool, error) {
	// Try to get user's preference for this notification type
	pref, err := s.repo.GetByUserAndType(ctx, userID, notifType)
	if err != nil {
		// If no preference exists, default to sending
		return true, nil
	}

	// Check if this notification type is enabled
	if !pref.Enabled {
		s.logger.Debug("Notification disabled by user preference",
			zap.String("user_id", userID),
			zap.String("type", string(notifType)),
		)
		return false, nil
	}

	// TODO: Add quiet hours check
	// TODO: Add minimum interval check

	return true, nil
}
