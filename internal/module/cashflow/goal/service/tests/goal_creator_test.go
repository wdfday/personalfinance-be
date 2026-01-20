package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoalCreator_CreateGoal_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	year := 2024
	goal := &domain.Goal{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		AccountID:    uuid.New(),
		Name:         "Emergency Fund",
		TargetAmount: 10000,
		Category:     domain.GoalCategoryEmergency,
		Behavior:     domain.GoalBehaviorFlexible,
		Status:       domain.GoalStatusActive,
		Priority:     domain.GoalPriorityHigh, // Validation requires Priority
		StartDate:    time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Goal")).Return(nil)

	err := svc.CreateGoal(ctx, goal)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalCreator_CreateGoal_ValidationError(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	// Case 1: Missing ID (should fail validation)
	// Note: validation logic is often internal to service or domain.

	goalInvalidAmount := &domain.Goal{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		AccountID:    uuid.New(),
		Name:         "Invalid Amount",
		TargetAmount: -100, // Invalid
	}

	err := svc.CreateGoal(ctx, goalInvalidAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target amount must be greater than 0")

	mockRepo.AssertNotCalled(t, "Create")
}
