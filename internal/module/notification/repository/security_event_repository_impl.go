package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type securityEventRepository struct {
	db *gorm.DB
}

// NewSecurityEventRepository creates a new security event repository
func NewSecurityEventRepository(db *gorm.DB) SecurityEventRepository {
	return &securityEventRepository{db: db}
}

func (r *securityEventRepository) Create(ctx context.Context, event *domain.SecurityEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return err
	}
	return nil
}

func (r *securityEventRepository) GetByID(ctx context.Context, id string) (*domain.SecurityEvent, error) {
	eventID, err := uuid.Parse(id)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	var event domain.SecurityEvent
	if err := r.db.WithContext(ctx).First(&event, "id = ?", eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &event, nil
}

func (r *securityEventRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.SecurityEvent, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var events []domain.SecurityEvent
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *securityEventRepository) ListByType(ctx context.Context, eventType domain.SecurityEventType, limit, offset int) ([]domain.SecurityEvent, error) {
	var events []domain.SecurityEvent
	query := r.db.WithContext(ctx).
		Where("type = ?", eventType).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *securityEventRepository) ListByEmail(ctx context.Context, email string, limit, offset int) ([]domain.SecurityEvent, error) {
	var events []domain.SecurityEvent
	query := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *securityEventRepository) ListFailedEvents(ctx context.Context, limit, offset int) ([]domain.SecurityEvent, error) {
	var events []domain.SecurityEvent
	query := r.db.WithContext(ctx).
		Where("success = ?", false).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
