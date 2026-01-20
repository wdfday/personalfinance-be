package tests

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDebtUpdater_UpdateDebt_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()

	debt := &domain.Debt{
		ID:              uuid.New(),
		Name:            "Updated Debt",
		Type:            domain.DebtTypeCreditCard,
		Behavior:        domain.DebtBehaviorRevolving,
		PrincipalAmount: 10000000,
		CurrentBalance:  8000000,
		Currency:        "VND",
		Status:          domain.DebtStatusActive,
		StartDate:       time.Now(),
	}

	mockRepo.On("Update", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.Name == "Updated Debt"
	})).Return(nil)

	err := svc.UpdateDebt(ctx, debt)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDebtUpdater_UpdateDebt_ValidationError(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()

	debt := &domain.Debt{
		Name:            "Invalid Update",
		PrincipalAmount: -500,
	}

	err := svc.UpdateDebt(ctx, debt)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDebtUpdater_CalculateProgress_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()

	debt := &domain.Debt{
		ID:              id,
		PrincipalAmount: 1000,
		CurrentBalance:  500,
	}

	mockRepo.On("FindByID", ctx, id).Return(debt, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.PercentagePaid == 50.0 // 500/1000 * 100
	})).Return(nil)

	err := svc.CalculateProgress(ctx, id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDebtUpdater_CheckOverdueDebts_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	overdueDebts := []domain.Debt{
		{ID: uuid.New(), Name: "Overdue 1"},
	}

	mockRepo.On("FindOverdueDebts", ctx, userID).Return(overdueDebts, nil)
	// Iterate logic in service: marks status = Defaulted and Updates
	mockRepo.On("Update", ctx, mock.MatchedBy(func(d *domain.Debt) bool {
		return d.Status == domain.DebtStatusDefaulted
	})).Return(nil)

	err := svc.CheckOverdueDebts(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
