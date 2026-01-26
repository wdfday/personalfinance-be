package service

import (
	"personalfinancedss/internal/module/cashflow/budget/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type budgetService struct {
	repo   repository.Repository
	db     *gorm.DB
	logger *zap.Logger
}

// NewService creates a new budget service
func NewService(repo repository.Repository, db *gorm.DB, logger *zap.Logger) Service {
	return &budgetService{
		repo:   repo,
		db:     db,
		logger: logger.Named("budget.service"),
	}
}
