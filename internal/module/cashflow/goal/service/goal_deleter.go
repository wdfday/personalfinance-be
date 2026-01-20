package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteGoal deletes a goal
func (s *goalService) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	// Verify goal exists before deletion
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to find goal for deletion",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Deleting goal",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.String("user_id", goal.UserID.String()),
	)

	return s.repo.Delete(ctx, goalID)
}
