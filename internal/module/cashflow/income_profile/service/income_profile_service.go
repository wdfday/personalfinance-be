package service

import (
	"go.uber.org/zap"
	"personalfinancedss/internal/module/cashflow/income_profile/repository"
)

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
