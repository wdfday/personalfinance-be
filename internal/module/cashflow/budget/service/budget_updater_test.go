package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBudgetUpdater_UpdateBudget_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "Updated Budget",
		Amount:    8000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	mockRepo.On("Update", ctx, budget).Return(nil)

	err := service.UpdateBudget(ctx, budget)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_UpdateBudget_ValidationError(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    -1000, // Invalid
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	err := service.UpdateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "budget amount must be greater than 0")
}

func TestBudgetUpdater_UpdateBudget_RepositoryError(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    5000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	expectedError := errors.New("database error")
	mockRepo.On("Update", ctx, budget).Return(expectedError)

	err := service.UpdateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_CheckBudgetAlerts_NoAlerts(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:              budgetID,
		UserID:          uuid.New(),
		Name:            "Test Budget",
		Amount:          10000000,
		SpentAmount:     3000000,
		PercentageSpent: 30.0,
		EnableAlerts:    true,
		AlertThresholds: []domain.AlertThreshold{domain.AlertAt75, domain.AlertAt90},
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)

	alerts, err := service.CheckBudgetAlerts(ctx, budgetID)

	assert.NoError(t, err)
	assert.Empty(t, alerts)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_CheckBudgetAlerts_HasAlerts(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:              budgetID,
		UserID:          uuid.New(),
		Name:            "Test Budget",
		Amount:          10000000,
		SpentAmount:     8000000,
		PercentageSpent: 80.0,
		EnableAlerts:    true,
		AlertThresholds: []domain.AlertThreshold{domain.AlertAt75, domain.AlertAt90},
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)

	alerts, err := service.CheckBudgetAlerts(ctx, budgetID)

	assert.NoError(t, err)
	assert.NotEmpty(t, alerts)
	assert.Contains(t, alerts, domain.AlertAt75)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_CheckBudgetAlerts_AlertsDisabled(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:              budgetID,
		UserID:          uuid.New(),
		Name:            "Test Budget",
		Amount:          10000000,
		SpentAmount:     8000000,
		PercentageSpent: 80.0,
		EnableAlerts:    false,
		AlertThresholds: []domain.AlertThreshold{domain.AlertAt75, domain.AlertAt90},
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)

	alerts, err := service.CheckBudgetAlerts(ctx, budgetID)

	assert.NoError(t, err)
	assert.Empty(t, alerts)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_MarkExpiredBudgets_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	yesterday := time.Now().AddDate(0, 0, -1)
	expiredBudgets := []domain.Budget{
		{
			ID:      uuid.New(),
			UserID:  uuid.New(),
			Name:    "Expired Budget 1",
			EndDate: &yesterday,
			Status:  domain.BudgetStatusActive,
		},
		{
			ID:      uuid.New(),
			UserID:  uuid.New(),
			Name:    "Expired Budget 2",
			EndDate: &yesterday,
			Status:  domain.BudgetStatusActive,
		},
	}

	mockRepo.On("FindExpiredBudgets", ctx).Return(expiredBudgets, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(b *domain.Budget) bool {
		return b.ID == expiredBudgets[0].ID && b.Status == domain.BudgetStatusExpired
	})).Return(nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(b *domain.Budget) bool {
		return b.ID == expiredBudgets[1].ID && b.Status == domain.BudgetStatusExpired
	})).Return(nil)

	err := service.MarkExpiredBudgets(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_MarkExpiredBudgets_NoExpiredBudgets(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	emptyBudgets := []domain.Budget{}

	mockRepo.On("FindExpiredBudgets", ctx).Return(emptyBudgets, nil)

	err := service.MarkExpiredBudgets(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetUpdater_MarkExpiredBudgets_RepositoryError(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	expectedError := errors.New("database error")
	mockRepo.On("FindExpiredBudgets", ctx).Return([]domain.Budget{}, expectedError)

	err := service.MarkExpiredBudgets(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
