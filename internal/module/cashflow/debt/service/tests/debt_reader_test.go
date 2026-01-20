package tests

import (
	"context"
	"errors"
	"testing"

	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDebtReader_GetDebtByID_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()
	expectedDebt := &domain.Debt{ID: id, Name: "Test Debt"}

	mockRepo.On("FindByID", ctx, id).Return(expectedDebt, nil)

	debt, err := svc.GetDebtByID(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, expectedDebt, debt)
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetDebtByID_NotFound(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("FindByID", ctx, id).Return(nil, errors.New("not found"))

	debt, err := svc.GetDebtByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, debt)
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetUserDebts_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	expectedDebts := []domain.Debt{{ID: uuid.New()}, {ID: uuid.New()}}

	mockRepo.On("FindByUserID", ctx, userID).Return(expectedDebts, nil)

	debts, err := svc.GetUserDebts(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, debts, 2)
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetActiveDebts_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	expectedDebts := []domain.Debt{{ID: uuid.New(), Status: domain.DebtStatusActive}}

	mockRepo.On("FindActiveByUserID", ctx, userID).Return(expectedDebts, nil)

	debts, err := svc.GetActiveDebts(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, debts, 1)
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetPaidOffDebts_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	expectedDebts := []domain.Debt{{ID: uuid.New(), Status: domain.DebtStatusPaidOff}}

	mockRepo.On("FindPaidOffDebts", ctx, userID).Return(expectedDebts, nil)

	debts, err := svc.GetPaidOffDebts(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, debts, 1)
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetDebtSummary_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	debts := []domain.Debt{
		{
			ID:              uuid.New(),
			Type:            domain.DebtTypeCreditCard,
			Status:          domain.DebtStatusActive,
			PrincipalAmount: 1000,
			CurrentBalance:  500,
			TotalPaid:       500,
		},
		{
			ID:              uuid.New(),
			Type:            domain.DebtTypePersonalLoan,
			Status:          domain.DebtStatusPaidOff,
			PrincipalAmount: 2000,
			CurrentBalance:  0,
			TotalPaid:       2000,
		},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(debts, nil)

	summary, err := svc.GetDebtSummary(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 2, summary.TotalDebts)
	assert.Equal(t, 1, summary.ActiveDebts)
	assert.Equal(t, 1, summary.PaidOffDebts)
	// Additional assertions on logic...
	mockRepo.AssertExpectations(t)
}

func TestDebtReader_GetDebtSummary_RepoError(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("FindByUserID", ctx, userID).Return(nil, errors.New("db error"))

	summary, err := svc.GetDebtSummary(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, summary)
	mockRepo.AssertExpectations(t)
}
