package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGoalCreator_CreateGoal_Success(t *testing.T) {
	service, mockRepo := setupGoalService()
	ctx := context.Background()

	goal := &domain.Goal{
		UserID:       uuid.New(),
		Name:         "Emergency Fund",
		Type:         domain.GoalTypeEmergency,
		Priority:     domain.GoalPriorityHigh,
		TargetAmount: 50000000,
		StartDate:    time.Now(),
	}

	mockRepo.On("Create", ctx, goal).Return(nil)

	err := service.CreateGoal(ctx, goal)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalCreator_CreateGoal_ValidationError(t *testing.T) {
	service, _ := setupGoalService()
	ctx := context.Background()

	goal := &domain.Goal{
		UserID:       uuid.Nil,
		Name:         "Test Goal",
		TargetAmount: 5000000,
	}

	err := service.CreateGoal(ctx, goal)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}
