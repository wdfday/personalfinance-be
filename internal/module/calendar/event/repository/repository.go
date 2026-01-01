package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"personalfinancedss/internal/module/calendar/event/domain"
)

// Repository exposes CRUD-style operations for user events.
type Repository interface {
	Create(ctx context.Context, event *domain.Event) error
	Update(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	GetByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Event, error)

	// Calendar view queries
	ListByUserAndDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*domain.Event, error)
	ListUpcomingByUser(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.Event, error)

	// Holiday management
	CheckHolidayExists(ctx context.Context, userID uuid.UUID, date time.Time, name string) (bool, error)

	// Delete
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
