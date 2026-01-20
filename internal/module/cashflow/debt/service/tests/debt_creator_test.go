package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDebtCreator_CreateDebt_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()

	debt := &domain.Debt{
		Name:            "Test Debt",
		Type:            domain.DebtTypeCreditCard,
		Behavior:        domain.DebtBehaviorRevolving,
		PrincipalAmount: 10000000,
		CurrentBalance:  5000000,
		Currency:        "VND",
		Status:          domain.DebtStatusActive,
		UserID:          uuid.New(),
		StartDate:       time.Now(),
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.Name == debt.Name && d.Type == debt.Type && d.Behavior == debt.Behavior
	})).Return(nil)

	err := svc.CreateDebt(ctx, debt)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDebtCreator_CreateDebt_ValidationError(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()

	debt := &domain.Debt{
		Name:            "Invalid Debt",
		PrincipalAmount: -100, // Invalid
	}

	err := svc.CreateDebt(ctx, debt)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestDebtCreator_CreateDebt_RepoError(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()

	debt := &domain.Debt{
		Name:            "Test Debt",
		Type:            domain.DebtTypePersonalLoan,
		Behavior:        domain.DebtBehaviorInstallment,
		PrincipalAmount: 10000000,
		CurrentBalance:  10000000,
		Currency:        "VND",
		Status:          domain.DebtStatusActive,
		UserID:          uuid.New(),
		StartDate:       time.Now(),
	}

	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

	err := svc.CreateDebt(ctx, debt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}
