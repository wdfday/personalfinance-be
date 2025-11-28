package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalContributor handles goal contribution operations
type GoalContributor struct {
	service *goalService
}

// NewGoalContributor creates a new goal contributor
func NewGoalContributor(service *goalService) *GoalContributor {
	return &GoalContributor{service: service}
}

// AddContribution adds a contribution to a goal
func (c *GoalContributor) AddContribution(ctx context.Context, goalID uuid.UUID, amount float64) error {
	if amount <= 0 {
		return errors.New("contribution amount must be greater than 0")
	}

	goal, err := c.service.repo.FindByID(ctx, goalID)
	if err != nil {
		c.service.logger.Error("Failed to find goal for contribution",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	// Record original amount for logging
	originalAmount := goal.CurrentAmount

	// Add contribution using domain logic
	goal.AddContribution(amount)

	if err := c.service.repo.Update(ctx, goal); err != nil {
		c.service.logger.Error("Failed to update goal after contribution",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return err
	}

	c.service.logger.Info("Contribution added to goal",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.Float64("amount", amount),
		zap.Float64("original_amount", originalAmount),
		zap.Float64("new_amount", goal.CurrentAmount),
		zap.Float64("percentage_complete", goal.PercentageComplete),
	)

	return nil
}
