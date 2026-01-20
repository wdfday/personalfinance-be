package service

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/calendar/month/domain"
	"personalfinancedss/internal/module/calendar/month/dto"
	"personalfinancedss/internal/module/calendar/month/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// ==================== Repository Mock ====================

type mockMonthRepository struct {
	mock.Mock
}

func (m *mockMonthRepository) GetMonthByID(ctx context.Context, id uuid.UUID) (*domain.Month, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Month), args.Error(1)
}

func (m *mockMonthRepository) UpdateMonth(ctx context.Context, month *domain.Month) error {
	return m.Called(ctx, month).Error(0)
}

func (m *mockMonthRepository) CreateMonth(ctx context.Context, month *domain.Month) error {
	return m.Called(ctx, month).Error(0)
}

func (m *mockMonthRepository) GetMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Month), args.Error(1)
}

func (m *mockMonthRepository) FindMonthsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Month, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Month), args.Error(1)
}

func (m *mockMonthRepository) FindMonthByUserIDAndPeriod(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Month), args.Error(1)
}

func (m *mockMonthRepository) FindOpenMonthByUserID(ctx context.Context, userID uuid.UUID) (*domain.Month, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Month), args.Error(1)
}

// Verify interface implementation at compile time
var _ repository.Repository = (*mockMonthRepository)(nil)

// ==================== Test Helpers ====================

