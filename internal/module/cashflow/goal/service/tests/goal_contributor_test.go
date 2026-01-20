package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoalContributor_AddContribution_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	goalID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()
	amount := 1000.0

	goal := &domain.Goal{
		ID:            goalID,
		AccountID:     accountID,
		UserID:        userID,
		TargetAmount:  5000,
		CurrentAmount: 2000,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("CreateContribution", ctx, mock.MatchedBy(func(c *domain.GoalContribution) bool {
		return c.Amount == amount && c.Type == domain.ContributionTypeDeposit
	})).Return(nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.CurrentAmount == 3000.0 // 2000 + 1000
	})).Return(nil)

	mockRepo.On("GetNetContributionsByAccountID", ctx, accountID).Return(500.0, nil)
	// Note: AccountService mock logic is skipped in setup but logged in real implementation.
	// Since we passed nil account service in setupGoalService, it won't be called.

	updatedGoal, err := svc.AddContribution(ctx, goalID, amount, nil, "manual")

	assert.NoError(t, err)
	assert.NotNil(t, updatedGoal)
	assert.Equal(t, 3000.0, updatedGoal.CurrentAmount)

	mockRepo.AssertExpectations(t)
}

func TestGoalContributor_AddContribution_ValidationError(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	_, err := svc.AddContribution(ctx, uuid.New(), -100, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "greater than 0")
}

func TestGoalContributor_WithdrawContribution_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	goalID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()
	amount := 500.0

	goal := &domain.Goal{
		ID:            goalID,
		AccountID:     accountID,
		UserID:        userID,
		TargetAmount:  5000,
		CurrentAmount: 2000,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("CreateContribution", ctx, mock.MatchedBy(func(c *domain.GoalContribution) bool {
		return c.Amount == amount && c.Type == domain.ContributionTypeWithdrawal
	})).Return(nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.CurrentAmount == 1500.0 // 2000 - 500
	})).Return(nil)

	mockRepo.On("GetNetContributionsByAccountID", ctx, accountID).Return(1500.0, nil)

	updatedGoal, err := svc.WithdrawContribution(ctx, goalID, amount, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, updatedGoal)
	assert.Equal(t, 1500.0, updatedGoal.CurrentAmount)

	mockRepo.AssertExpectations(t)
}

func TestGoalContributor_WithdrawContribution_InsufficientFunds(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	goalID := uuid.New()
	amount := 5000.0 // More than current

	goal := &domain.Goal{
		ID:            goalID,
		CurrentAmount: 2000,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)

	_, err := svc.WithdrawContribution(ctx, goalID, amount, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient goal balance")
	mockRepo.AssertNotCalled(t, "CreateContribution")
}
