package tests

import (
	"context"
	"testing"

	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDebtPayment_AddPayment_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()
	amount := 1000.0

	debt := &domain.Debt{
		ID:              id,
		CurrentBalance:  5000,
		PrincipalAmount: 10000,
	}

	mockRepo.On("FindByID", ctx, id).Return(debt, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.CurrentBalance == 4000 // 5000 - 1000
	})).Return(nil)

	updatedDebt, err := svc.AddPayment(ctx, id, amount)

	assert.NoError(t, err)
	assert.NotNil(t, updatedDebt)
	assert.Equal(t, 4000.0, updatedDebt.CurrentBalance)
	mockRepo.AssertExpectations(t)
}

func TestDebtPayment_AddPayment_InvalidAmount(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()
	amount := -100.0

	_, err := svc.AddPayment(ctx, id, amount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment amount must be greater than 0")
	mockRepo.AssertNotCalled(t, "FindByID")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDebtPayment_MarkAsPaidOff_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()

	debt := &domain.Debt{
		ID:             id,
		Status:         domain.DebtStatusActive,
		CurrentBalance: 1000,
	}

	mockRepo.On("FindByID", ctx, id).Return(debt, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.Status == domain.DebtStatusPaidOff && d.CurrentBalance == 0 && d.PaidOffDate != nil
	})).Return(nil)

	err := svc.MarkAsPaidOff(ctx, id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
