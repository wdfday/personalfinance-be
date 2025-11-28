package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBudgetCalculator_RecalculateBudgetSpending_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:              budgetID,
		UserID:          uuid.New(),
		Name:            "Test Budget",
		Amount:          10000000,
		SpentAmount:     5000000,
		RemainingAmount: 5000000,
		PercentageSpent: 50.0,
		Status:          domain.BudgetStatusActive,
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)
	mockRepo.On("Update", ctx, budget).Return(nil)

	err := service.RecalculateBudgetSpending(ctx, budgetID)

	assert.NoError(t, err)
	assert.NotNil(t, budget.LastCalculatedAt)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RecalculateBudgetSpending_NotFound(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	expectedError := errors.New("budget not found")

	mockRepo.On("FindByID", ctx, budgetID).Return(nil, expectedError)

	err := service.RecalculateBudgetSpending(ctx, budgetID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RecalculateAllBudgets_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	budgets := []domain.Budget{
		{
			ID:              uuid.New(),
			UserID:          userID,
			Name:            "Budget 1",
			Amount:          10000000,
			SpentAmount:     5000000,
			RemainingAmount: 5000000,
			PercentageSpent: 50.0,
			Status:          domain.BudgetStatusActive,
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			Name:            "Budget 2",
			Amount:          5000000,
			SpentAmount:     4000000,
			RemainingAmount: 1000000,
			PercentageSpent: 80.0,
			Status:          domain.BudgetStatusWarning,
		},
	}

	mockRepo.On("FindActiveByUserID", ctx, userID).Return(budgets, nil)
	mockRepo.On("Update", ctx, &budgets[0]).Return(nil)
	mockRepo.On("Update", ctx, &budgets[1]).Return(nil)

	err := service.RecalculateAllBudgets(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RecalculateAllBudgets_NoBudgets(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	emptyBudgets := []domain.Budget{}

	mockRepo.On("FindActiveByUserID", ctx, userID).Return(emptyBudgets, nil)

	err := service.RecalculateAllBudgets(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RolloverBudgets_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	yesterday := time.Now().AddDate(0, 0, -1)
	carryOverPercent := 50

	budgets := []domain.Budget{
		{
			ID:               uuid.New(),
			UserID:           userID,
			Name:             "Budget with Rollover",
			Amount:           10000000,
			SpentAmount:      7000000,
			RemainingAmount:  3000000,
			Period:           domain.BudgetPeriodMonthly,
			StartDate:        yesterday.AddDate(0, -1, 0),
			EndDate:          &yesterday,
			AllowRollover:    true,
			CarryOverPercent: &carryOverPercent,
			Status:           domain.BudgetStatusActive,
		},
	}

	mockRepo.On("FindExpiredBudgets", ctx).Return(budgets, nil)
	mockRepo.On("Update", ctx, &budgets[0]).Return(nil)

	err := service.RolloverBudgets(ctx, userID)

	assert.NoError(t, err)
	// Rollover amount should be 50% of 3000000 = 1500000
	assert.Equal(t, 1500000.0, budgets[0].RolloverAmount)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RolloverBudgets_NoRollover(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	yesterday := time.Now().AddDate(0, 0, -1)

	budgets := []domain.Budget{
		{
			ID:              uuid.New(),
			UserID:          userID,
			Name:            "Budget without Rollover",
			Amount:          10000000,
			SpentAmount:     7000000,
			RemainingAmount: 3000000,
			Period:          domain.BudgetPeriodMonthly,
			StartDate:       yesterday.AddDate(0, -1, 0),
			EndDate:         &yesterday,
			AllowRollover:   false,
			Status:          domain.BudgetStatusActive,
		},
	}

	mockRepo.On("FindExpiredBudgets", ctx).Return(budgets, nil)

	err := service.RolloverBudgets(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, budgets[0].RolloverAmount)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCalculator_RolloverBudgets_ExceededBudget(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	yesterday := time.Now().AddDate(0, 0, -1)
	carryOverPercent := 50

	budgets := []domain.Budget{
		{
			ID:               uuid.New(),
			UserID:           userID,
			Name:             "Exceeded Budget",
			Amount:           10000000,
			SpentAmount:      11000000,
			RemainingAmount:  -1000000,
			Period:           domain.BudgetPeriodMonthly,
			StartDate:        yesterday.AddDate(0, -1, 0),
			EndDate:          &yesterday,
			AllowRollover:    true,
			CarryOverPercent: &carryOverPercent,
			Status:           domain.BudgetStatusExceeded,
		},
	}

	mockRepo.On("FindExpiredBudgets", ctx).Return(budgets, nil)

	err := service.RolloverBudgets(ctx, userID)

	assert.NoError(t, err)
	// No rollover for exceeded budgets
	assert.Equal(t, 0.0, budgets[0].RolloverAmount)
	mockRepo.AssertExpectations(t)
}
