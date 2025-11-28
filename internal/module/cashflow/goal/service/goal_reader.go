package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalReader handles goal read operations
type GoalReader struct {
	service *goalService
}

// NewGoalReader creates a new goal reader
func NewGoalReader(service *goalService) *GoalReader {
	return &GoalReader{service: service}
}

// GetGoalByID retrieves a goal by ID
func (r *GoalReader) GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error) {
	goal, err := r.service.repo.FindByID(ctx, goalID)
	if err != nil {
		r.service.logger.Error("Failed to get goal by ID",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goal, nil
}

// GetUserGoals retrieves all goals for a user
func (r *GoalReader) GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := r.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		r.service.logger.Error("Failed to get user goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetActiveGoals retrieves all active goals for a user
func (r *GoalReader) GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := r.service.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		r.service.logger.Error("Failed to get active goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetGoalsByType retrieves goals of a specific type
func (r *GoalReader) GetGoalsByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error) {
	goals, err := r.service.repo.FindByType(ctx, userID, goalType)
	if err != nil {
		r.service.logger.Error("Failed to get goals by type",
			zap.String("user_id", userID.String()),
			zap.String("goal_type", string(goalType)),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetCompletedGoals retrieves completed goals
func (r *GoalReader) GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	goals, err := r.service.repo.FindCompletedGoals(ctx, userID)
	if err != nil {
		r.service.logger.Error("Failed to get completed goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return goals, nil
}

// GetGoalSummary calculates and returns a summary of all goals for a user
func (r *GoalReader) GetGoalSummary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error) {
	goals, err := r.service.repo.FindByUserID(ctx, userID)
	if err != nil {
		r.service.logger.Error("Failed to get goals for summary",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	summary := &GoalSummary{
		GoalsByType:     make(map[string]*GoalTypeSum),
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

		// Sum by type
		typeKey := string(goal.Type)
		if summary.GoalsByType[typeKey] == nil {
			summary.GoalsByType[typeKey] = &GoalTypeSum{}
		}
		typeSum := summary.GoalsByType[typeKey]
		typeSum.Count++
		typeSum.TargetAmount += goal.TargetAmount
		typeSum.CurrentAmount += goal.CurrentAmount
		if typeSum.TargetAmount > 0 {
			typeSum.Progress = (typeSum.CurrentAmount / typeSum.TargetAmount) * 100
		}

		// Count by priority
		priorityKey := string(goal.Priority)
		summary.GoalsByPriority[priorityKey]++
	}

	if summary.TotalGoals > 0 {
		summary.AverageProgress = totalProgress / float64(summary.TotalGoals)
	}

	r.service.logger.Info("Goal summary calculated",
		zap.String("user_id", userID.String()),
		zap.Int("total_goals", summary.TotalGoals),
		zap.Int("active_goals", summary.ActiveGoals),
	)

	return summary, nil
}

// GetGoalProgress retrieves detailed progress information for a goal
func (r *GoalReader) GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*GoalProgress, error) {
	goal, err := r.service.repo.FindByID(ctx, goalID)
	if err != nil {
		r.service.logger.Error("Failed to get goal for progress",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	progress := &GoalProgress{
		GoalID:             goal.ID,
		Name:               goal.Name,
		Type:               goal.Type,
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
	if goal.ContributionFrequency != nil && goal.SuggestedContribution != nil && *goal.SuggestedContribution > 0 {
		periodsRemaining := goal.RemainingAmount / *goal.SuggestedContribution
		daysPerPeriod := goal.ContributionFrequency.DaysPerPeriod()
		daysToCompletion := int(periodsRemaining * float64(daysPerPeriod))
		projectedDate := now.AddDate(0, 0, daysToCompletion)
		progress.ProjectedCompletionDate = &projectedDate
	}

	return progress, nil
}

// GetGoalAnalytics retrieves analytics for a goal
func (r *GoalReader) GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*GoalAnalytics, error) {
	goal, err := r.service.repo.FindByID(ctx, goalID)
	if err != nil {
		r.service.logger.Error("Failed to get goal for analytics",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	analytics := &GoalAnalytics{
		GoalID:             goal.ID,
		Name:               goal.Name,
		Type:               goal.Type,
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
