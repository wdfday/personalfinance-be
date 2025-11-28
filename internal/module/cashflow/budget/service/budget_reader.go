package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BudgetReader handles budget read operations
type BudgetReader struct {
	service *budgetService
}

// NewBudgetReader creates a new budget reader
func NewBudgetReader(service *budgetService) *BudgetReader {
	return &BudgetReader{service: service}
}

// GetBudgetByID retrieves a budget by ID
func (r *BudgetReader) GetBudgetByID(ctx context.Context, budgetID uuid.UUID) (*domain.Budget, error) {
	r.service.logger.Debug("Getting budget by ID", zap.String("budget_id", budgetID.String()))
	return r.service.repo.FindByID(ctx, budgetID)
}

// GetUserBudgets retrieves all budgets for a user
func (r *BudgetReader) GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	r.service.logger.Debug("Getting user budgets", zap.String("user_id", userID.String()))
	return r.service.repo.FindByUserID(ctx, userID)
}

// GetActiveBudgets retrieves all active budgets for a user
func (r *BudgetReader) GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	r.service.logger.Debug("Getting active budgets", zap.String("user_id", userID.String()))
	return r.service.repo.FindActiveByUserID(ctx, userID)
}

// GetBudgetsByCategory retrieves budgets for a specific category
func (r *BudgetReader) GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error) {
	r.service.logger.Debug("Getting budgets by category",
		zap.String("user_id", userID.String()),
		zap.String("category_id", categoryID.String()),
	)
	return r.service.repo.FindByUserIDAndCategory(ctx, userID, categoryID)
}

// GetBudgetsByAccount retrieves budgets for a specific account
func (r *BudgetReader) GetBudgetsByAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error) {
	r.service.logger.Debug("Getting budgets by account",
		zap.String("user_id", userID.String()),
		zap.String("account_id", accountID.String()),
	)
	return r.service.repo.FindByUserIDAndAccount(ctx, userID, accountID)
}

// GetBudgetsByPeriod gets budgets for a specific period
func (r *BudgetReader) GetBudgetsByPeriod(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod) ([]domain.Budget, error) {
	r.service.logger.Debug("Getting budgets by period",
		zap.String("user_id", userID.String()),
		zap.String("period", string(period)),
	)

	budgets, err := r.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var filtered []domain.Budget
	for _, budget := range budgets {
		if budget.Period == period {
			filtered = append(filtered, budget)
		}
	}

	return filtered, nil
}

// GetBudgetSummary gets a summary of budget performance
func (r *BudgetReader) GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*BudgetSummary, error) {
	r.service.logger.Debug("Getting budget summary", zap.String("user_id", userID.String()))

	budgets, err := r.service.repo.FindByUserID(ctx, userID)
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

		// Group by category if available
		if budget.CategoryID != nil {
			categoryKey := budget.CategoryID.String()
			if _, exists := summary.BudgetsByCategory[categoryKey]; !exists {
				summary.BudgetsByCategory[categoryKey] = &CategoryBudgetSum{
					CategoryID: *budget.CategoryID,
				}
			}
			catSum := summary.BudgetsByCategory[categoryKey]
			catSum.Amount += budget.Amount
			catSum.Spent += budget.SpentAmount
			catSum.Remaining += budget.RemainingAmount
			if catSum.Amount > 0 {
				catSum.Percentage = (catSum.Spent / catSum.Amount) * 100
			}
		}
	}

	if summary.TotalBudgets > 0 {
		summary.AveragePercentage = totalPercentage / float64(summary.TotalBudgets)
	}

	return summary, nil
}

// GetBudgetVsActual gets budget vs actual spending comparison
func (r *BudgetReader) GetBudgetVsActual(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod, startDate, endDate time.Time) ([]*BudgetVsActual, error) {
	r.service.logger.Debug("Getting budget vs actual",
		zap.String("user_id", userID.String()),
		zap.String("period", string(period)),
	)

	budgets, err := r.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []*BudgetVsActual

	for _, budget := range budgets {
		if budget.Period != period {
			continue
		}

		// Calculate actual spending from transactions
		var actualSpent float64
		query := r.service.db.Table("transactions").
			Where("user_id = ?", userID).
			Where("direction = ?", "DEBIT").
			Where("booking_date BETWEEN ? AND ?", startDate, endDate)

		if budget.CategoryID != nil {
			query = query.Where("classification->>'user_category_id' = ?", budget.CategoryID.String())
		}

		if err := query.Select("COALESCE(SUM(amount), 0) / 100.0").Scan(&actualSpent).Error; err != nil {
			r.service.logger.Error("Failed to calculate actual spending", zap.Error(err))
			continue
		}

		difference := budget.Amount - actualSpent
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (actualSpent / budget.Amount) * 100
		}

		status := "on_track"
		if actualSpent < budget.Amount*0.8 {
			status = "under"
		} else if actualSpent > budget.Amount {
			status = "over"
		}

		results = append(results, &BudgetVsActual{
			BudgetID:     budget.ID,
			CategoryID:   budget.CategoryID,
			BudgetAmount: budget.Amount,
			ActualSpent:  actualSpent,
			Difference:   difference,
			Percentage:   percentage,
			Status:       status,
		})
	}

	return results, nil
}

