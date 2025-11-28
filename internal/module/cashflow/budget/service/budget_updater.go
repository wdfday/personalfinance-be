package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetUpdater handles budget update operations
type BudgetUpdater struct {
	service *budgetService
}

// NewBudgetUpdater creates a new budget updater
func NewBudgetUpdater(service *budgetService) *BudgetUpdater {
	return &BudgetUpdater{service: service}
}

// UpdateBudget updates an existing budget
func (u *BudgetUpdater) UpdateBudget(ctx context.Context, budget *domain.Budget) error {
	// Validate budget
	creator := NewBudgetCreator(u.service)
	if err := creator.validateBudget(budget); err != nil {
		return err
	}

	// Recalculate fields before saving
	budget.UpdateCalculatedFields()

	u.service.logger.Info("Updating budget",
		zap.String("budget_id", budget.ID.String()),
		zap.Float64("amount", budget.Amount),
	)

	return u.service.repo.Update(ctx, budget)
}

// CheckBudgetAlerts checks if any budget alerts should be triggered
func (u *BudgetUpdater) CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error) {
	u.service.logger.Debug("Checking budget alerts", zap.String("budget_id", budgetID.String()))

	budget, err := u.service.repo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	if !budget.EnableAlerts {
		return nil, nil
	}

	var triggeredAlerts []domain.AlertThreshold

	for _, threshold := range budget.AlertThresholds {
		var thresholdValue float64
		switch threshold {
		case domain.AlertAt50:
			thresholdValue = 50
		case domain.AlertAt75:
			thresholdValue = 75
		case domain.AlertAt90:
			thresholdValue = 90
		case domain.AlertAt100:
			thresholdValue = 100
		}

		if budget.PercentageSpent >= thresholdValue {
			triggeredAlerts = append(triggeredAlerts, threshold)
		}
	}

	return triggeredAlerts, nil
}

// MarkExpiredBudgets marks expired budgets as expired
func (u *BudgetUpdater) MarkExpiredBudgets(ctx context.Context) error {
	u.service.logger.Info("Marking expired budgets")

	expiredBudgets, err := u.service.repo.FindExpiredBudgets(ctx)
	if err != nil {
		return err
	}

	for _, budget := range expiredBudgets {
		budget.Status = domain.BudgetStatusExpired
		if err := u.service.repo.Update(ctx, &budget); err != nil {
			u.service.logger.Error("Failed to mark budget as expired",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
		}
	}

	u.service.logger.Info("Marked expired budgets", zap.Int("count", len(expiredBudgets)))
	return nil
}
