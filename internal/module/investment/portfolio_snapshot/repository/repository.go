package repository

import (
	"context"
	"time"

	"personalfinancedss/internal/module/investment/portfolio_snapshot/domain"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/dto"

	"github.com/google/uuid"
)

// Repository defines portfolio snapshot data access operations
type Repository interface {
	// Create creates a new portfolio snapshot
	Create(ctx context.Context, snapshot *domain.PortfolioSnapshot) error

	// GetByID retrieves a portfolio snapshot by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PortfolioSnapshot, error)

	// GetByUserID retrieves a portfolio snapshot by ID and user ID
	GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.PortfolioSnapshot, error)

	// GetLatest retrieves the most recent snapshot for a user
	GetLatest(ctx context.Context, userID uuid.UUID) (*domain.PortfolioSnapshot, error)

	// GetByDateRange retrieves snapshots within a date range
	GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.PortfolioSnapshot, error)

	// List retrieves portfolio snapshots with filters and pagination
	List(ctx context.Context, userID uuid.UUID, query dto.ListSnapshotsQuery) ([]*domain.PortfolioSnapshot, int64, error)

	// GetByPeriod retrieves snapshots for a specific period (daily, weekly, monthly)
	GetByPeriod(ctx context.Context, userID uuid.UUID, period domain.SnapshotPeriod, limit int) ([]*domain.PortfolioSnapshot, error)

	// Update updates a portfolio snapshot
	Update(ctx context.Context, snapshot *domain.PortfolioSnapshot) error

	// Delete soft deletes a portfolio snapshot
	Delete(ctx context.Context, id uuid.UUID) error

	// GetPerformanceMetrics calculates performance metrics over a time period
	GetPerformanceMetrics(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*dto.PerformanceMetrics, error)
}