// GetBudgetProgress gets detailed progress for a budget
func (r *BudgetReader) GetBudgetProgress(ctx context.Context, budgetID uuid.UUID) (*BudgetProgress, error) {
	r.service.logger.Debug("Getting budget progress", zap.String("budget_id", budgetID.String()))

	budget, err := r.service.repo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	daysElapsed := int(now.Sub(budget.StartDate).Hours() / 24)
	daysRemaining := 0
	if budget.EndDate != nil {
		daysRemaining = int(budget.EndDate.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	}

	dailyAverage := 0.0
	if daysElapsed > 0 {
		dailyAverage = budget.SpentAmount / float64(daysElapsed)
	}

	projectedTotal := 0.0
	if budget.EndDate != nil {
		totalDays := int(budget.EndDate.Sub(budget.StartDate).Hours() / 24)
		if totalDays > 0 {
			projectedTotal = dailyAverage * float64(totalDays)
		}
	}

	onTrack := projectedTotal <= budget.Amount

	// Count transactions
	var transactionCount int64
	var lastTransactionDate *time.Time

	query := r.service.db.Table("transactions").
		Where("user_id = ?", budget.UserID).
		Where("direction = ?", "DEBIT").
		Where("booking_date >= ?", budget.StartDate)

	if budget.EndDate != nil {
		query = query.Where("booking_date <= ?", budget.EndDate)
	}

	if budget.CategoryID != nil {
		query = query.Where("classification->>'user_category_id' = ?", budget.CategoryID.String())
	}

	query.Count(&transactionCount)

	var lastTx time.Time
	if err := query.Select("MAX(booking_date)").Scan(&lastTx).Error; err == nil && !lastTx.IsZero() {
		lastTransactionDate = &lastTx
	}

	return &BudgetProgress{
		BudgetID:         budget.ID,
		Name:             budget.Name,
		Period:           budget.Period,
		StartDate:        budget.StartDate,
		EndDate:          budget.EndDate,
		Amount:           budget.Amount,
		SpentAmount:      budget.SpentAmount,
		RemainingAmount:  budget.RemainingAmount,
		PercentageSpent:  budget.PercentageSpent,
		Status:           budget.Status,
		DaysElapsed:      daysElapsed,
		DaysRemaining:    daysRemaining,
		DailyAverage:     dailyAverage,
		ProjectedTotal:   projectedTotal,
		OnTrack:          onTrack,
		TransactionCount: int(transactionCount),
		LastTransaction:  lastTransactionDate,
	}, nil
}

// GetBudgetAnalytics gets analytics for a budget
func (r *BudgetReader) GetBudgetAnalytics(ctx context.Context, budgetID uuid.UUID) (*BudgetAnalytics, error) {
	r.service.logger.Debug("Getting budget analytics", zap.String("budget_id", budgetID.String()))

	budget, err := r.service.repo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Calculate historical average (last 6 months)
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	var historicalAvg float64

	subQuery := r.service.db.Table("transactions").
		Select("DATE_TRUNC('month', booking_date) as month, SUM(amount) / 100.0 as monthly_sum").
		Where("user_id = ?", budget.UserID).
		Where("direction = ?", "DEBIT").
		Where("booking_date >= ?", sixMonthsAgo)

	if budget.CategoryID != nil {
		subQuery = subQuery.Where("classification->>'user_category_id' = ?", budget.CategoryID.String())
	}

	subQuery = subQuery.Group("month")

	r.service.db.Table("(?) as monthly", subQuery).
		Select("COALESCE(AVG(monthly_sum), 0)").
		Scan(&historicalAvg)

	// Determine trend
	trend := "stable"
	if budget.SpentAmount > historicalAvg*1.1 {
		trend = "increasing"
	} else if budget.SpentAmount < historicalAvg*0.9 {
		trend = "decreasing"
	}

	// Calculate volatility (simplified - would need more historical data)
	volatility := 0.15 // Placeholder

	// Compliance rate (how often budget is met - would need historical data)
	complianceRate := 0.85 // Placeholder

	// Recommended amount based on historical data
	recommendedAmount := historicalAvg * 1.1

	// Optimization score
	optimizationScore := 0.8
	if budget.SpentAmount <= budget.Amount {
		optimizationScore = 0.9
	}

	return &BudgetAnalytics{
		BudgetID:          budget.ID,
		HistoricalAverage: historicalAvg,
		Trend:             trend,
		Volatility:        volatility,
		ComplianceRate:    complianceRate,
		RecommendedAmount: recommendedAmount,
		OptimizationScore: optimizationScore,
	}, nil
}
