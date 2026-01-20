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

func TestGoalUpdater_UpdateGoal_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	goalID := uuid.New()
	goal := &domain.Goal{
		ID:           goalID,
		UserID:       uuid.New(),
		AccountID:    uuid.New(),
		Name:         "Updated Goal",
		TargetAmount: 20000,
		Category:     domain.GoalCategorySavings,
		Priority:     domain.GoalPriorityHigh,
		Status:       domain.GoalStatusActive,
		StartDate:    time.Now().Add(-24 * time.Hour),
	}

	mockRepo.On("Update", ctx, goal).Return(nil)

	err := svc.UpdateGoal(ctx, goal)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalUpdater_UpdateGoal_RecalculateSuggestion(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	freq := domain.FrequencyMonthly
	goal := &domain.Goal{
		ID:                    uuid.New(),
		Name:                  "Recalc Goal",
		Category:              domain.GoalCategorySavings,
		Priority:              domain.GoalPriorityMedium,
		TargetAmount:          1200,
		CurrentAmount:         0,
		RemainingAmount:       0, // Will be recalc
		ContributionFrequency: &freq,
		StartDate:             time.Now(),
		TargetDate:            ptrTime(time.Now().Add(24 * time.Hour * 365)), // 1 year
	}
	// Expected suggestion: 1200 / 12 = 100

	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		if g.SuggestedContribution == nil {
			return false
		}
		return *g.SuggestedContribution >= 98 && *g.SuggestedContribution <= 102
	})).Return(nil)

	err := svc.UpdateGoal(ctx, goal)

	if err != nil {
		t.Logf("UpdateGoal failed: %v", err)
	}
	assert.NoError(t, err)

	if goal.SuggestedContribution == nil {
		t.Fatal("SuggestedContribution is nil")
	}
	// 365 days / 30 = 12.166 periods. 1200 / 12.166 = 98.6.
	// 366 days / 30 = 12.2 periods. 1200 / 12.2 = 98.36
	assert.InDelta(t, 100.0, *goal.SuggestedContribution, 2.0)

	mockRepo.AssertExpectations(t)
}

func TestGoalUpdater_CalculateProgress_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()

	goal := &domain.Goal{
		ID:                 goalID,
		TargetAmount:       1000,
		CurrentAmount:      500,
		PercentageComplete: 0, // Stale
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.PercentageComplete == 50.0
	})).Return(nil)

	err := svc.CalculateProgress(ctx, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalUpdater_MarkAsCompleted_Success(t *testing.T) {
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
		return g.Status == domain.GoalStatusCompleted && g.CompletedAt != nil
	})).Return(nil)

	err := svc.MarkAsCompleted(ctx, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
