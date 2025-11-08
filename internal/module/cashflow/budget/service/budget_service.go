package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"time"

	"github.com/google/uuid"
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
		logger: logger,
	}
}

func (s *budgetService) CreateBudget(ctx context.Context, budget *domain.Budget) error {
	// Validate budget
	if err := s.validateBudget(budget); err != nil {
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

	return s.repo.Create(ctx, budget)
}

func (s *budgetService) GetBudgetByID(ctx context.Context, budgetID uuid.UUID) (*domain.Budget, error) {
	return s.repo.FindByID(ctx, budgetID)
}

func (s *budgetService) GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *budgetService) GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	return s.repo.FindActiveByUserID(ctx, userID)
}

func (s *budgetService) GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error) {
	return s.repo.FindByUserIDAndCategory(ctx, userID, categoryID)
}

func (s *budgetService) GetBudgetsByAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error) {
	return s.repo.FindByUserIDAndAccount(ctx, userID, accountID)
}

func (s *budgetService) UpdateBudget(ctx context.Context, budget *domain.Budget) error {
	// Validate budget
	if err := s.validateBudget(budget); err != nil {
		return err
	}

	// Recalculate fields before saving
	budget.UpdateCalculatedFields()

	return s.repo.Update(ctx, budget)
}

func (s *budgetService) DeleteBudget(ctx context.Context, budgetID uuid.UUID) error {
	return s.repo.Delete(ctx, budgetID)
}

func (s *budgetService) RecalculateBudgetSpending(ctx context.Context, budgetID uuid.UUID) error {
	budget, err := s.repo.FindByID(ctx, budgetID)
	if err != nil {
		return err
	}

	// Calculate spent amount from transactions
	var spentAmount float64
	query := s.db.Table("transactions").
		Where("user_id = ?", budget.UserID).
		Where("transaction_type = ?", "expense").
		Where("transaction_date >= ?", budget.StartDate)

	if budget.EndDate != nil {
		query = query.Where("transaction_date <= ?", budget.EndDate)
	}

	if budget.CategoryID != nil {
		query = query.Where("category_id = ?", budget.CategoryID)
	}

	if budget.AccountID != nil {
		query = query.Where("account_id = ?", budget.AccountID)
	}

	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&spentAmount).Error; err != nil {
		return fmt.Errorf("failed to calculate spent amount: %w", err)
	}

	// Update budget with new spent amount
	budget.SpentAmount = spentAmount
	budget.UpdateCalculatedFields()

	return s.repo.Update(ctx, budget)
}

func (s *budgetService) RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error {
	budgets, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		if err := s.RecalculateBudgetSpending(ctx, budget.ID); err != nil {
			s.logger.Error("Failed to recalculate budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
			// Continue with other budgets even if one fails
		}
	}

	return nil
}

func (s *budgetService) CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error) {
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

func (s *budgetService) MarkExpiredBudgets(ctx context.Context) error {
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

	return nil
}

func (s *budgetService) RolloverBudgets(ctx context.Context, userID uuid.UUID) error {
	// Find budgets that allow rollover and have ended
	budgets, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()

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
		}

		// Add rollover to budget amount
		newBudget.Amount += rolloverAmount

		if err := s.CreateBudget(ctx, newBudget); err != nil {
			s.logger.Error("Failed to rollover budget",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *budgetService) GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*BudgetSummary, error) {
	budgets, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &BudgetSummary{
		BudgetsByCategory: make(map[string]*CategoryBudgetSum),
	}

	var totalPercentage float64

	for _, budget := range budgets {
		summary.TotalBudgets++
		summary.TotalAmount += budget.Amount
		summary.TotalSpent += budget.SpentAmount
		summary.TotalRemaining += budget.RemainingAmount
		totalPercentage += budget.PercentageSpent

		switch budget.Status {
		case domain.BudgetStatusActive:
			summary.ActiveBudgets++
		case domain.BudgetStatusExceeded:
			summary.ExceededBudgets++
		case domain.BudgetStatusWarning:
			summary.WarningBudgets++
		}
	}

	if summary.TotalBudgets > 0 {
		summary.AveragePercentage = totalPercentage / float64(summary.TotalBudgets)
	}

	return summary, nil
}

func (s *budgetService) validateBudget(budget *domain.Budget) error {
	if budget.Amount <= 0 {
		return errors.New("budget amount must be greater than 0")
	}

	if !budget.Period.IsValid() {
		return fmt.Errorf("invalid budget period: %s", budget.Period)
	}

	if budget.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if budget.EndDate != nil && budget.EndDate.Before(budget.StartDate) {
		return errors.New("end date must be after start date")
	}

	return nil
}
