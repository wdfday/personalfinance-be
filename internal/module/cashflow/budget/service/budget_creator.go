package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetCreator handles budget creation operations
type BudgetCreator struct {
	service *budgetService
}

// NewBudgetCreator creates a new budget creator
func NewBudgetCreator(service *budgetService) *BudgetCreator {
	return &BudgetCreator{service: service}
}

// CreateBudget creates a new budget
func (c *BudgetCreator) CreateBudget(ctx context.Context, budget *domain.Budget) error {
	// Validate budget
	if err := c.validateBudget(budget); err != nil {
		return err
	}

	// Set initial calculated fields
	budget.SpentAmount = 0
	budget.UpdateCalculatedFields()

	// Set default alert thresholds if not provided
	if len(budget.AlertThresholds) == 0 && budget.EnableAlerts {
		budget.AlertThresholds = []domain.AlertThreshold{
			domain.AlertAt75,
			domain.AlertAt90,
			domain.AlertAt100,
		}
	}

	c.service.logger.Info("Creating budget",
		zap.String("user_id", budget.UserID.String()),
		zap.String("name", budget.Name),
		zap.Float64("amount", budget.Amount),
	)

	return c.service.repo.Create(ctx, budget)
}

// validateBudget validates budget data
func (c *BudgetCreator) validateBudget(budget *domain.Budget) error {
	if budget.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	if budget.Amount <= 0 {
		return errors.New("budget amount must be greater than 0")
	}

	if budget.Name == "" {
		return errors.New("budget name is required")
	}

	if !budget.Period.IsValid() {
		return errors.New("invalid budget period")
	}

	if budget.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if budget.EndDate != nil && budget.EndDate.Before(budget.StartDate) {
		return errors.New("end date must be after start date")
	}

	// Validate that either category or account is specified (or both can be nil for general budget)
	// No validation needed - both can be nil for general budget

	return nil
}
