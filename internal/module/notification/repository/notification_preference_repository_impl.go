package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationPreferenceRepository struct {
	db *gorm.DB
}

// NewNotificationPreferenceRepository creates a new notification preference repository
func NewNotificationPreferenceRepository(db *gorm.DB) NotificationPreferenceRepository {
	return &notificationPreferenceRepository{db: db}
}

func (r *notificationPreferenceRepository) Create(ctx context.Context, pref *domain.NotificationPreference) error {
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(pref).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationPreferenceRepository) GetByID(ctx context.Context, id string) (*domain.NotificationPreference, error) {
	prefID, err := uuid.Parse(id)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	var pref domain.NotificationPreference
	if err := r.db.WithContext(ctx).First(&pref, "id = ?", prefID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &pref, nil
}

func (r *notificationPreferenceRepository) GetByUserAndType(ctx context.Context, userID string, notifType domain.NotificationType) (*domain.NotificationPreference, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var pref domain.NotificationPreference
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userUUID, notifType).
		First(&pref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &pref, nil
}

func (r *notificationPreferenceRepository) Update(ctx context.Context, pref *domain.NotificationPreference) error {
	result := r.db.WithContext(ctx).Save(pref)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationPreferenceRepository) Delete(ctx context.Context, id string) error {
	prefID, err := uuid.Parse(id)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	result := r.db.WithContext(ctx).Delete(&domain.NotificationPreference{}, "id = ?", prefID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *notificationPreferenceRepository) ListByUserID(ctx context.Context, userID string) ([]domain.NotificationPreference, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var prefs []domain.NotificationPreference
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		Order("type ASC").
		Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

func (r *notificationPreferenceRepository) Upsert(ctx context.Context, pref *domain.NotificationPreference) error {
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}

	// Use GORM's Clauses with OnConflict for upsert behavior
	// On conflict on the unique index (user_id, type), update all fields
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"enabled",
				"preferred_channels",
				"min_interval",
				"quiet_hours_from",
				"quiet_hours_to",
				"timezone",
				"updated_at",
			}),
		}).
		Create(pref)

	return result.Error
}
