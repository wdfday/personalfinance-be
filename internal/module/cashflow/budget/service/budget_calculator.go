package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetCalculator handles budget calculation operations
type BudgetCalculator struct {
	service *budgetService
}

// NewBudgetCalculator creates a new budget calculator
func NewBudgetCalculator(service *budgetService) *BudgetCalculator {
	return &BudgetCalculator{service: service}
}

// RecalculateBudgetSpending recalculates the spent amount for a budget
func (c *BudgetCalculator) RecalculateBudgetSpending(ctx context.Context, budgetID uuid.UUID) error {
	c.service.logger.Debug("Recalculating budget spending", zap.String("budget_id", budgetID.String()))

	budget, err := c.service.repo.FindByID(ctx, budgetID)
	if err != nil {
		return err
	}

	// Calculate spent amount from transactions
	var spentAmount float64
	query := c.service.db.Table("transactions").
		Where("user_id = ?", budget.UserID).
		Where("direction = ?", "DEBIT").
		Where("booking_date >= ?", budget.StartDate)

	if budget.EndDate != nil {
		query = query.Where("booking_date <= ?", budget.EndDate)
	}

	if budget.CategoryID != nil {
		query = query.Where("classification->>'user_category_id' = ?", budget.CategoryID.String())
	}

	if budget.AccountID != nil {
		query = query.Where("account_id = ?", budget.AccountID)
	}

	if err := query.Select("COALESCE(SUM(amount), 0) / 100.0").Scan(&spentAmount).Error; err != nil {
		return fmt.Errorf("failed to calculate spent amount: %w", err)
	}

	// Update budget with new spent amount
	budget.SpentAmount = spentAmount
	budget.UpdateCalculatedFields()

	c.service.logger.Info("Recalculated budget spending",
		zap.String("budget_id", budgetID.String()),
		zap.Float64("spent_amount", spentAmount),
	)

	return c.service.repo.Update(ctx, budget)
}

// RecalculateAllBudgets recalculates spending for all active budgets
func (c *BudgetCalculator) RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error {
	c.service.logger.Info("Recalculating all budgets", zap.String("user_id", userID.String()))

	budgets, err := c.service.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	successCount := 0
	errorCount := 0

	for _, budget := range budgets {
		if err := c.RecalculateBudgetSpending(ctx, budget.ID); err != nil {
			c.service.logger.Error("Failed to recalculate budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	c.service.logger.Info("Finished recalculating budgets",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// RolloverBudgets processes budget rollovers for the new period
func (c *BudgetCalculator) RolloverBudgets(ctx context.Context, userID uuid.UUID) error {
	c.service.logger.Info("Processing budget rollovers", zap.String("user_id", userID.String()))

	// Find budgets that allow rollover and have ended
	budgets, err := c.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	rolloverCount := 0

	for _, budget := range budgets {
		// Skip if not allowing rollover or hasn't ended yet
		if !budget.AllowRollover || budget.EndDate == nil || budget.EndDate.After(now) {
			continue
		}

		// Calculate rollover amount
		rolloverAmount := budget.RemainingAmount
		if budget.CarryOverPercent != nil {
			rolloverAmount = budget.RemainingAmount * float64(*budget.CarryOverPercent) / 100
		}

		// Create new budget for next period
		newBudget := &domain.Budget{
			UserID:           budget.UserID,
			Name:             budget.Name,
			Description:      budget.Description,
			Amount:           budget.Amount,
			Currency:         budget.Currency,
			Period:           budget.Period,
			CategoryID:       budget.CategoryID,
			AccountID:        budget.AccountID,
			EnableAlerts:     budget.EnableAlerts,
			AlertThresholds:  budget.AlertThresholds,
			AllowRollover:    budget.AllowRollover,
			CarryOverPercent: budget.CarryOverPercent,
			RolloverAmount:   rolloverAmount,
		}

		// Set new period dates
		newBudget.StartDate = budget.EndDate.AddDate(0, 0, 1)
		switch budget.Period {
		case domain.BudgetPeriodMonthly:
			endDate := newBudget.StartDate.AddDate(0, 1, -1)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodWeekly:
			endDate := newBudget.StartDate.AddDate(0, 0, 6)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodYearly:
			endDate := newBudget.StartDate.AddDate(1, 0, -1)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodQuarterly:
			endDate := newBudget.StartDate.AddDate(0, 3, -1)
			newBudget.EndDate = &endDate
		}

		// Add rollover to budget amount
		newBudget.Amount += rolloverAmount

		creator := NewBudgetCreator(c.service)
		if err := creator.CreateBudget(ctx, newBudget); err != nil {
			c.service.logger.Error("Failed to rollover budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
		} else {
			rolloverCount++
		}
	}

	c.service.logger.Info("Finished budget rollovers",
		zap.Int("count", rolloverCount),
	)

	return nil
}
