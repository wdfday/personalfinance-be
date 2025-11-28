package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GoalCreator handles goal creation operations
type GoalCreator struct {
	service *goalService
}

// NewGoalCreator creates a new goal creator
func NewGoalCreator(service *goalService) *GoalCreator {
	return &GoalCreator{service: service}
}

// CreateGoal creates a new financial goal
func (c *GoalCreator) CreateGoal(ctx context.Context, goal *domain.Goal) error {
	if err := c.validateGoal(goal); err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	// Calculate suggested contribution if frequency is provided
	if goal.ContributionFrequency != nil {
		contribution := goal.CalculateSuggestedContribution(*goal.ContributionFrequency)
		goal.SuggestedContribution = &contribution
	}

	c.service.logger.Info("Creating goal",
		zap.String("user_id", goal.UserID.String()),
		zap.String("name", goal.Name),
		zap.Float64("target_amount", goal.TargetAmount),
		zap.String("type", string(goal.Type)),
	)

	return c.service.repo.Create(ctx, goal)
}

// validateGoal validates goal data
func (c *GoalCreator) validateGoal(goal *domain.Goal) error {
	if goal.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	if goal.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}

	if goal.Name == "" {
		return errors.New("goal name is required")
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
