package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGoalReader_GetGoalByID_Success(t *testing.T) {
	service, mockRepo := setupGoalService()
	ctx := context.Background()

	goalID := uuid.New()
	expectedGoal := &domain.Goal{
		ID:           goalID,
		UserID:       uuid.New(),
		Name:         "Test Goal",
		TargetAmount: 10000000,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(expectedGoal, nil)

	result, err := service.GetGoalByID(ctx, goalID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGoal, result)
	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetUserGoals_Success(t *testing.T) {
	service, mockRepo := setupGoalService()
	ctx := context.Background()

	userID := uuid.New()
	expectedGoals := []domain.Goal{
		{ID: uuid.New(), UserID: userID, Name: "Goal 1"},
		{ID: uuid.New(), UserID: userID, Name: "Goal 2"},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(expectedGoals, nil)

	result, err := service.GetUserGoals(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}
