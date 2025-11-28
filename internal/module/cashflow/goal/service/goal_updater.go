package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalUpdater handles goal update operations
type GoalUpdater struct {
	service *goalService
}

// NewGoalUpdater creates a new goal updater
func NewGoalUpdater(service *goalService) *GoalUpdater {
	return &GoalUpdater{service: service}
}

// UpdateGoal updates an existing goal
func (u *GoalUpdater) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	if err := u.validateGoal(goal); err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	// Recalculate suggested contribution if needed
	if goal.ContributionFrequency != nil {
		contribution := goal.CalculateSuggestedContribution(*goal.ContributionFrequency)
		goal.SuggestedContribution = &contribution
	}

	u.service.logger.Info("Updating goal",
		zap.String("goal_id", goal.ID.String()),
		zap.String("name", goal.Name),
	)

	return u.service.repo.Update(ctx, goal)
}

// CalculateProgress recalculates progress for a goal
func (u *GoalUpdater) CalculateProgress(ctx context.Context, goalID uuid.UUID) error {
	goal, err := u.service.repo.FindByID(ctx, goalID)
	if err != nil {
		u.service.logger.Error("Failed to find goal for progress calculation",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	goal.UpdateCalculatedFields()

	u.service.logger.Info("Recalculated goal progress",
		zap.String("goal_id", goalID.String()),
		zap.Float64("percentage_complete", goal.PercentageComplete),
	)

	return u.service.repo.Update(ctx, goal)
}

// MarkAsCompleted marks a goal as completed
func (u *GoalUpdater) MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error {
	goal, err := u.service.repo.FindByID(ctx, goalID)
	if err != nil {
		u.service.logger.Error("Failed to find goal to mark as completed",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	goal.Status = domain.GoalStatusCompleted
	goal.UpdateCalculatedFields()

	u.service.logger.Info("Marked goal as completed",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.Float64("target_amount", goal.TargetAmount),
		zap.Float64("current_amount", goal.CurrentAmount),
	)

	return u.service.repo.Update(ctx, goal)
}

// CheckOverdueGoals checks and marks overdue goals
func (u *GoalUpdater) CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error {
	overdueGoals, err := u.service.repo.FindOverdueGoals(ctx, userID)
	if err != nil {
		u.service.logger.Error("Failed to find overdue goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	for _, goal := range overdueGoals {
		goal.Status = domain.GoalStatusOverdue
		if err := u.service.repo.Update(ctx, &goal); err != nil {
			u.service.logger.Error("Failed to mark goal as overdue",
				zap.String("goal_id", goal.ID.String()),
				zap.Error(err),
			)
		}
	}

	if len(overdueGoals) > 0 {
		u.service.logger.Info("Marked goals as overdue",
			zap.String("user_id", userID.String()),
			zap.Int("count", len(overdueGoals)),
		)
	}

	return nil
}

// validateGoal validates goal data
func (u *GoalUpdater) validateGoal(goal *domain.Goal) error {
	if goal.ID == uuid.Nil {
		return errors.New("goal ID is required")
	}

	if goal.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}

	if !goal.Type.IsValid() {
		return fmt.Errorf("invalid goal type: %s", goal.Type)
	}

	if !goal.Priority.IsValid() {
		return fmt.Errorf("invalid goal priority: %s", goal.Priority)
	}

	if goal.TargetDate != nil && goal.TargetDate.Before(goal.StartDate) {
		return errors.New("target date must be after start date")
	}

	if goal.ContributionFrequency != nil && !goal.ContributionFrequency.IsValid() {
		return fmt.Errorf("invalid contribution frequency: %s", *goal.ContributionFrequency)
	}

	return nil
}
