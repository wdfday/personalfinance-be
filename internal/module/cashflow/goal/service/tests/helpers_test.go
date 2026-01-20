package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock implementation of repository.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, goal *domain.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Goal), args.Error(1)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) FindByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error) {
	args := m.Called(ctx, userID, category)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) FindByStatus(ctx context.Context, userID uuid.UUID, status domain.GoalStatus) ([]domain.Goal, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) FindCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) FindOverdueGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, goal *domain.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) AddContribution(ctx context.Context, id uuid.UUID, amount float64) error {
	args := m.Called(ctx, id, amount)
	return args.Error(0)
}

func (m *MockRepository) CreateContribution(ctx context.Context, contribution *domain.GoalContribution) error {
	args := m.Called(ctx, contribution)
	return args.Error(0)
}

func (m *MockRepository) FindContributionsByGoalID(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	args := m.Called(ctx, goalID)
	return args.Get(0).([]domain.GoalContribution), args.Error(1)
}

func (m *MockRepository) FindContributionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]domain.GoalContribution, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]domain.GoalContribution), args.Error(1)
}

func (m *MockRepository) GetNetContributionsByAccountID(ctx context.Context, accountID uuid.UUID) (float64, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) GetNetContributionsByGoalID(ctx context.Context, goalID uuid.UUID) (float64, error) {
	args := m.Called(ctx, goalID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) GetContributionsByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.GoalContribution, error) {
	args := m.Called(ctx, userID, start, end)
	return args.Get(0).([]domain.GoalContribution), args.Error(1)
}

func (m *MockRepository) CalculateProgress(ctx context.Context, goalID uuid.UUID) error {
	args := m.Called(ctx, goalID)
	return args.Error(0)
}

func (m *MockRepository) FindArchivedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

// MockAccountService is a mock implementation of account service
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) UpdateAvailableBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	args := m.Called(ctx, accountID, amount)
	return args.Error(0)
}

// Test helper to create service with mocks
func setupGoalService(t *testing.T, mockRepo *MockRepository) service.Service {
	logger := zap.NewNop()
	// For now, pass nil for account service since it's not used in basic tests
	return service.NewService(mockRepo, nil, logger)
}
