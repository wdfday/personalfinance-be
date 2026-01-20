package service_test

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock implementation of goal repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Goal), args.Error(1)
}

func (m *MockRepository) GetContributionsByDateRange(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error) {
	args := m.Called(ctx, goalID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GoalContribution), args.Error(1)
}

func (m *MockRepository) FindContributionsByGoalID(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	args := m.Called(ctx, goalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GoalContribution), args.Error(1)
}

// Implement other required repository methods as no-ops for testing
func (m *MockRepository) Create(ctx context.Context, goal *domain.Goal) error { return nil }
func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) FindByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) FindByStatus(ctx context.Context, userID uuid.UUID, status domain.GoalStatus) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) FindCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) FindOverdueGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockRepository) Update(ctx context.Context, goal *domain.Goal) error { return nil }
func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockRepository) AddContribution(ctx context.Context, id uuid.UUID, amount float64) error {
	return nil
}
func (m *MockRepository) CreateContribution(ctx context.Context, contribution *domain.GoalContribution) error {
	return nil
}
func (m *MockRepository) FindContributionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]domain.GoalContribution, error) {
	return nil, nil
}
func (m *MockRepository) GetNetContributionsByAccountID(ctx context.Context, accountID uuid.UUID) (float64, error) {
	return 0, nil
}
func (m *MockRepository) GetNetContributionsByGoalID(ctx context.Context, goalID uuid.UUID) (float64, error) {
	return 0, nil
}

func TestGetMonthSummary(t *testing.T) {
	goalID := uuid.New()
	startDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC)

	t.Run("Success - Valid contributions", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		// Mock goal
		goal := &domain.Goal{
			ID:   goalID,
			Name: "Emergency Fund",
		}

		// Mock contributions
		contributions := []domain.GoalContribution{
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    1000000,
				CreatedAt: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    500000,
				CreatedAt: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			},
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("GetContributionsByDateRange", mock.Anything, goalID, startDate, endDate).Return(contributions, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetMonthSummary(context.Background(), goalID, startDate, endDate)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, goalID, result.GoalID)
		assert.Equal(t, "Emergency Fund", result.Name)
		assert.Equal(t, 1500000.0, result.TotalContributed)
		assert.Equal(t, 2, result.ContributionCount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Mixed deposits and withdrawals", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		goal := &domain.Goal{
			ID:   goalID,
			Name: "Emergency Fund",
		}

		contributions := []domain.GoalContribution{
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    2000000,
				CreatedAt: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeWithdrawal,
				Amount:    500000,
				CreatedAt: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			},
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("GetContributionsByDateRange", mock.Anything, goalID, startDate, endDate).Return(contributions, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetMonthSummary(context.Background(), goalID, startDate, endDate)

		assert.NoError(t, err)
		assert.Equal(t, 1500000.0, result.TotalContributed) // 2M - 500K
		assert.Equal(t, 2, result.ContributionCount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Empty contributions", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		goal := &domain.Goal{
			ID:   goalID,
			Name: "Emergency Fund",
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("GetContributionsByDateRange", mock.Anything, goalID, startDate, endDate).Return([]domain.GoalContribution{}, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetMonthSummary(context.Background(), goalID, startDate, endDate)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.TotalContributed)
		assert.Equal(t, 0, result.ContributionCount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Goal not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		mockRepo.On("FindByID", mock.Anything, goalID).Return(nil, assert.AnError)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetMonthSummary(context.Background(), goalID, startDate, endDate)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get goal")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllTimeSummary(t *testing.T) {
	goalID := uuid.New()

	t.Run("Success - Multiple contributions", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		goal := &domain.Goal{
			ID:   goalID,
			Name: "Emergency Fund",
		}

		// Contributions ordered DESC (most recent first)
		firstDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		lastDate := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)

		contributions := []domain.GoalContribution{
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    500000,
				CreatedAt: lastDate, // Most recent
			},
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    1000000,
				CreatedAt: firstDate, // Oldest
			},
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("FindContributionsByGoalID", mock.Anything, goalID).Return(contributions, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetAllTimeSummary(context.Background(), goalID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, goalID, result.GoalID)
		assert.Equal(t, "Emergency Fund", result.Name)
		assert.Equal(t, 1500000.0, result.TotalContributed)
		assert.Equal(t, 0.0, result.TotalWithdrawn)
		assert.Equal(t, 1500000.0, result.NetContributed)
		assert.Equal(t, 2, result.ContributionCount)
		assert.NotNil(t, result.FirstContribution)
		assert.NotNil(t, result.LastContribution)
		assert.Equal(t, firstDate, *result.FirstContribution)
		assert.Equal(t, lastDate, *result.LastContribution)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - No contributions", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		goal := &domain.Goal{
			ID:   goalID,
			Name: "New Goal",
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("FindContributionsByGoalID", mock.Anything, goalID).Return([]domain.GoalContribution{}, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetAllTimeSummary(context.Background(), goalID)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.NetContributed)
		assert.Equal(t, 0, result.ContributionCount)
		assert.Nil(t, result.FirstContribution)
		assert.Nil(t, result.LastContribution)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - With withdrawals", func(t *testing.T) {
		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		goal := &domain.Goal{
			ID:   goalID,
			Name: "Vacation Fund",
		}

		contributions := []domain.GoalContribution{
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeWithdrawal,
				Amount:    300000,
				CreatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				GoalID:    goalID,
				Type:      domain.ContributionTypeDeposit,
				Amount:    1000000,
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
		}

		mockRepo.On("FindByID", mock.Anything, goalID).Return(goal, nil)
		mockRepo.On("FindContributionsByGoalID", mock.Anything, goalID).Return(contributions, nil)

		svc := service.NewService(mockRepo, nil, logger)

		result, err := svc.GetAllTimeSummary(context.Background(), goalID)

		assert.NoError(t, err)
		assert.Equal(t, 1000000.0, result.TotalContributed)
		assert.Equal(t, 300000.0, result.TotalWithdrawn)
		assert.Equal(t, 700000.0, result.NetContributed)

		mockRepo.AssertExpectations(t)
	})
}
