package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock implementation of repository.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, budget *domain.Budget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) FindByUserIDAndCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID, categoryID)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) FindByUserIDAndAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID, accountID)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) FindByPeriod(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]domain.Budget, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, budget *domain.Budget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) UpdateSpentAmount(ctx context.Context, id uuid.UUID, spentAmount float64) error {
	args := m.Called(ctx, id, spentAmount)
	return args.Error(0)
}

func (m *MockRepository) FindExpiredBudgets(ctx context.Context) ([]domain.Budget, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) FindBudgetsNeedingRecalculation(ctx context.Context, threshold time.Duration) ([]domain.Budget, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

func (m *MockRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRepository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockRepository) FindByUserIDPaginated(ctx context.Context, userID uuid.UUID, params repository.PaginationParams) (*repository.PaginatedResult, error) {
	args := m.Called(ctx, userID, params)
	return args.Get(0).(*repository.PaginatedResult), args.Error(1)
}

// Test helpers
func setupBudgetService() (*budgetService, *MockRepository) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err := db.Exec("CREATE TABLE transactions (user_id text, direction text, booking_date datetime, classification text, amount integer)").Error; err != nil {
		panic(err)
	}

	service := &budgetService{
		repo:   mockRepo,
		logger: logger,
		db:     db,
	}

	return service, mockRepo
}

func TestBudgetCreator_CreateBudget_Success(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	userID := uuid.New()
	categoryID := uuid.New()
	budget := &domain.Budget{
		UserID:     userID,
		Name:       "Monthly Groceries",
		Amount:     5000000,
		Period:     domain.BudgetPeriodMonthly,
		StartDate:  time.Now(),
		CategoryID: &categoryID,
	}

	mockRepo.On("FindByUserIDAndCategory", ctx, userID, categoryID).Return([]domain.Budget{}, nil)
	mockRepo.On("Create", ctx, budget).Return(nil)

	err := service.CreateBudget(ctx, budget)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, budget.SpentAmount)
	assert.Equal(t, 5000000.0, budget.RemainingAmount)
	assert.Equal(t, 0.0, budget.PercentageSpent)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCreator_CreateBudget_ValidationError_NoUserID(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:    uuid.Nil,
		Name:      "Test Budget",
		Amount:    5000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestBudgetCreator_CreateBudget_ValidationError_NegativeAmount(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    -1000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "budget amount must be greater than 0")
}

func TestBudgetCreator_CreateBudget_ValidationError_NoName(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:    uuid.New(),
		Name:      "",
		Amount:    5000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "budget name is required")
}

func TestBudgetCreator_CreateBudget_ValidationError_InvalidPeriod(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    5000000,
		Period:    domain.BudgetPeriod("invalid"),
		StartDate: time.Now(),
	}

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown budget period")
}

func TestBudgetCreator_CreateBudget_ValidationError_EndDateBeforeStart(t *testing.T) {
	service, _ := setupBudgetService()
	ctx := context.Background()

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, -1)

	budget := &domain.Budget{
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    5000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: startDate,
		EndDate:   &endDate,
	}

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestBudgetCreator_CreateBudget_WithDefaultAlertThresholds(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:       uuid.New(),
		Name:         "Test Budget",
		Amount:       5000000,
		Period:       domain.BudgetPeriodMonthly,
		StartDate:    time.Now(),
		EnableAlerts: true,
	}

	mockRepo.On("Create", ctx, budget).Return(nil)

	err := service.CreateBudget(ctx, budget)

	assert.NoError(t, err)
	assert.Len(t, budget.AlertThresholds, 3)
	assert.Contains(t, budget.AlertThresholds, domain.AlertAt75)
	assert.Contains(t, budget.AlertThresholds, domain.AlertAt90)
	assert.Contains(t, budget.AlertThresholds, domain.AlertAt100)
	mockRepo.AssertExpectations(t)
}

func TestBudgetCreator_CreateBudget_RepositoryError(t *testing.T) {
	service, mockRepo := setupBudgetService()
	ctx := context.Background()

	budget := &domain.Budget{
		UserID:    uuid.New(),
		Name:      "Test Budget",
		Amount:    5000000,
		Period:    domain.BudgetPeriodMonthly,
		StartDate: time.Now(),
	}

	expectedError := errors.New("database error")
	mockRepo.On("Create", ctx, budget).Return(expectedError)

	err := service.CreateBudget(ctx, budget)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
