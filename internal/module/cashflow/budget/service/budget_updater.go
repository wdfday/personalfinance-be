package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateBudgetForUser updates a budget with ownership verification
func (s *budgetService) UpdateBudgetForUser(ctx context.Context, budget *domain.Budget, userID uuid.UUID) error {
	// Verify ownership first
	existingBudget, err := s.repo.FindByIDAndUserID(ctx, budget.ID, userID)
	if err != nil {
		return err
	}

	// Ensure we're not changing the owner
	budget.UserID = existingBudget.UserID

	// Validate and update
	if err := s.validateBudget(budget); err != nil {
		return err
	}

	// Recalculate fields before saving
	budget.UpdateCalculatedFields()

	s.logger.Info("Updating budget for user",
		zap.String("budget_id", budget.ID.String()),
		zap.String("user_id", userID.String()),
		zap.Float64("amount", budget.Amount),
	)

	return s.repo.Update(ctx, budget)
}

// CheckBudgetAlerts checks if any budget alerts should be triggered
func (s *budgetService) CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error) {
	s.logger.Debug("Checking budget alerts", zap.String("budget_id", budgetID.String()))

	budget, err := s.repo.FindByID(ctx, budgetID)
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
func (s *budgetService) MarkExpiredBudgets(ctx context.Context) error {
	s.logger.Info("Marking expired budgets")

	expiredBudgets, err := s.repo.FindExpiredBudgets(ctx)
	if err != nil {
		return err
	}

	for _, budget := range expiredBudgets {
		budget.Status = domain.BudgetStatusExpired
		if err := s.repo.Update(ctx, &budget); err != nil {
			s.logger.Error("Failed to mark budget as expired",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Marked expired budgets", zap.Int("count", len(expiredBudgets)))
	return nil
}
