package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGoalDeleter_DeleteGoal_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()
	goal := &domain.Goal{ID: goalID, UserID: uuid.New(), Name: "To Delete"}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("Delete", ctx, goalID).Return(nil)

	err := svc.DeleteGoal(ctx, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
