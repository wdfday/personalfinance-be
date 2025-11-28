package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetDeleter handles budget deletion operations
type BudgetDeleter struct {
	service *budgetService
}

// NewBudgetDeleter creates a new budget deleter
func NewBudgetDeleter(service *budgetService) *BudgetDeleter {
	return &BudgetDeleter{service: service}
}

// DeleteBudget deletes a budget (soft delete)
func (d *BudgetDeleter) DeleteBudget(ctx context.Context, budgetID uuid.UUID) error {
	d.service.logger.Info("Deleting budget", zap.String("budget_id", budgetID.String()))

	// Verify budget exists
	_, err := d.service.repo.FindByID(ctx, budgetID)
	if err != nil {
		return err
	}

	return d.service.repo.Delete(ctx, budgetID)
}
