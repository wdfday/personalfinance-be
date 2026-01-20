package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/calendar/month/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormRepository implements Repository using GORM
type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// ===== Read Model Operations =====

func (r *gormRepository) CreateMonth(ctx context.Context, month *domain.Month) error {
	return r.db.WithContext(ctx).Create(month).Error
}

func (r *gormRepository) GetMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error) {
	var result domain.Month
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND month = ?", userID, month).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("month not found: %s", month)
		}
		return nil, err
	}

	return &result, nil
}

func (r *gormRepository) GetMonthByID(ctx context.Context, monthID uuid.UUID) (*domain.Month, error) {
	var result domain.Month
	err := r.db.WithContext(ctx).Where("id = ?", monthID).First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("month not found: %s", monthID)
		}
		return nil, err
	}

	return &result, nil
}

func (r *gormRepository) UpdateMonthState(ctx context.Context, monthID uuid.UUID, state *domain.MonthState, version int64) error {
	// Optimistic locking: only update if version matches
	result := r.db.WithContext(ctx).
		Model(&domain.Month{}).
		Where("id = ? AND version = ?", monthID, version).
		Updates(map[string]interface{}{
			"state":      state,
			"version":    version + 1,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic locking failed: version mismatch")
	}

	return nil
}

func (r *gormRepository) UpdateMonth(ctx context.Context, month *domain.Month) error {
	return r.db.WithContext(ctx).Save(month).Error
}

func (r *gormRepository) ListMonths(ctx context.Context, userID uuid.UUID) ([]*domain.Month, error) {
	var months []*domain.Month
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("month DESC").
		Find(&months).Error

	return months, err
}

func (r *gormRepository) GetPreviousMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error) {
	var result domain.Month
	// Find the month immediately before the given month
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND month < ?", userID, month).
		Order("month DESC").
		Limit(1).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No previous month exists (this is the first month)
		}
		return nil, err
	}

	return &result, nil
}
