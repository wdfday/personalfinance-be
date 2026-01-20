package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoalArchiver_ArchiveGoal_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()

	goal := &domain.Goal{
		ID:     goalID,
		Status: domain.GoalStatusActive,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.Status == domain.GoalStatusArchived
	})).Return(nil)

	err := svc.ArchiveGoal(ctx, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalArchiver_UnarchiveGoal_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()

	goal := &domain.Goal{
		ID:     goalID,
		Status: domain.GoalStatusArchived,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.Status == domain.GoalStatusActive
	})).Return(nil)

	err := svc.UnarchiveGoal(ctx, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
