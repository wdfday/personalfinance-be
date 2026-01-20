package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type budgetService struct {
	repo   repository.Repository
	db     *gorm.DB
	logger *zap.Logger

	// Sub-services
	creator    *BudgetCreator
	reader     *BudgetReader
	updater    *BudgetUpdater
	deleter    *BudgetDeleter
	calculator *BudgetCalculator
}

// NewService creates a new budget service
func NewService(repo repository.Repository, db *gorm.DB, logger *zap.Logger) Service {
	service := &budgetService{
		repo:   repo,
		db:     db,
		logger: logger,
	}

	// Initialize sub-services
	service.creator = NewBudgetCreator(service)
	service.reader = NewBudgetReader(service)
	service.updater = NewBudgetUpdater(service)
	service.deleter = NewBudgetDeleter(service)
	service.calculator = NewBudgetCalculator(service)

	return service
}

// Create operations
func (s *budgetService) CreateBudget(ctx context.Context, budget *domain.Budget) error {
	return s.creator.CreateBudget(ctx, budget)
}

// Read operations
func (s *budgetService) GetBudgetByID(ctx context.Context, budgetID uuid.UUID) (*domain.Budget, error) {
	return s.reader.GetBudgetByID(ctx, budgetID)
}

// GetBudgetByIDForUser retrieves a budget by ID with ownership verification
func (s *budgetService) GetBudgetByIDForUser(ctx context.Context, budgetID, userID uuid.UUID) (*domain.Budget, error) {
	return s.reader.GetBudgetByIDForUser(ctx, budgetID, userID)
}

func (s *budgetService) GetUserBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	return s.reader.GetUserBudgets(ctx, userID)
}

// GetUserBudgetsPaginated retrieves budgets for a user with pagination
func (s *budgetService) GetUserBudgetsPaginated(ctx context.Context, userID uuid.UUID, page, pageSize int) (*repository.PaginatedResult, error) {
	return s.reader.GetUserBudgetsPaginated(ctx, userID, page, pageSize)
}

func (s *budgetService) GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	return s.reader.GetActiveBudgets(ctx, userID)
}

func (s *budgetService) GetBudgetsByCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error) {
	return s.reader.GetBudgetsByCategory(ctx, userID, categoryID)
}

func (s *budgetService) GetBudgetsByAccount(ctx context.Context, userID, accountID uuid.UUID) ([]domain.Budget, error) {
	return s.reader.GetBudgetsByAccount(ctx, userID, accountID)
}

func (s *budgetService) GetBudgetsByPeriod(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod) ([]domain.Budget, error) {
	return s.reader.GetBudgetsByPeriod(ctx, userID, period)
}

func (s *budgetService) GetBudgetSummary(ctx context.Context, userID uuid.UUID, period time.Time) (*BudgetSummary, error) {
	return s.reader.GetBudgetSummary(ctx, userID, period)
}

func (s *budgetService) GetBudgetVsActual(ctx context.Context, userID uuid.UUID, period domain.BudgetPeriod, startDate, endDate time.Time) ([]*BudgetVsActual, error) {
	return s.reader.GetBudgetVsActual(ctx, userID, period, startDate, endDate)
}

// GetBudgetProgress retrieves detailed budget progress with ownership verification
func (s *budgetService) GetBudgetProgress(ctx context.Context, budgetID, userID uuid.UUID) (*BudgetProgress, error) {
	return s.reader.GetBudgetProgressForUser(ctx, budgetID, userID)
}

// GetBudgetAnalytics retrieves budget analytics with ownership verification
func (s *budgetService) GetBudgetAnalytics(ctx context.Context, budgetID, userID uuid.UUID) (*BudgetAnalytics, error) {
	return s.reader.GetBudgetAnalyticsForUser(ctx, budgetID, userID)
}

// Update operations
func (s *budgetService) UpdateBudget(ctx context.Context, budget *domain.Budget) error {
	return s.updater.UpdateBudget(ctx, budget)
}

// UpdateBudgetForUser updates a budget with ownership verification
func (s *budgetService) UpdateBudgetForUser(ctx context.Context, budget *domain.Budget, userID uuid.UUID) error {
	return s.updater.UpdateBudgetForUser(ctx, budget, userID)
}

func (s *budgetService) CheckBudgetAlerts(ctx context.Context, budgetID uuid.UUID) ([]domain.AlertThreshold, error) {
	return s.updater.CheckBudgetAlerts(ctx, budgetID)
}

func (s *budgetService) MarkExpiredBudgets(ctx context.Context) error {
	return s.updater.MarkExpiredBudgets(ctx)
}

// Delete operations
func (s *budgetService) DeleteBudget(ctx context.Context, budgetID uuid.UUID) error {
	return s.deleter.DeleteBudget(ctx, budgetID)
}

// DeleteBudgetForUser deletes a budget with ownership verification
func (s *budgetService) DeleteBudgetForUser(ctx context.Context, budgetID, userID uuid.UUID) error {
	return s.deleter.DeleteBudgetForUser(ctx, budgetID, userID)
}

// Calculation operations
func (s *budgetService) RecalculateBudgetSpending(ctx context.Context, budgetID uuid.UUID) error {
	return s.calculator.RecalculateBudgetSpending(ctx, budgetID)
}

// RecalculateBudgetSpendingForUser recalculates budget spending with ownership verification
func (s *budgetService) RecalculateBudgetSpendingForUser(ctx context.Context, budgetID, userID uuid.UUID) error {
	return s.calculator.RecalculateBudgetSpendingForUser(ctx, budgetID, userID)
}

func (s *budgetService) RecalculateAllBudgets(ctx context.Context, userID uuid.UUID) error {
	return s.calculator.RecalculateAllBudgets(ctx, userID)
}

func (s *budgetService) RolloverBudgets(ctx context.Context, userID uuid.UUID) error {
	return s.calculator.RolloverBudgets(ctx, userID)
}
