package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type goalService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewService creates a new goal service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &goalService{
		repo:   repo,
		logger: logger,
	}
}

func (s *goalService) CreateGoal(ctx context.Context, goal *domain.Goal) error {
	if err := s.validateGoal(goal); err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	// Calculate suggested contribution if frequency is provided
	if goal.ContributionFrequency != nil {
		contribution := goal.CalculateSuggestedContribution(*goal.ContributionFrequency)
		goal.SuggestedContribution = &contribution
	}

	return s.repo.Create(ctx, goal)
}

func (s *goalService) GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error) {
	return s.repo.FindByID(ctx, goalID)
}

func (s *goalService) GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *goalService) GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.repo.FindActiveByUserID(ctx, userID)
}

func (s *goalService) GetGoalsByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error) {
	return s.repo.FindByType(ctx, userID, goalType)
}

func (s *goalService) GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.repo.FindCompletedGoals(ctx, userID)
}

func (s *goalService) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	if err := s.validateGoal(goal); err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	// Recalculate suggested contribution if needed
	if goal.ContributionFrequency != nil {
		contribution := goal.CalculateSuggestedContribution(*goal.ContributionFrequency)
		goal.SuggestedContribution = &contribution
	}

	return s.repo.Update(ctx, goal)
}

func (s *goalService) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	return s.repo.Delete(ctx, goalID)
}

func (s *goalService) AddContribution(ctx context.Context, goalID uuid.UUID, amount float64) (*domain.Goal, error) {
	if amount <= 0 {
		return nil, errors.New("contribution amount must be greater than 0")
	}

	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		return nil, err
	}

	goal.AddContribution(amount)

	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}

	return goal, nil
}

func (s *goalService) CalculateProgress(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		return err
	}

	goal.UpdateCalculatedFields()

	return s.repo.Update(ctx, goal)
}

func (s *goalService) MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error {
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		return err
	}

	goal.Status = domain.GoalStatusCompleted
	goal.UpdateCalculatedFields()

	return s.repo.Update(ctx, goal)
}

func (s *goalService) CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error {
	overdueGoals, err := s.repo.FindOverdueGoals(ctx, userID)
	if err != nil {
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

	return nil
}

func (s *goalService) GetGoalSummary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error) {
	goals, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
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

	return summary, nil
}

func (s *goalService) validateGoal(goal *domain.Goal) error {
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
