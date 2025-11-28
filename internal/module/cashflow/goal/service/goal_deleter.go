package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalDeleter handles goal deletion operations
type GoalDeleter struct {
	service *goalService
}

// NewGoalDeleter creates a new goal deleter
func NewGoalDeleter(service *goalService) *GoalDeleter {
	return &GoalDeleter{service: service}
}

// DeleteGoal deletes a goal
func (d *GoalDeleter) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	// Verify goal exists before deletion
	goal, err := d.service.repo.FindByID(ctx, goalID)
	if err != nil {
		d.service.logger.Error("Failed to find goal for deletion",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	d.service.logger.Info("Deleting goal",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.String("user_id", goal.UserID.String()),
	)

	return d.service.repo.Delete(ctx, goalID)
}
