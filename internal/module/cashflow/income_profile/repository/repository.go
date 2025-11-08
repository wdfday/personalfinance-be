package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"

	"github.com/google/uuid"
)

// Repository defines income profile data access operations
type Repository interface {
	// Create creates a new income profile
	Create(ctx context.Context, ip *domain.IncomeProfile) error

	// GetByID retrieves an income profile by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IncomeProfile, error)

	// GetByUserAndPeriod retrieves an income profile by user and period (year, month)
	GetByUserAndPeriod(ctx context.Context, userID uuid.UUID, year, month int) (*domain.IncomeProfile, error)

	// GetByUser retrieves all income profiles for a user
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error)

	// GetByUserAndYear retrieves all income profiles for a user in a specific year
	GetByUserAndYear(ctx context.Context, userID uuid.UUID, year int) ([]*domain.IncomeProfile, error)

	// List retrieves income profiles with filters
	List(ctx context.Context, userID uuid.UUID, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error)

	// Update updates an existing income profile
	Update(ctx context.Context, ip *domain.IncomeProfile) error

	// Delete deletes an income profile
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if an income profile exists for user and period
	Exists(ctx context.Context, userID uuid.UUID, year, month int) (bool, error)
}
