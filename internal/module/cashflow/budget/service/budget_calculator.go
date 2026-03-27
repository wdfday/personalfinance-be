package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecalculateBudgetSpendingForUser recalculates spending with ownership verification
func (s *budgetService) RecalculateBudgetSpendingForUser(ctx context.Context, budgetID, userID uuid.UUID) error {
	s.logger.Debug("Recalculating budget spending for user",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)

	budget, err := s.repo.FindByIDAndUserID(ctx, budgetID, userID)
	if err != nil {
		return err
	}

	return s.recalculateSpending(ctx, budget)
}

// RecalculateAllBudgets recalculates spending for all active budgets
func (s *budgetService) RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Recalculating all budgets", zap.String("user_id", userID.String()))

	budgets, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	successCount := 0
	errorCount := 0

	for _, budget := range budgets {
		if err := s.recalculateSpending(ctx, &budget); err != nil {
			s.logger.Error("Failed to recalculate budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("Finished recalculating budgets",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// recalculateSpending performs the actual spending recalculation
func (s *budgetService) recalculateSpending(ctx context.Context, budget *domain.Budget) error {
	// Calculate spent amount from transactions that have link to this budget
	var spentAmount float64
	query := s.db.Table("transactions").
		Where("user_id = ?", budget.UserID).
		Where("direction = ?", "DEBIT")
	// 	Where("booking_date >= ?", budget.StartDate)

	// if budget.EndDate != nil {
	// 	query = query.Where("booking_date <= ?", budget.EndDate)
	// }

	// Only count transactions that have a link to this budget
	// Use PostgreSQL JSONB @> operator to check if links array contains the budget link
	linkJSON := fmt.Sprintf(`[{"type":"BUDGET","id":"%s"}]`, budget.ID.String())
	query = query.Where("links @> ?", linkJSON)

	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&spentAmount).Error; err != nil {
		return fmt.Errorf("failed to calculate spent amount: %w", err)
	}

	// Update budget with new spent amount
	budget.SpentAmount = spentAmount
	budget.UpdateCalculatedFields()

	s.logger.Info("Recalculated budget spending",
		zap.String("budget_id", budget.ID.String()),
		zap.Float64("spent_amount", spentAmount),
	)

	return s.repo.Update(ctx, budget)
}
