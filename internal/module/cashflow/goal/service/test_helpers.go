package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/goal/domain"

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

func (m *MockRepository) FindByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error) {
	args := m.Called(ctx, userID, goalType)
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

// Test helpers
func setupGoalService() (*goalService, *MockRepository) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()

	service := &goalService{
		repo:   mockRepo,
		logger: logger,
	}

	// Initialize sub-services
	service.creator = NewGoalCreator(service)
	service.reader = NewGoalReader(service)
	service.updater = NewGoalUpdater(service)
	service.deleter = NewGoalDeleter(service)
	service.contributor = NewGoalContributor(service)

	return service, mockRepo
}
