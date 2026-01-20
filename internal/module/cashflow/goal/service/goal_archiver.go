package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ArchiveGoal marks a goal as archived
func (s *goalService) ArchiveGoal(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.GetGoalByID(ctx, goalID)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}

	if goal.Status == domain.GoalStatusArchived {
		return fmt.Errorf("goal is already archived")
	}

	goal.Status = domain.GoalStatusArchived

	if err := s.repo.Update(ctx, goal); err != nil {
		return fmt.Errorf("failed to archive goal: %w", err)
	}

	s.logger.Info("Goal archived",
		zap.String("goal_id", goalID.String()),
		zap.String("goal_name", goal.Name),
	)

	return nil
}

// UnarchiveGoal restores an archived goal to active status
func (s *goalService) UnarchiveGoal(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.GetGoalByID(ctx, goalID)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}

	if goal.Status != domain.GoalStatusArchived {
		return fmt.Errorf("goal is not archived")
	}

	// Restore to active status (or determine based on completion/deadline)
	if goal.CompletedAt != nil {
		goal.Status = domain.GoalStatusCompleted
	} else if goal.TargetDate != nil && goal.TargetDate.Before(time.Now()) {
		goal.Status = domain.GoalStatusOverdue
	} else {
		goal.Status = domain.GoalStatusActive
	}

	if err := s.repo.Update(ctx, goal); err != nil {
		return fmt.Errorf("failed to unarchive goal: %w", err)
	}

	s.logger.Info("Goal unarchived",
		zap.String("goal_id", goalID.String()),
		zap.String("goal_name", goal.Name),
		zap.String("new_status", string(goal.Status)),
	)

	return nil
}
