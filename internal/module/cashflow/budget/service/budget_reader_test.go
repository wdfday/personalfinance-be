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

func TestBudgetReader_GetBudgetByID_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	expectedBudget := &domain.Budget{
		ID:     budgetID,
		UserID: uuid.New(),
		Name:   "Test Budget",
		Amount: 5000000,
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(expectedBudget, nil)

	result, err := service.GetBudgetByID(ctx, budgetID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBudget, result)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetByID_NotFound(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	expectedError := errors.New("budget not found")

	mockRepo.On("FindByID", ctx, budgetID).Return(nil, expectedError)

	result, err := service.GetBudgetByID(ctx, budgetID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetUserBudgets_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	expectedBudgets := []domain.Budget{
		{
			ID:     uuid.New(),
			UserID: userID,
			Name:   "Budget 1",
			Amount: 5000000,
		},
		{
			ID:     uuid.New(),
			UserID: userID,
			Name:   "Budget 2",
			Amount: 3000000,
		},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(expectedBudgets, nil)

	result, err := service.GetUserBudgets(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedBudgets, result)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetActiveBudgets_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	expectedBudgets := []domain.Budget{
		{
			ID:     uuid.New(),
			UserID: userID,
			Name:   "Active Budget",
			Amount: 5000000,
			Status: domain.BudgetStatusActive,
		},
	}

	mockRepo.On("FindActiveByUserID", ctx, userID).Return(expectedBudgets, nil)

	result, err := service.GetActiveBudgets(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.BudgetStatusActive, result[0].Status)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetsByCategory_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	categoryID := uuid.New()
	expectedBudgets := []domain.Budget{
		{
			ID:         uuid.New(),
			UserID:     userID,
			CategoryID: &categoryID,
			Name:       "Category Budget",
			Amount:     5000000,
		},
	}

	mockRepo.On("FindByUserIDAndCategory", ctx, userID, categoryID).Return(expectedBudgets, nil)

	result, err := service.GetBudgetsByCategory(ctx, userID, categoryID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, &categoryID, result[0].CategoryID)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetsByAccount_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	accountID := uuid.New()
	expectedBudgets := []domain.Budget{
		{
			ID:        uuid.New(),
			UserID:    userID,
			AccountID: &accountID,
			Name:      "Account Budget",
			Amount:    5000000,
		},
	}

	mockRepo.On("FindByUserIDAndAccount", ctx, userID, accountID).Return(expectedBudgets, nil)

	result, err := service.GetBudgetsByAccount(ctx, userID, accountID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, &accountID, result[0].AccountID)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetsByPeriod_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	expectedBudgets := []domain.Budget{
		{
			ID:     uuid.New(),
			UserID: userID,
			Name:   "Monthly Budget",
			Amount: 5000000,
			Period: domain.BudgetPeriodMonthly,
		},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(expectedBudgets, nil)

	result, err := service.GetBudgetsByPeriod(ctx, userID, domain.BudgetPeriodMonthly)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.BudgetPeriodMonthly, result[0].Period)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetSummary_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	categoryID1 := uuid.New()
	categoryID2 := uuid.New()

	budgets := []domain.Budget{
		{
			ID:              uuid.New(),
			UserID:          userID,
			CategoryID:      &categoryID1,
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
			CategoryID:      &categoryID2,
			Name:            "Budget 2",
			Amount:          5000000,
			SpentAmount:     4000000,
			RemainingAmount: 1000000,
			PercentageSpent: 80.0,
			Status:          domain.BudgetStatusWarning,
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			CategoryID:      &categoryID1,
			Name:            "Budget 3",
			Amount:          3000000,
			SpentAmount:     3500000,
			RemainingAmount: -500000,
			PercentageSpent: 116.67,
			Status:          domain.BudgetStatusExceeded,
		},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(budgets, nil)

	result, err := service.GetBudgetSummary(ctx, userID, time.Now())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.TotalBudgets)
	assert.Equal(t, 2, result.ActiveBudgets) // Active + Warning
	assert.Equal(t, 1, result.ExceededBudgets)
	assert.Equal(t, 1, result.WarningBudgets)
	assert.Equal(t, 18000000.0, result.TotalAmount)
	assert.Equal(t, 12500000.0, result.TotalSpent)
	assert.Equal(t, 5500000.0, result.TotalRemaining)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetSummary_EmptyBudgets(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	emptyBudgets := []domain.Budget{}

	mockRepo.On("FindByUserID", ctx, userID).Return(emptyBudgets, nil)

	result, err := service.GetBudgetSummary(ctx, userID, time.Now())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalBudgets)
	assert.Equal(t, 0, result.ActiveBudgets)
	assert.Equal(t, 0.0, result.TotalAmount)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetProgress_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	startDate := time.Now().AddDate(0, 0, -15) // 15 days ago
	endDate := time.Now().AddDate(0, 0, 15)    // 15 days from now

	budget := &domain.Budget{
		ID:              budgetID,
		UserID:          uuid.New(),
		Name:            "Test Budget",
		Period:          domain.BudgetPeriodMonthly,
		StartDate:       startDate,
		EndDate:         &endDate,
		Amount:          10000000,
		SpentAmount:     5000000,
		RemainingAmount: 5000000,
		PercentageSpent: 50.0,
		Status:          domain.BudgetStatusActive,
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)

	result, err := service.GetBudgetProgress(ctx, budgetID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, budgetID, result.BudgetID)
	assert.Equal(t, "Test Budget", result.Name)
	assert.Equal(t, 10000000.0, result.Amount)
	assert.Equal(t, 5000000.0, result.SpentAmount)
	assert.Equal(t, 50.0, result.PercentageSpent)
	assert.True(t, result.DaysElapsed > 0)
	assert.True(t, result.DaysRemaining > 0)
	mockRepo.AssertExpectations(t)
}

func TestBudgetReader_GetBudgetAnalytics_Success(t *testing.T) {
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

	result, err := service.GetBudgetAnalytics(ctx, budgetID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, budgetID, result.BudgetID)
	// More assertions can be added based on analytics implementation
	mockRepo.AssertExpectations(t)
}