func createTestMonth(userID uuid.UUID, status string) *domain.Month {
	monthID := uuid.New()
	now := time.Now()

	state := domain.NewMonthState()
	state.Version = 1
	state.CreatedAt = now
	state.DSSWorkflow = &domain.DSSWorkflowResults{
		CurrentStep:    0,
		CompletedSteps: []int{},
		StartedAt:      now,
		LastUpdated:    now,
	}

	month := &domain.Month{
		ID:        monthID,
		UserID:    userID,
		Month:     "2024-01",
		Status:    domain.MonthStatus(status),
		StartDate: now,
		EndDate:   now.AddDate(0, 1, 0),
		States:    []domain.MonthState{*state},
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	return month
}

// ==================== Apply Tests ====================

func TestApplyGoalPrioritizationSuccess(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{
		repo:   mockRepo,
		logger: zap.NewNop(),
	}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID

	goalIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	req := dto.ApplyGoalPrioritizationRequest{
		MonthID:         monthID,
		AcceptedRanking: goalIDs,
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)
	mockRepo.On("UpdateMonth", mock.Anything, month).Return(nil)

	err := svc.ApplyGoalPrioritization(context.Background(), req, &userID)

	assert.NoError(t, err)
	assert.NotNil(t, month.CurrentState().DSSWorkflow.GoalPrioritization)
	assert.Equal(t, "user_approved", month.CurrentState().DSSWorkflow.GoalPrioritization.Method)
	assert.Len(t, month.CurrentState().DSSWorkflow.GoalPrioritization.Rankings, 3)
	assert.Contains(t, month.CurrentState().DSSWorkflow.CompletedSteps, 1)

	mockRepo.AssertExpectations(t)
}

func TestApplyDebtStrategySuccess(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{
		repo:   mockRepo,
		logger: zap.NewNop(),
	}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID

	// Setup workflow - step 1 completed
	month.States[0].DSSWorkflow.CurrentStep = 1
	month.States[0].DSSWorkflow.CompletedSteps = []int{1}
	month.States[0].DSSWorkflow.GoalPrioritization = &domain.GoalPriorityResult{Method: "ahp"}

	req := dto.ApplyDebtStrategyRequest{
		MonthID:          monthID,
		SelectedStrategy: "avalanche",
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)
	mockRepo.On("UpdateMonth", mock.Anything, month).Return(nil)

	err := svc.ApplyDebtStrategy(context.Background(), req, &userID)

	assert.NoError(t, err)
	assert.NotNil(t, month.CurrentState().DSSWorkflow.DebtStrategy)
	assert.Equal(t, "avalanche", month.CurrentState().DSSWorkflow.DebtStrategy.Strategy)
	assert.Contains(t, month.CurrentState().DSSWorkflow.CompletedSteps, 2)

	mockRepo.AssertExpectations(t)
}

func TestApplyGoalDebtTradeoffSuccess(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{
		repo:   mockRepo,
		logger: zap.NewNop(),
	}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID

	// Steps 1 & 2 completed
	month.States[0].DSSWorkflow.CurrentStep = 2
	month.States[0].DSSWorkflow.CompletedSteps = []int{1, 2}

	req := dto.ApplyGoalDebtTradeoffRequest{
		MonthID:               monthID,
		GoalAllocationPercent: 60.0,
		DebtAllocationPercent: 40.0,
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)
	mockRepo.On("UpdateMonth", mock.Anything, month).Return(nil)

	err := svc.ApplyGoalDebtTradeoff(context.Background(), req, &userID)

	assert.NoError(t, err)
	assert.NotNil(t, month.CurrentState().DSSWorkflow.GoalDebtTradeoff)
	assert.Equal(t, 60.0, month.CurrentState().DSSWorkflow.GoalDebtTradeoff.GoalAllocationPercent)
	assert.Contains(t, month.CurrentState().DSSWorkflow.CompletedSteps, 3)

	mockRepo.AssertExpectations(t)
}

func TestApplyBudgetAllocationSuccess(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{
		repo:   mockRepo,
		logger: zap.NewNop(),
	}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID

	// Steps 1-3 completed
	month.States[0].DSSWorkflow.CurrentStep = 3
	month.States[0].DSSWorkflow.CompletedSteps = []int{1, 2, 3}

	req := dto.ApplyBudgetAllocationRequest{
		MonthID:          monthID,
		SelectedScenario: "balanced",
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)
	mockRepo.On("UpdateMonth", mock.Anything, month).Return(nil)

	err := svc.ApplyBudgetAllocation(context.Background(), req, &userID)

	assert.NoError(t, err)
	assert.NotNil(t, month.CurrentState().DSSWorkflow.BudgetAllocation)
	assert.Contains(t, month.CurrentState().DSSWorkflow.CompletedSteps, 4)
	assert.True(t, month.CurrentState().DSSWorkflow.IsComplete())

	mockRepo.AssertExpectations(t)
}

// ==================== Error Case Tests ====================

func TestApplyGoalPrioritizationUnauthorized(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	ownerID := uuid.New()
	otherUserID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(ownerID, "OPEN")
	month.ID = monthID

	req := dto.ApplyGoalPrioritizationRequest{
		MonthID:         monthID,
		AcceptedRanking: []uuid.UUID{uuid.New()},
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)

	err := svc.ApplyGoalPrioritization(context.Background(), req, &otherUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockRepo.AssertExpectations(t)
}

func TestApplyGoalPrioritizationClosedMonth(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "CLOSED")
	month.ID = monthID

	req := dto.ApplyGoalPrioritizationRequest{
		MonthID:         monthID,
		AcceptedRanking: []uuid.UUID{uuid.New()},
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)

	err := svc.ApplyGoalPrioritization(context.Background(), req, &userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot modify closed month")
	mockRepo.AssertExpectations(t)
}

func TestApplyDebtStrategyMissingPreviousStep(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID
	// Workflow still at step 0

	req := dto.ApplyDebtStrategyRequest{
		MonthID:          monthID,
		SelectedStrategy: "avalanche",
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)

	err := svc.ApplyDebtStrategy(context.Background(), req, &userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must complete step 1")
	mockRepo.AssertExpectations(t)
}

func TestApplyGoalDebtTradeoffInvalidAllocation(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID
	month.States[0].DSSWorkflow.CurrentStep = 2
	month.States[0].DSSWorkflow.CompletedSteps = []int{1, 2}

	req := dto.ApplyGoalDebtTradeoffRequest{
		MonthID:               monthID,
		GoalAllocationPercent: 70.0,
		DebtAllocationPercent: 20.0, // Sum = 90, invalid
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)

	err := svc.ApplyGoalDebtTradeoff(context.Background(), req, &userID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestApplyBudgetAllocationInvalidScenario(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID
	month.States[0].DSSWorkflow.CurrentStep = 3
	month.States[0].DSSWorkflow.CompletedSteps = []int{1, 2, 3}

	req := dto.ApplyBudgetAllocationRequest{
		MonthID:          monthID,
		SelectedScenario: "invalid_scenario",
	}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)

	err := svc.ApplyBudgetAllocation(context.Background(), req, &userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scenario")
	mockRepo.AssertExpectations(t)
}

// ==================== Workflow Management Tests ====================

func TestResetDSSWorkflowSuccess(t *testing.T) {
	mockRepo := new(mockMonthRepository)
	svc := &monthService{repo: mockRepo, logger: zap.NewNop()}

	userID := uuid.New()
	monthID := uuid.New()
	month := createTestMonth(userID, "OPEN")
	month.ID = monthID

	// Workflow has completed steps
	month.States[0].DSSWorkflow.CurrentStep = 2
	month.States[0].DSSWorkflow.CompletedSteps = []int{1, 2}
	month.States[0].DSSWorkflow.GoalPrioritization = &domain.GoalPriorityResult{}
	month.States[0].DSSWorkflow.DebtStrategy = &domain.DebtStrategyResult{}

	req := dto.ResetDSSWorkflowRequest{MonthID: monthID}

	mockRepo.On("GetMonthByID", mock.Anything, monthID).Return(month, nil)
	mockRepo.On("UpdateMonth", mock.Anything, month).Return(nil)

	err := svc.ResetDSSWorkflow(context.Background(), req, &userID)

	assert.NoError(t, err)
	assert.Equal(t, 0, month.CurrentState().DSSWorkflow.CurrentStep)
	assert.Empty(t, month.CurrentState().DSSWorkflow.CompletedSteps)
	assert.Nil(t, month.CurrentState().DSSWorkflow.GoalPrioritization)
	mockRepo.AssertExpectations(t)
}
