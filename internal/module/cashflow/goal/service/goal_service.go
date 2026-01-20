package service

import (
	accountservice "personalfinancedss/internal/module/cashflow/account/service"
	"personalfinancedss/internal/module/cashflow/goal/repository"

	"go.uber.org/zap"
)

type goalService struct {
	repo           repository.Repository
	accountService accountservice.Service
	logger         *zap.Logger
}

// NewService creates a new goal service
func NewService(
	repo repository.Repository,
	accountService accountservice.Service,
	logger *zap.Logger,
) Service {
	return &goalService{
		repo:           repo,
		accountService: accountService,
		logger:         logger,
	}
}
