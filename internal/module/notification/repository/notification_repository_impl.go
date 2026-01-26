package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var notifications []domain.Notification
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) CountUnreadByUserID(ctx context.Context, userID string) (int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND status = ?", userUUID, "pending").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, userID, id string) error {
	notificationID, err := uuid.Parse(id)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userUUID).
		Update("status", "read")

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationRepository) MarkAllAsReadByUserID(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND status = ?", userUUID, "pending").
		Update("status", "read").Error; err != nil {
		return err
	}
	return nil
}
