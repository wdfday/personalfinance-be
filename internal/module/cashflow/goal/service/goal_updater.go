package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateGoal updates an existing goal
func (s *goalService) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	if err := s.validateGoal(goal); err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	s.logger.Info("Updating goal",
		zap.String("goal_id", goal.ID.String()),
		zap.String("name", goal.Name),
	)

	return s.repo.Update(ctx, goal)
}

// CalculateProgress recalculates progress for a goal
func (s *goalService) CalculateProgress(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to find goal for progress calculation",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	goal.UpdateCalculatedFields()

	s.logger.Info("Recalculated goal progress",
		zap.String("goal_id", goalID.String()),
		zap.Float64("percentage_complete", goal.PercentageComplete),
	)

	return s.repo.Update(ctx, goal)
}

// MarkAsCompleted marks a goal as completed
func (s *goalService) MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to find goal to mark as completed",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	goal.Status = domain.GoalStatusCompleted
	if goal.CompletedAt == nil {
		now := time.Now()
		goal.CompletedAt = &now
	}
	goal.UpdateCalculatedFields()

	s.logger.Info("Marked goal as completed",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.Float64("target_amount", goal.TargetAmount),
		zap.Float64("current_amount", goal.CurrentAmount),
	)

	return s.repo.Update(ctx, goal)
}

// CheckOverdueGoals checks and marks overdue goals
func (s *goalService) CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error {
	overdueGoals, err := s.repo.FindOverdueGoals(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find overdue goals",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	for _, goal := range overdueGoals {
		goal.Status = domain.GoalStatusOverdue
		if err := s.repo.Update(ctx, &goal); err != nil {
			s.logger.Error("Failed to mark goal as overdue",
				zap.String("goal_id", goal.ID.String()),
				zap.Error(err),
			)
		}
	}

	if len(overdueGoals) > 0 {
		s.logger.Info("Marked goals as overdue",
			zap.String("user_id", userID.String()),
			zap.Int("count", len(overdueGoals)),
		)
	}

	return nil
}

// validateGoal validates goal data
func (s *goalService) validateGoal(goal *domain.Goal) error {
	if goal.ID == uuid.Nil {
		return errors.New("goal ID is required")
	}

	if goal.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}

	if !goal.Category.IsValid() {
		return fmt.Errorf("invalid goal category: %s", goal.Category)
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
