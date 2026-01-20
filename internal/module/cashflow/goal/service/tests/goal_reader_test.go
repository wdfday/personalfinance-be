package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGoalReader_GetGoalByID_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()

	goalID := uuid.New()
	expectedGoal := &domain.Goal{
		ID:           goalID,
		UserID:       uuid.New(),
		AccountID:    uuid.New(),
		Name:         "Emergency Fund",
		TargetAmount: 10000,
		Category:     domain.GoalCategoryEmergency,
		Status:       domain.GoalStatusActive,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(expectedGoal, nil)

	goal, err := svc.GetGoalByID(ctx, goalID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGoal, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetGoalSummary_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	goals := []domain.Goal{
		{
			ID:                 uuid.New(), // Fixed missing ID
			UserID:             userID,
			Category:           domain.GoalCategorySavings,
			Priority:           domain.GoalPriorityHigh,
			Status:             domain.GoalStatusActive,
			TargetAmount:       1000,
			CurrentAmount:      500,
			RemainingAmount:    500,
			PercentageComplete: 50,
		},
		{
			ID:                 uuid.New(), // Fixed missing ID
			UserID:             userID,
			Category:           domain.GoalCategoryDebt,
			Priority:           domain.GoalPriorityMedium,
			Status:             domain.GoalStatusCompleted,
			TargetAmount:       2000,
			CurrentAmount:      2000,
			RemainingAmount:    0,
			PercentageComplete: 100,
		},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(goals, nil)

	summary, err := svc.GetGoalSummary(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 2, summary.TotalGoals)
	assert.Equal(t, 1, summary.ActiveGoals)
	assert.Equal(t, 1, summary.CompletedGoals)
	assert.Equal(t, 3000.0, summary.TotalTargetAmount)
	assert.Equal(t, 2500.0, summary.TotalCurrentAmount)
	assert.Equal(t, 75.0, summary.AverageProgress) // (50+100)/2

	// Check Categories
	assert.Contains(t, summary.GoalsByCategory, string(domain.GoalCategorySavings))
	assert.Equal(t, 1, summary.GoalsByCategory[string(domain.GoalCategorySavings)].Count)
	assert.Equal(t, 50.0, summary.GoalsByCategory[string(domain.GoalCategorySavings)].Progress)

	// Check Priorities
	assert.Equal(t, 1, summary.GoalsByPriority[string(domain.GoalPriorityHigh)])

	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetGoalProgress_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()

	startDate := time.Now().Add(-24 * time.Hour * 10)         // Started 10 days ago
	targetDate := time.Now().Add(24*time.Hour*10 + time.Hour) // Ends in 10 days + buffer (Total 20 days)

	goal := &domain.Goal{
		ID:                 goalID,
		Name:               "Test Goal",
		TargetAmount:       1000,
		CurrentAmount:      600,
		RemainingAmount:    400,
		PercentageComplete: 60,
		StartDate:          startDate,
		TargetDate:         &targetDate,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)

	progress, err := svc.GetGoalProgress(ctx, goalID)

	assert.NoError(t, err)
	assert.NotNil(t, progress)
	assert.Equal(t, goalID, progress.GoalID)
	assert.Equal(t, 10, progress.DaysElapsed)
	assert.Equal(t, 10, *progress.DaysRemaining)

	// Time progress is roughly 50% (10/20 days)
	// Goal progress is 60%
	// So OnTrack should be true
	if progress.TimeProgress != nil {
		assert.InDelta(t, 50.0, *progress.TimeProgress, 1.0)
	}
	if progress.OnTrack != nil {
		assert.True(t, *progress.OnTrack)
	}

	// Case 2: Verify Projected Completion Date and fresh Suggested Contribution
	monthlyFreq := domain.FrequencyMonthly
	staleSuggestion := 5.0 // DB has stale suggestion
	goal.ContributionFrequency = &monthlyFreq
	goal.SuggestedContribution = &staleSuggestion
	// remaining is 400.
	// Real suggestion should be 400 / (10 days remaining / 30) = 400 / 0.333 = 1200
	// Actually: 10 days remaining. Monthly freq.
	// CalculateSuggestedContribution:
	// periodsRemaining = 10 / 30 = 0.3333
	// suggest = 400 / 0.3333 = 1200.

	// Using stale suggestion (5.0):
	// periods = 400 / 5 = 80 periods.
	// days = 80 * 30 = 2400 days.

	// Using fresh suggestion (1200):
	// periods = 400 / 1200 = 0.333 periods.
	// days = 0.333 * 30 = 10 days.

	progressWithProj, err := svc.GetGoalProgress(ctx, goalID)
	assert.NoError(t, err)
	assert.NotNil(t, progressWithProj.SuggestedContribution)

	// Fresh suggestion should be calculated
	assert.Greater(t, *progressWithProj.SuggestedContribution, 1000.0) // ~1200

	// Projected completion should use fresh suggestion, so ~10 days from now (target date)
	if progressWithProj.ProjectedCompletionDate != nil {
		daysUntil := progressWithProj.ProjectedCompletionDate.Sub(time.Now()).Hours() / 24
		assert.InDelta(t, 10.0, daysUntil, 1.0)
	}

	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetGoalAnalytics_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	goalID := uuid.New()

	startDate := time.Now().Add(-24 * time.Hour * 10) // Started 10 days ago

	goal := &domain.Goal{
		ID:              goalID,
		Name:            "Test Goal Analytics",
		TargetAmount:    1000,
		CurrentAmount:   100, // 100 per 10 days = 10/day velocity
		RemainingAmount: 900,
		StartDate:       startDate,
	}

	mockRepo.On("FindByID", ctx, goalID).Return(goal, nil)

	analytics, err := svc.GetGoalAnalytics(ctx, goalID)

	assert.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.InDelta(t, 10.0, analytics.Velocity, 0.1)

	// Estimated completion: 900 / 10 = 90 days from now
	if analytics.EstimatedCompletionDate != nil {
		daysUntil := analytics.EstimatedCompletionDate.Sub(time.Now()).Hours() / 24
		assert.InDelta(t, 90.0, daysUntil, 1.0)
	}

	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetUserGoals_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	goals := []domain.Goal{
		{ID: uuid.New(), UserID: userID},
	}

	mockRepo.On("FindByUserID", ctx, userID).Return(goals, nil)

	result, err := svc.GetUserGoals(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetActiveGoals_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	goals := []domain.Goal{
		{ID: uuid.New(), UserID: userID, Status: domain.GoalStatusActive},
	}

	mockRepo.On("FindActiveByUserID", ctx, userID).Return(goals, nil)

	result, err := svc.GetActiveGoals(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetGoalsByCategory_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	category := domain.GoalCategorySavings

	goals := []domain.Goal{
		{ID: uuid.New(), UserID: userID, Category: category},
	}

	mockRepo.On("FindByCategory", ctx, userID, category).Return(goals, nil)

	result, err := svc.GetGoalsByCategory(ctx, userID, category)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestGoalReader_GetCompletedGoals_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := setupGoalService(t, mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	goals := []domain.Goal{
		{ID: uuid.New(), UserID: userID, Status: domain.GoalStatusCompleted},
	}

	mockRepo.On("FindCompletedGoals", ctx, userID).Return(goals, nil)

	result, err := svc.GetCompletedGoals(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}
