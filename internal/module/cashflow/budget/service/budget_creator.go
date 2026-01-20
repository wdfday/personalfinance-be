package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

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
	// Auto-calculate end_date if not provided (for non-custom periods)
	if budget.EndDate == nil && budget.Period != domain.BudgetPeriodCustom {
		if err := c.autoCalculateEndDate(budget); err != nil {
			return fmt.Errorf("failed to auto-calculate end date: %w", err)
		}
	}

	// Validate budget
	if err := c.validateBudget(budget); err != nil {
		return err
	}

	// Check for conflicts with existing budgets
	if err := c.checkConflicts(ctx, budget); err != nil {
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

	// Validate alert thresholds ordering
	if err := c.validateAlertThresholds(budget.AlertThresholds); err != nil {
		return err
	}

	c.service.logger.Info("Creating budget",
		zap.String("user_id", budget.UserID.String()),
		zap.String("name", budget.Name),
		zap.Float64("amount", budget.Amount),
		zap.String("period", string(budget.Period)),
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

// autoCalculateEndDate automatically calculates end_date based on period
func (c *BudgetCreator) autoCalculateEndDate(budget *domain.Budget) error {
	if budget.StartDate.IsZero() {
		return errors.New("cannot calculate end date without start date")
	}

	var endDate time.Time
	switch budget.Period {
	case domain.BudgetPeriodDaily:
		endDate = budget.StartDate.AddDate(0, 0, 1).Add(-time.Second) // End of same day
	case domain.BudgetPeriodWeekly:
		endDate = budget.StartDate.AddDate(0, 0, 7).Add(-time.Second) // End of 7th day
	case domain.BudgetPeriodMonthly:
		endDate = budget.StartDate.AddDate(0, 1, 0).Add(-time.Second) // End of month
	case domain.BudgetPeriodQuarterly:
		endDate = budget.StartDate.AddDate(0, 3, 0).Add(-time.Second) // End of quarter
	case domain.BudgetPeriodYearly:
		endDate = budget.StartDate.AddDate(1, 0, 0).Add(-time.Second) // End of year
	case domain.BudgetPeriodCustom:
		// For custom periods, end_date must be provided manually
		return nil
	default:
		return fmt.Errorf("unknown budget period: %s", budget.Period)
	}

	budget.EndDate = &endDate
	c.service.logger.Debug("Auto-calculated end date",
		zap.String("period", string(budget.Period)),
		zap.Time("start_date", budget.StartDate),
		zap.Time("end_date", endDate),
	)

	return nil
}

// checkConflicts checks if there is a conflicting budget with same period and category
func (c *BudgetCreator) checkConflicts(ctx context.Context, budget *domain.Budget) error {
	// Only check conflicts if both category and period are specified
	if budget.CategoryID == nil {
		return nil // General budgets don't conflict
	}

	// Get existing budgets for the same category
	existingBudgets, err := c.service.repo.FindByUserIDAndCategory(ctx, budget.UserID, *budget.CategoryID)
	if err != nil {
		c.service.logger.Warn("Failed to check budget conflicts", zap.Error(err))
		return nil // Don't fail creation if we can't check
	}

	// Check for overlapping date ranges
	for _, existing := range existingBudgets {
		// Skip if periods are different
		if existing.Period != budget.Period {
			continue
		}

		// Check date overlap
		if budget.EndDate != nil && existing.EndDate != nil {
			// Both have end dates - check for overlap
			if budget.StartDate.Before(*existing.EndDate) && existing.StartDate.Before(*budget.EndDate) {
				return fmt.Errorf("budget conflicts with existing budget '%s' for the same category and period", existing.Name)
			}
		} else {
			// At least one is recurring - check if there's any overlap
			if budget.EndDate == nil && existing.EndDate == nil {
				return fmt.Errorf("cannot create recurring budget: another recurring budget '%s' already exists for this category", existing.Name)
			}
		}
	}

	return nil
}

// validateAlertThresholds validates that alert thresholds are in ascending order
func (c *BudgetCreator) validateAlertThresholds(thresholds []domain.AlertThreshold) error {
	if len(thresholds) <= 1 {
		return nil // No need to validate if only 0 or 1 threshold
	}

	// Convert thresholds to float64 and check ordering
	var prev float64 = 0
	for i, threshold := range thresholds {
		current := threshold.ToFloat64()
		if current <= prev {
			return fmt.Errorf("alert thresholds must be in ascending order (threshold at index %d: %.0f%% is not greater than previous %.0f%%)", i, current, prev)
		}
		prev = current
	}

	return nil
}
