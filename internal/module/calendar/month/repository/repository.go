package repository

import (
	"context"

	"personalfinancedss/internal/module/calendar/month/domain"

	"github.com/google/uuid"
)

// Repository defines the data access interface for the Month module
// It combines both the Read Model (monthly_budgets) and Event Store (month_event_logs)
type Repository interface {
	// ===== Read Model Operations (monthly_budgets) =====

	// CreateMonth creates a new month record in the read model
	CreateMonth(ctx context.Context, month *domain.Month) error

	// GetMonth retrieves a month by user ID and month string
	GetMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error)

	// GetMonthByID retrieves a month by its ID
	GetMonthByID(ctx context.Context, monthID uuid.UUID) (*domain.Month, error)

	// UpdateMonthState updates the JSONB state of a month (optimistic locking)
	UpdateMonthState(ctx context.Context, monthID uuid.UUID, state *domain.MonthState, version int64) error

	// UpdateMonth updates the entire month record
	UpdateMonth(ctx context.Context, month *domain.Month) error

	// ListMonths retrieves all months for a user
	ListMonths(ctx context.Context, userID uuid.UUID) ([]*domain.Month, error)

	// GetPreviousMonth retrieves the month before the given month
	GetPreviousMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error)
}
