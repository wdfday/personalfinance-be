package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBudgetDeleter_DeleteBudget_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:     budgetID,
		UserID: uuid.New(),
		Name:   "Test Budget",
		Amount: 5000000,
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)
	mockRepo.On("Delete", ctx, budgetID).Return(nil)

	err := service.DeleteBudget(ctx, budgetID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetDeleter_DeleteBudget_NotFound(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	expectedError := errors.New("budget not found")

	mockRepo.On("FindByID", ctx, budgetID).Return(nil, expectedError)

	err := service.DeleteBudget(ctx, budgetID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestBudgetDeleter_DeleteBudget_RepositoryError(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budgetID := uuid.New()
	budget := &domain.Budget{
		ID:     budgetID,
		UserID: uuid.New(),
		Name:   "Test Budget",
		Amount: 5000000,
	}

	mockRepo.On("FindByID", ctx, budgetID).Return(budget, nil)

	expectedError := errors.New("database error")
	mockRepo.On("Delete", ctx, budgetID).Return(expectedError)

	err := service.DeleteBudget(ctx, budgetID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
