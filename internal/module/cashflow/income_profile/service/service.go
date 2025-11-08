package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/module/cashflow/income_profile/repository"

	"go.uber.org/zap"
)

// IncomeProfileCreator defines income profile creation operations
type IncomeProfileCreator interface {
	CreateIncomeProfile(ctx context.Context, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error)
}

// IncomeProfileReader defines income profile read operations
type IncomeProfileReader interface {
	GetIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error)
	GetIncomeProfileByPeriod(ctx context.Context, userID string, year, month int) (*domain.IncomeProfile, error)
	ListIncomeProfiles(ctx context.Context, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error)
}

// IncomeProfileUpdater defines income profile update operations
type IncomeProfileUpdater interface {
	UpdateIncomeProfile(ctx context.Context, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error)
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
