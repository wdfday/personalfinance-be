package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"

	"github.com/google/uuid"
)

// Repository defines income profile data access operations with versioning support
type Repository interface {
	// Create creates a new income profile
	Create(ctx context.Context, ip *domain.IncomeProfile) error

	// GetByID retrieves an income profile by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IncomeProfile, error)

	// GetByUser retrieves all active income profiles for a user (not archived)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error)

	// GetActiveByUser retrieves all currently active income profiles for a user
	GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error)

	// GetArchivedByUser retrieves all archived income profiles for a user
	GetArchivedByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error)

	// GetByStatus retrieves income profiles by user and status
	GetByStatus(ctx context.Context, userID uuid.UUID, status domain.IncomeStatus) ([]*domain.IncomeProfile, error)

	// GetVersionHistory retrieves all versions of an income profile chain
	GetVersionHistory(ctx context.Context, profileID uuid.UUID) ([]*domain.IncomeProfile, error)

	// GetLatestVersion retrieves the latest version of an income profile chain
	GetLatestVersion(ctx context.Context, profileID uuid.UUID) (*domain.IncomeProfile, error)

	// GetBySource retrieves income profiles by user and source
	GetBySource(ctx context.Context, userID uuid.UUID, source string) ([]*domain.IncomeProfile, error)

	// GetRecurringByUser retrieves all recurring income profiles for a user
	GetRecurringByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error)

	// List retrieves income profiles with filters
	List(ctx context.Context, userID uuid.UUID, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error)

	// Update updates an existing income profile
	Update(ctx context.Context, ip *domain.IncomeProfile) error

	// Delete soft deletes an income profile
	Delete(ctx context.Context, id uuid.UUID) error

	// Archive archives an income profile
	Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error
}
