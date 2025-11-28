package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type goalService struct {
	repo   repository.Repository
	logger *zap.Logger

	// Sub-services
	creator     *GoalCreator
	reader      *GoalReader
	updater     *GoalUpdater
	deleter     *GoalDeleter
	contributor *GoalContributor
}

// NewService creates a new goal service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	service := &goalService{
		repo:   repo,
		logger: logger,
	}

	// Initialize sub-services
	service.creator = NewGoalCreator(service)
	service.reader = NewGoalReader(service)
	service.updater = NewGoalUpdater(service)
	service.deleter = NewGoalDeleter(service)
	service.contributor = NewGoalContributor(service)

	return service
}

// Create operations
func (s *goalService) CreateGoal(ctx context.Context, goal *domain.Goal) error {
	return s.creator.CreateGoal(ctx, goal)
}

// Read operations
func (s *goalService) GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error) {
	return s.reader.GetGoalByID(ctx, goalID)
}

func (s *goalService) GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.reader.GetUserGoals(ctx, userID)
}

func (s *goalService) GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.reader.GetActiveGoals(ctx, userID)
}

func (s *goalService) GetGoalsByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error) {
	return s.reader.GetGoalsByType(ctx, userID, goalType)
}

func (s *goalService) GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return s.reader.GetCompletedGoals(ctx, userID)
}

func (s *goalService) GetGoalSummary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error) {
	return s.reader.GetGoalSummary(ctx, userID)
}

func (s *goalService) GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*GoalProgress, error) {
	return s.reader.GetGoalProgress(ctx, goalID)
}

func (s *goalService) GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*GoalAnalytics, error) {
	return s.reader.GetGoalAnalytics(ctx, goalID)
}

// Update operations
func (s *goalService) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	return s.updater.UpdateGoal(ctx, goal)
}

func (s *goalService) CalculateProgress(ctx context.Context, goalID uuid.UUID) error {
	return s.updater.CalculateProgress(ctx, goalID)
}

func (s *goalService) MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error {
	return s.updater.MarkAsCompleted(ctx, goalID)
}

func (s *goalService) CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error {
	return s.updater.CheckOverdueGoals(ctx, userID)
}

// Delete operations
func (s *goalService) DeleteGoal(ctx context.Context, goalID uuid.UUID) error {
	return s.deleter.DeleteGoal(ctx, goalID)
}

// Contribution operations
func (s *goalService) AddContribution(ctx context.Context, goalID uuid.UUID, amount float64) error {
	return s.contributor.AddContribution(ctx, goalID, amount)
}
