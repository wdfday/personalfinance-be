package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"personalfinancedss/internal/module/calendar/event/domain"
	"personalfinancedss/internal/shared"
)

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository builds a GORM-backed event repository.
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, event *domain.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *gormRepository) Update(ctx context.Context, event *domain.Event) error {
	result := r.db.WithContext(ctx).Model(event).Updates(event)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &event, nil
}

func (r *gormRepository) GetByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &event, nil
}

// ListByUserAndDateRange returns all events for a user within a date range
// Perfect for calendar month/week view
func (r *gormRepository) ListByUserAndDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*domain.Event, error) {
	var events []*domain.Event

	// Query events that:
	// 1. Start within the range, OR
	// 2. Started before but end within/after the range
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Where("(start_date BETWEEN ? AND ?) OR (start_date < ? AND (end_date IS NULL OR end_date >= ?))",
			from, to, from, from).
		Order("start_date ASC, created_at ASC").
		Find(&events).Error

	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *gormRepository) ListUpcomingByUser(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.Event, error) {
	var events []domain.Event
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("(start_date BETWEEN ? AND ?) OR (start_date < ? AND end_date IS NOT NULL AND end_date >= ?)", from, to, from, from).
		Order("start_date ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// CheckHolidayExists checks if a holiday with the same date and name already exists for the user
func (r *gormRepository) CheckHolidayExists(ctx context.Context, userID uuid.UUID, date time.Time, name string) (bool, error) {
	var count int64

	// Extract date only (ignore time component)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Model(&domain.Event{}).
		Where("user_id = ?", userID).
		Where("name = ?", name).
		Where("start_date BETWEEN ? AND ?", startOfDay, endOfDay).
		Where("deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.Event{}, "id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
