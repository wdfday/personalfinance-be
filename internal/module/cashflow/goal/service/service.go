package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	"time"

	"github.com/google/uuid"
)

// GoalCreator defines the interface for creating goals
type GoalCreator interface {
	CreateGoal(ctx context.Context, goal *domain.Goal) error
}

// GoalReader defines the interface for reading goals
type GoalReader interface {
	GetGoalByID(ctx context.Context, goalID uuid.UUID) (*domain.Goal, error)
	GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetGoalsByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error)
	GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetArchivedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error)
	GetGoalSummary(ctx context.Context, userID uuid.UUID) (*domain.GoalSummary, error)
	GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*domain.GoalProgress, error)
	GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*domain.GoalAnalytics, error)
}

// GoalUpdater defines the interface for updating goals
type GoalUpdater interface {
	UpdateGoal(ctx context.Context, goal *domain.Goal) error
	CalculateProgress(ctx context.Context, goalID uuid.UUID) error
	MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error
	CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error
}

// GoalArchiver defines the interface for archiving/unarchiving goals
type GoalArchiver interface {
	ArchiveGoal(ctx context.Context, goalID uuid.UUID) error
	UnarchiveGoal(ctx context.Context, goalID uuid.UUID) error
}

// GoalDeleter defines the interface for deleting goals
type GoalDeleter interface {
	DeleteGoal(ctx context.Context, goalID uuid.UUID) error
}

// GoalContributor defines the interface for goal contributions
type GoalContributor interface {
	AddContribution(ctx context.Context, goalID uuid.UUID, amount float64, accountID *uuid.UUID, note *string, source string) (*domain.Goal, error)
	WithdrawContribution(ctx context.Context, goalID uuid.UUID, amount float64, note *string, reversingID *uuid.UUID) (*domain.Goal, error)
	GetContributions(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error)
	GetGoalNetContributions(ctx context.Context, goalID uuid.UUID) (float64, error)

	// Time-series query methods for month aggregation
	GetMonthContributions(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error)
	GetMonthSummary(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) (*dto.GoalMonthlySummary, error)

	// All-time summary (from inception to present)
	GetAllTimeSummary(ctx context.Context, goalID uuid.UUID) (*dto.GoalAllTimeSummary, error)
}

// Service is the composite interface for all goal operations
type Service interface {
	GoalCreator
	GoalReader
	GoalUpdater
	GoalArchiver
	GoalDeleter
	GoalContributor
}
