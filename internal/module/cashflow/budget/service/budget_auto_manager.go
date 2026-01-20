package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetAutoManager handles automated budget management tasks
type BudgetAutoManager struct {
	service *budgetService
}

// NewBudgetAutoManager creates a new budget auto manager
func NewBudgetAutoManager(service *budgetService) *BudgetAutoManager {
	return &BudgetAutoManager{service: service}
}

// AutoRolloverExpiredBudgets automatically rolls over budgets that have ended and allow rollover
func (m *BudgetAutoManager) AutoRolloverExpiredBudgets(ctx context.Context, userID uuid.UUID) (int, error) {
	m.service.logger.Info("Auto-rollover: Processing budget rollovers",
		zap.String("user_id", userID.String()),
	)

	calculator := NewBudgetCalculator(m.service)
	return m.rolloverBudgetsInternal(ctx, userID, calculator)
}

// AutoMarkExpiredBudgets marks budgets as expired if they have passed their end date
func (m *BudgetAutoManager) AutoMarkExpiredBudgets(ctx context.Context) (int, error) {
	m.service.logger.Info("Auto-marking expired budgets")

	// Find all expired budgets
	expiredBudgets, err := m.service.repo.FindExpiredBudgets(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired budgets: %w", err)
	}

	markedCount := 0
	now := time.Now()

	for _, budget := range expiredBudgets {
		// Only mark as expired if status is not already expired
		if budget.Status != domain.BudgetStatusExpired {
			budget.Status = domain.BudgetStatusExpired
			if err := m.service.repo.Update(ctx, &budget); err != nil {
				m.service.logger.Error("Failed to mark budget as expired",
					zap.String("budget_id", budget.ID.String()),
					zap.Error(err),
				)
				continue
			}
			markedCount++
			m.service.logger.Debug("Marked budget as expired",
				zap.String("budget_id", budget.ID.String()),
				zap.String("name", budget.Name),
				zap.Time("end_date", *budget.EndDate),
				zap.Time("current_time", now),
			)
		}
	}

	m.service.logger.Info("Auto-marking expired budgets completed",
		zap.Int("marked_count", markedCount),
		zap.Int("total_expired", len(expiredBudgets)),
	)

	return markedCount, nil
}

// AutoCreateBudgetsFromConstraints creates budgets based on active budget constraints
// This is useful for automating budget creation at the start of a new period
func (m *BudgetAutoManager) AutoCreateBudgetsFromConstraints(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod, startDate time.Time) (int, error) {
	m.service.logger.Info("Auto-creating budgets from constraints",
		zap.String("user_id", userID.String()),
		zap.String("period", string(period)),
		zap.Time("start_date", startDate),
	)

	// Note: This requires budget_profile module integration
	// For now, we'll return a placeholder implementation
	// TODO: Integrate with budget_profile module to get active constraints

	m.service.logger.Warn("AutoCreateBudgetsFromConstraints is not yet fully implemented - requires budget_profile integration")
	return 0, nil
}

// rolloverBudgetsInternal is the internal implementation of budget rollover
func (m *BudgetAutoManager) rolloverBudgetsInternal(ctx context.Context, userID uuid.UUID, calculator *BudgetCalculator) (int, error) {
	// Find budgets that allow rollover and have ended
	budgets, err := m.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to find budgets: %w", err)
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

		// Only rollover if there's something to roll over
		if rolloverAmount <= 0 {
			m.service.logger.Debug("Skipping rollover: no remaining amount",
				zap.String("budget_id", budget.ID.String()),
				zap.Float64("remaining", budget.RemainingAmount),
			)
			continue
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
		case domain.BudgetPeriodDaily:
			endDate := newBudget.StartDate
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodWeekly:
			endDate := newBudget.StartDate.AddDate(0, 0, 6)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodMonthly:
			endDate := newBudget.StartDate.AddDate(0, 1, -1)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodQuarterly:
			endDate := newBudget.StartDate.AddDate(0, 3, -1)
			newBudget.EndDate = &endDate
		case domain.BudgetPeriodYearly:
			endDate := newBudget.StartDate.AddDate(1, 0, -1)
			newBudget.EndDate = &endDate
		}

		// Add rollover to budget amount
		newBudget.Amount += rolloverAmount

		creator := NewBudgetCreator(m.service)
		if err := creator.CreateBudget(ctx, newBudget); err != nil {
			m.service.logger.Error("Failed to rollover budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
			continue
		}

		rolloverCount++
		m.service.logger.Info("Successfully rolled over budget",
			zap.String("old_budget_id", budget.ID.String()),
			zap.String("new_budget_id", newBudget.ID.String()),
			zap.Float64("rollover_amount", rolloverAmount),
			zap.Float64("new_total_amount", newBudget.Amount),
		)
	}

	m.service.logger.Info("Finished budget rollovers",
		zap.Int("rollover_count", rolloverCount),
	)

	return rolloverCount, nil
}
