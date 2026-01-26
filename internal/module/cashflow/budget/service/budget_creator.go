package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/dto"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateBudget creates a new budget from request
func (s *budgetService) CreateBudget(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*domain.Budget, error) {
	// Handle optional period - default to one-time if not provided
	period := req.Period
	if period == nil || *period == "" {
		oneTime := domain.BudgetPeriodOneTime
		period = &oneTime
	}

	// Create domain object from request
	budget := &domain.Budget{
		UserID:               userID,
		Name:                 req.Name,
		Description:          req.Description,
		Amount:               req.Amount,
		Currency:             req.Currency,
		Period:               *period,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		CategoryID:           req.CategoryID,
		ConstraintID:         req.ConstraintID,
		EnableAlerts:         req.EnableAlerts,
		AlertThresholds:      domain.AlertThresholdsJSON(req.AlertThresholds),
		AllowRollover:        req.AllowRollover,
		CarryOverPercent:     req.CarryOverPercent,
		AutoAdjust:           req.AutoAdjust,
		AutoAdjustPercentage: req.AutoAdjustPercentage,
		AutoAdjustBasedOn:    req.AutoAdjustBasedOn,
	}

	// Auto-calculate end_date if not provided
	if budget.EndDate == nil && budget.Period != domain.BudgetPeriodCustom && budget.Period != domain.BudgetPeriodOneTime {
		if err := s.autoCalculateEndDate(budget); err != nil {
			fmt.Printf("ERROR: failed to auto-calculate end date: %v\n", err)
			return nil, fmt.Errorf("failed to auto-calculate end date: %w", err)
		}
	}

	// Validate budget
	if err := s.validateBudget(budget); err != nil {
		fmt.Printf("ERROR: validate budget: %v\n", err)
		return nil, err
	}

	// Check for conflicts
	if err := s.checkConflicts(ctx, budget); err != nil {
		fmt.Printf("ERROR: check conflicts: %v\n", err)
		return nil, err
	}

	// Set initial calculated fields
	budget.SpentAmount = 0
	budget.UpdateCalculatedFields()

	// Set default alert thresholds if not provided
	if len(budget.AlertThresholds) == 0 && budget.EnableAlerts {
		budget.AlertThresholds = domain.AlertThresholdsJSON{
			domain.AlertAt75,
			domain.AlertAt90,
			domain.AlertAt100,
		}
	}

	// Validate alert thresholds ordering
	if err := s.validateAlertThresholds([]domain.AlertThreshold(budget.AlertThresholds)); err != nil {
		fmt.Printf("ERROR: validate alert thresholds: %v\n", err)
		return nil, err
	}

	s.logger.Info("Creating budget",
		zap.String("user_id", userID.String()),
		zap.String("name", budget.Name),
		zap.Float64("amount", budget.Amount),
		zap.String("period", string(budget.Period)),
	)

	if err := s.repo.Create(ctx, budget); err != nil {
		fmt.Printf("ERROR: failed to create budget: %v\n", err)
		return nil, err
	}

	return budget, nil
}

// CreateBudgetFromDomain creates a budget from an existing domain object (DSS, rollover, etc.)
func (s *budgetService) CreateBudgetFromDomain(ctx context.Context, budget *domain.Budget) error {
	if budget.EndDate == nil && budget.Period != domain.BudgetPeriodCustom && budget.Period != domain.BudgetPeriodOneTime {
		if err := s.autoCalculateEndDate(budget); err != nil {
			fmt.Printf("ERROR: failed to auto-calculate end date: %v\n", err)
			return fmt.Errorf("failed to auto-calculate end date: %w", err)
		}
	}
	if err := s.validateBudget(budget); err != nil {
		fmt.Printf("ERROR: validate budget: %v\n", err)
		return err
	}
	if err := s.checkConflicts(ctx, budget); err != nil {
		fmt.Printf("ERROR: check conflicts: %v\n", err)
		return err
	}
	budget.SpentAmount = 0
	budget.UpdateCalculatedFields()
	if len(budget.AlertThresholds) == 0 && budget.EnableAlerts {
		budget.AlertThresholds = domain.AlertThresholdsJSON{
			domain.AlertAt75,
			domain.AlertAt90,
			domain.AlertAt100,
		}
	}
	if err := s.validateAlertThresholds([]domain.AlertThreshold(budget.AlertThresholds)); err != nil {
		fmt.Printf("ERROR: validate alert thresholds: %v\n", err)
		return err
	}
	s.logger.Info("Creating budget from domain",
		zap.String("user_id", budget.UserID.String()),
		zap.String("name", budget.Name),
		zap.Float64("amount", budget.Amount),
		zap.String("period", string(budget.Period)),
	)
	if err := s.repo.Create(ctx, budget); err != nil {
		fmt.Printf("ERROR: failed to create budget from domain: %v\n", err)
		return err
	}
	return nil
}

// validateBudget validates budget data
func (s *budgetService) validateBudget(budget *domain.Budget) error {
	if budget.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	if budget.Amount <= 0 {
		return errors.New("budget amount must be greater than 0")
	}

	if budget.Name == "" {
		return errors.New("budget name is required")
	}

	// Period is required, but can be one-time
	if budget.Period == "" {
		budget.Period = domain.BudgetPeriodOneTime
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

	return nil
}

// autoCalculateEndDate automatically calculates end_date based on period
func (s *budgetService) autoCalculateEndDate(budget *domain.Budget) error {
	if budget.StartDate.IsZero() {
		return errors.New("cannot calculate end date without start date")
	}

	var endDate time.Time
	switch budget.Period {
	case domain.BudgetPeriodDaily:
		endDate = budget.StartDate.AddDate(0, 0, 1).Add(-time.Second)
	case domain.BudgetPeriodWeekly:
		endDate = budget.StartDate.AddDate(0, 0, 7).Add(-time.Second)
	case domain.BudgetPeriodMonthly:
		endDate = budget.StartDate.AddDate(0, 1, 0).Add(-time.Second)
	case domain.BudgetPeriodQuarterly:
		endDate = budget.StartDate.AddDate(0, 3, 0).Add(-time.Second)
	case domain.BudgetPeriodYearly:
		endDate = budget.StartDate.AddDate(1, 0, 0).Add(-time.Second)
	case domain.BudgetPeriodCustom:
		return nil
	default:
		return fmt.Errorf("unknown budget period: %s", budget.Period)
	}

	budget.EndDate = &endDate
	s.logger.Debug("Auto-calculated end date",
		zap.String("period", string(budget.Period)),
		zap.Time("start_date", budget.StartDate),
		zap.Time("end_date", endDate),
	)

	return nil
}

// checkConflicts checks if there is a conflicting budget with same period and category
func (s *budgetService) checkConflicts(ctx context.Context, budget *domain.Budget) error {
	// Only check conflicts if both category and period are specified
	if budget.CategoryID == nil {
		return nil // General budgets don't conflict
	}

	// Get existing budgets for the same category
	existingBudgets, err := s.repo.FindByUserIDAndCategory(ctx, budget.UserID, *budget.CategoryID)
	if err != nil {
		fmt.Printf("ERROR: Failed to check budget conflicts: %v\n", err)
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
func (s *budgetService) validateAlertThresholds(thresholds []domain.AlertThreshold) error {
	if len(thresholds) <= 1 {
		return nil
	}

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
