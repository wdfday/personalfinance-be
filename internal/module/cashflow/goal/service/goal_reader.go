package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetGoalByID retrieves a goal by ID
func (s *goalService) GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error) {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get goal by ID",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goal, nil
}

// GetUserGoals retrieves all goals for a user
func (s *goalService) GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetActiveGoals retrieves all active goals for a user
func (s *goalService) GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get active goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetGoalsByCategory retrieves goals of a specific category
func (s *goalService) GetGoalsByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error) {
	goals, err := s.repo.FindByCategory(ctx, userID, category)
	if err != nil {
		s.logger.Error("Failed to get goals by category",
			zap.String("user_id", userID.String()),
			zap.String("goal_category", string(category)),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetCompletedGoals retrieves completed goals
func (s *goalService) GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := s.repo.FindCompletedGoals(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get completed goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetArchivedGoals retrieves archived goals
func (s *goalService) GetArchivedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := s.repo.FindByStatus(ctx, userID, domain.GoalStatusArchived)
	if err != nil {
		s.logger.Error("Failed to get archived goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetGoalSummary calculates and returns a summary of all goals for a user
func (s *goalService) GetGoalSummary(ctx context.Context, userID uuid.UUID) (*domain.GoalSummary, error) {
	goals, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get goals for summary",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	summary := &domain.GoalSummary{
		GoalsByCategory: make(map[string]*domain.GoalCategorySum),
		GoalsByPriority: make(map[string]int),
	}

	var totalProgress float64

	for _, goal := range goals {
		summary.TotalGoals++
		summary.TotalTargetAmount += goal.TargetAmount
		summary.TotalCurrentAmount += goal.CurrentAmount
		summary.TotalRemaining += goal.RemainingAmount
		totalProgress += goal.PercentageComplete

		// Count by status
		switch goal.Status {
		case domain.GoalStatusActive:
			summary.ActiveGoals++
		case domain.GoalStatusCompleted:
			summary.CompletedGoals++
		case domain.GoalStatusOverdue:
			summary.OverdueGoals++
		}

		// Sum by category
		categoryKey := string(goal.Category)
		if summary.GoalsByCategory[categoryKey] == nil {
			summary.GoalsByCategory[categoryKey] = &domain.GoalCategorySum{}
		}
		categorySum := summary.GoalsByCategory[categoryKey]
		categorySum.Count++
		categorySum.TargetAmount += goal.TargetAmount
		categorySum.CurrentAmount += goal.CurrentAmount
		if categorySum.TargetAmount > 0 {
			categorySum.Progress = (categorySum.CurrentAmount / categorySum.TargetAmount) * 100
		}

		// Count by priority
		priorityKey := string(goal.Priority)
		summary.GoalsByPriority[priorityKey]++
	}

	if summary.TotalGoals > 0 {
		summary.AverageProgress = totalProgress / float64(summary.TotalGoals)
	}

	s.logger.Info("Goal summary calculated",
		zap.String("user_id", userID.String()),
		zap.Int("total_goals", summary.TotalGoals),
		zap.Int("active_goals", summary.ActiveGoals),
	)

	return summary, nil
}

// GetGoalProgress retrieves detailed progress information for a goal
func (s *goalService) GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*domain.GoalProgress, error) {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get goal for progress",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	progress := &domain.GoalProgress{
		GoalID:             goal.ID,
		Name:               goal.Name,
		Behavior:           goal.Behavior,
		Category:           goal.Category,
		Priority:           goal.Priority,
		TargetAmount:       goal.TargetAmount,
		CurrentAmount:      goal.CurrentAmount,
		RemainingAmount:    goal.RemainingAmount,
		PercentageComplete: goal.PercentageComplete,
		Status:             goal.Status,
		StartDate:          goal.StartDate,
		TargetDate:         goal.TargetDate,
	}

	// Calculate time-based metrics
	now := time.Now()
	daysElapsed := int(now.Sub(goal.StartDate).Hours() / 24)
	progress.DaysElapsed = daysElapsed

	if goal.TargetDate != nil {
		daysRemaining := int(goal.TargetDate.Sub(now).Hours() / 24)
		progress.DaysRemaining = &daysRemaining

		totalDays := int(goal.TargetDate.Sub(goal.StartDate).Hours() / 24)
		if totalDays > 0 {
			timeProgress := float64(daysElapsed) / float64(totalDays) * 100
			progress.TimeProgress = &timeProgress

			// Check if on track
			onTrack := goal.PercentageComplete >= timeProgress
			progress.OnTrack = &onTrack
		}
	}

	// Calculate suggested contribution
	if goal.ContributionFrequency != nil {
		suggested := goal.CalculateSuggestedContribution(*goal.ContributionFrequency)
		progress.SuggestedContribution = &suggested
	}

	// Calculate projected completion date
	if goal.ContributionFrequency != nil && progress.SuggestedContribution != nil && *progress.SuggestedContribution > 0 {
		periodsRemaining := goal.RemainingAmount / *progress.SuggestedContribution
		daysPerPeriod := goal.ContributionFrequency.DaysPerPeriod()
		daysToCompletion := int(periodsRemaining * float64(daysPerPeriod))
		projectedDate := now.AddDate(0, 0, daysToCompletion)
		progress.ProjectedCompletionDate = &projectedDate
	}

	return progress, nil
}

// GetGoalAnalytics retrieves analytics for a goal
func (s *goalService) GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*domain.GoalAnalytics, error) {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get goal for analytics",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	analytics := &domain.GoalAnalytics{
		GoalID:             goal.ID,
		Name:               goal.Name,
		Behavior:           goal.Behavior,
		Category:           goal.Category,
		TargetAmount:       goal.TargetAmount,
		CurrentAmount:      goal.CurrentAmount,
		PercentageComplete: goal.PercentageComplete,
	}

	// Calculate velocity (amount per day)
	now := time.Now()
	daysElapsed := now.Sub(goal.StartDate).Hours() / 24
	if daysElapsed > 0 {
		analytics.Velocity = goal.CurrentAmount / daysElapsed
	}

	// Calculate estimated completion
	if analytics.Velocity > 0 {
		daysToCompletion := goal.RemainingAmount / analytics.Velocity
		estimatedDate := now.AddDate(0, 0, int(daysToCompletion))
		analytics.EstimatedCompletionDate = &estimatedDate
	}

	// Calculate risk level
	if goal.TargetDate != nil {
		daysRemaining := goal.TargetDate.Sub(now).Hours() / 24
		if daysRemaining > 0 {
			requiredVelocity := goal.RemainingAmount / daysRemaining
			if analytics.Velocity < requiredVelocity*0.8 {
				analytics.RiskLevel = "high"
			} else if analytics.Velocity < requiredVelocity {
				analytics.RiskLevel = "medium"
			} else {
				analytics.RiskLevel = "low"
			}
		} else if goal.RemainingAmount > 0 {
			analytics.RiskLevel = "overdue"
		}
	}

	// Calculate recommended contribution
	if goal.TargetDate != nil && goal.ContributionFrequency != nil {
		daysRemaining := goal.TargetDate.Sub(now).Hours() / 24
		if daysRemaining > 0 {
			daysPerPeriod := float64(goal.ContributionFrequency.DaysPerPeriod())
			periodsRemaining := daysRemaining / daysPerPeriod
			if periodsRemaining > 0 {
				recommended := goal.RemainingAmount / periodsRemaining
				analytics.RecommendedContribution = &recommended
			}
		}
	}

	return analytics, nil
}
