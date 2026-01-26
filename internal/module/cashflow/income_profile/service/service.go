package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/module/cashflow/income_profile/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// IncomeProfileCreator defines income profile creation operations
type IncomeProfileCreator interface {
	CreateIncomeProfile(ctx context.Context, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error)
}

// IncomeProfileReader defines income profile read operations
type IncomeProfileReader interface {
	GetIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error)
	GetIncomeProfileWithHistory(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, []*domain.IncomeProfile, error)
	ListIncomeProfiles(ctx context.Context, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error)
	GetActiveIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error)
	GetArchivedIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error)
	GetRecurringIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error)
}

// IncomeProfileUpdater defines income profile update operations (creates new version)
type IncomeProfileUpdater interface {
	// UpdateIncomeProfile creates a NEW version and archives the old one
	UpdateIncomeProfile(ctx context.Context, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error)

	// UpdateDSSMetadata updates DSS analysis metadata
	UpdateDSSMetadata(ctx context.Context, userID string, profileID string, req dto.UpdateDSSMetadataRequest) (*domain.IncomeProfile, error)

	// ArchiveIncomeProfile manually archives an income profile
	ArchiveIncomeProfile(ctx context.Context, userID string, profileID string) error

	// EndIncomeProfile marks an income profile as ended
	EndIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error)

	// CheckAndArchiveEnded checks and marks ended income profiles (does not archive)
	CheckAndArchiveEnded(ctx context.Context, userID string) (int, error)
}

// IncomeProfileDeleter defines income profile delete operations
type IncomeProfileDeleter interface {
	DeleteIncomeProfile(ctx context.Context, userID string, profileID string) error
}

// Service is the composite interface for all income profile operations
type Service interface {
	IncomeProfileCreator
	IncomeProfileReader
	IncomeProfileUpdater
	IncomeProfileDeleter
}

// incomeProfileService implements all income profile use cases
type incomeProfileService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewService creates a new income profile service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &incomeProfileService{
		repo:   repo,
		logger: logger,
	}
}

// parseUserID parses user ID string to UUID
func parseUserID(userID string) (uuid.UUID, error) {
	return uuid.Parse(userID)
}

// parseProfileID parses profile ID string to UUID
func parseProfileID(profileID string) (uuid.UUID, error) {
	return uuid.Parse(profileID)
}
