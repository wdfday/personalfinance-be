package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/shared"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationAnalyticsRepository struct {
	db *gorm.DB
}

// NewNotificationAnalyticsRepository creates a new notification analytics repository
func NewNotificationAnalyticsRepository(db *gorm.DB) NotificationAnalyticsRepository {
	return &notificationAnalyticsRepository{db: db}
}

func (r *notificationAnalyticsRepository) Create(ctx context.Context, analytics *domain.NotificationAnalytics) error {
	if analytics.ID == uuid.Nil {
		analytics.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(analytics).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationAnalyticsRepository) GetByNotificationID(ctx context.Context, notificationID string) (*domain.NotificationAnalytics, error) {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	var analytics domain.NotificationAnalytics
	if err := r.db.WithContext(ctx).
		Where("notification_id = ?", notifID).
		First(&analytics).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &analytics, nil
}

func (r *notificationAnalyticsRepository) Update(ctx context.Context, analytics *domain.NotificationAnalytics) error {
	result := r.db.WithContext(ctx).Save(analytics)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationAnalyticsRepository) MarkDelivered(ctx context.Context, notificationID string) error {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("notification_id = ?", notifID).
		Updates(map[string]interface{}{
			"delivered_at": now,
			"status":       "delivered",
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationAnalyticsRepository) MarkRead(ctx context.Context, notificationID string) error {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("notification_id = ?", notifID).
		Updates(map[string]interface{}{
			"read_at": now,
			"status":  "read",
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationAnalyticsRepository) MarkClicked(ctx context.Context, notificationID string) error {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("notification_id = ?", notifID).
		Updates(map[string]interface{}{
			"clicked_at":  now,
			"status":      "clicked",
			"click_count": gorm.Expr("click_count + 1"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationAnalyticsRepository) MarkFailed(ctx context.Context, notificationID string, reason string) error {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("notification_id = ?", notifID).
		Updates(map[string]interface{}{
			"failed_at":      now,
			"status":         "failed",
			"failure_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationAnalyticsRepository) IncrementOpenCount(ctx context.Context, notificationID string) error {
	notifID, err := uuid.Parse(notificationID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "notification_id").WithDetails("reason", "invalid UUID")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("notification_id = ?", notifID).
		Update("open_count", gorm.Expr("open_count + 1"))

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *notificationAnalyticsRepository) GetAnalyticsByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var analytics []domain.NotificationAnalytics
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		Order("queued_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&analytics).Error; err != nil {
		return nil, err
	}
	return analytics, nil
}

func (r *notificationAnalyticsRepository) GetAnalyticsByType(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	query := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("user_id = ?", userUUID)

	if startDate != nil {
		query = query.Where("queued_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("queued_at <= ?", endDate)
	}

	// Aggregate by type
	type TypeStats struct {
		Type         string  `json:"type"`
		Total        int64   `json:"total"`
		Sent         int64   `json:"sent"`
		Delivered    int64   `json:"delivered"`
		Read         int64   `json:"read"`
		Clicked      int64   `json:"clicked"`
		Failed       int64   `json:"failed"`
		DeliveryRate float64 `json:"delivery_rate"`
		ReadRate     float64 `json:"read_rate"`
		ClickRate    float64 `json:"click_rate"`
	}

	var typeStats []TypeStats
	if err := query.
		Select(`
			type,
			COUNT(*) as total,
			COUNT(CASE WHEN sent_at IS NOT NULL THEN 1 END) as sent,
			COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) as delivered,
			COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) as read,
			COUNT(CASE WHEN clicked_at IS NOT NULL THEN 1 END) as clicked,
			COUNT(CASE WHEN failed_at IS NOT NULL THEN 1 END) as failed,
			ROUND(CAST(COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as delivery_rate,
			ROUND(CAST(COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as read_rate,
			ROUND(CAST(COUNT(CASE WHEN clicked_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as click_rate
		`).
		Group("type").
		Find(&typeStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"by_type": typeStats,
	}, nil
}

func (r *notificationAnalyticsRepository) GetOverviewStats(ctx context.Context, userID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	query := r.db.WithContext(ctx).
		Model(&domain.NotificationAnalytics{}).
		Where("user_id = ?", userUUID)

	if startDate != nil {
		query = query.Where("queued_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("queued_at <= ?", endDate)
	}

	type OverviewStats struct {
		Total         int64   `json:"total"`
		Sent          int64   `json:"sent"`
		Delivered     int64   `json:"delivered"`
		Read          int64   `json:"read"`
		Clicked       int64   `json:"clicked"`
		Failed        int64   `json:"failed"`
		DeliveryRate  float64 `json:"delivery_rate"`
		ReadRate      float64 `json:"read_rate"`
		ClickRate     float64 `json:"click_rate"`
		FailureRate   float64 `json:"failure_rate"`
		AvgOpenCount  float64 `json:"avg_open_count"`
		AvgClickCount float64 `json:"avg_click_count"`
	}

	var stats OverviewStats
	if err := query.
		Select(`
			COUNT(*) as total,
			COUNT(CASE WHEN sent_at IS NOT NULL THEN 1 END) as sent,
			COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) as delivered,
			COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) as read,
			COUNT(CASE WHEN clicked_at IS NOT NULL THEN 1 END) as clicked,
			COUNT(CASE WHEN failed_at IS NOT NULL THEN 1 END) as failed,
			ROUND(CAST(COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as delivery_rate,
			ROUND(CAST(COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as read_rate,
			ROUND(CAST(COUNT(CASE WHEN clicked_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as click_rate,
			ROUND(CAST(COUNT(CASE WHEN failed_at IS NOT NULL THEN 1 END) AS FLOAT) / NULLIF(COUNT(*), 0) * 100, 2) as failure_rate,
			ROUND(AVG(open_count), 2) as avg_open_count,
			ROUND(AVG(click_count), 2) as avg_click_count
		`).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"overview": stats,
	}, nil
}

func (r *notificationAnalyticsRepository) GetFailedNotifications(ctx context.Context, userID string, limit, offset int) ([]domain.NotificationAnalytics, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var analytics []domain.NotificationAnalytics
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userUUID, "failed").
		Order("failed_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&analytics).Error; err != nil {
		return nil, err
	}
	return analytics, nil
}
