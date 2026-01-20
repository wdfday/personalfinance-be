package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"go.uber.org/zap"
)

// CreateGoal creates a new goal for a user
func (s *goalService) CreateGoal(ctx context.Context, goal *domain.Goal) error {
	// Calculate initial progress
	goal.UpdateCalculatedFields()

	if err := s.validateGoal(goal); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, goal); err != nil {
		s.logger.Error("Failed to create goal",
			zap.String("user_id", goal.UserID.String()),
			zap.String("goal_name", goal.Name),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Goal created successfully",
		zap.String("goal_id", goal.ID.String()),
		zap.String("user_id", goal.UserID.String()),
		zap.String("name", goal.Name),
		zap.String("behavior", string(goal.Behavior)),
		zap.String("category", string(goal.Category)),
		zap.Float64("target_amount", goal.TargetAmount),
	)

	return nil
}
