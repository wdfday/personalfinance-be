package service

import (
	"context"

	"personalfinancedss/internal/module/calendar/month/domain"
	"personalfinancedss/internal/module/calendar/month/dto"
	"personalfinancedss/internal/module/calendar/month/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"

	// External services
	budgetService "personalfinancedss/internal/module/cashflow/budget/service"
	categoryservice "personalfinancedss/internal/module/cashflow/category/service"
	incomeprofilerepo "personalfinancedss/internal/module/cashflow/income_profile/repository"

	// Analytics services
	budgetAllocationService "personalfinancedss/internal/module/analytics/budget_allocation/service"
	cashflowForecastService "personalfinancedss/internal/module/analytics/cashflow_forecast/service"
	debtStrategyService "personalfinancedss/internal/module/analytics/debt_strategy/service"
	debtTradeoffService "personalfinancedss/internal/module/analytics/debt_tradeoff/service"
	goalService "personalfinancedss/internal/module/analytics/goal_prioritization/service"
)

// MonthCreator defines the interface for month creation
type MonthCreator interface {
	// CreateMonth creates a new month with deep copy from previous month
	CreateMonth(ctx context.Context, userID uuid.UUID, month string) (*domain.Month, error)

	// ReceiveIncome adds income to To Be Budgeted
	ReceiveIncome(ctx context.Context, req dto.IncomeReceivedRequest, userID *uuid.UUID) error

	// CloseMonth closes a month by month string and user ID
	CloseMonth(ctx context.Context, userID uuid.UUID, monthStr string) error
}

// MonthReader defines the interface for month read operations
type MonthReader interface {
	// GetMonth retrieves an existing month - returns error if not found
	// Use this for accessing any month (past, present, or future)
	GetMonth(ctx context.Context, userID uuid.UUID, monthStr string) (*dto.MonthViewResponse, error)

	// GetOrCreateCurrentMonth gets or creates ONLY the current month based on system date
	// Historical months must already exist or will error
	GetOrCreateCurrentMonth(ctx context.Context, userID uuid.UUID) (*dto.MonthViewResponse, error)

	// ListMonths retrieves all months for a user
	ListMonths(ctx context.Context, userID uuid.UUID) ([]*dto.MonthResponse, error)
}

// Service is the composite interface for all month operations
// It composes all use case interfaces for convenience
type Service interface {
	MonthCreator
	MonthReader
	MonthDSSWorkflowHandler // Sequential 5-step DSS workflow (0-4)
}

// monthService implements the Service interface
type monthService struct {
	repo              repository.Repository
	categoryService   categoryservice.Service
	incomeProfileRepo incomeprofilerepo.Repository
	budgetService     budgetService.Service
	logger            *zap.Logger

	// Analytics services for DSS pipeline
	goalPrioritization goalService.Service
	debtStrategy       debtStrategyService.Service
	debtTradeoff       debtTradeoffService.Service
	budgetAllocation   budgetAllocationService.Service
	cashflowForecast   cashflowForecastService.Service
	incomeMapper       *cashflowForecastService.IncomeForecastMapper

	// DSS cache for preview results
	dssCache *DSSCache
}

// NewService creates a new month service
func NewService(
	repo repository.Repository,
	categoryService categoryservice.Service,
	incomeProfileRepo incomeprofilerepo.Repository,
	budgetService budgetService.Service,
	goalPrioritization goalService.Service,
	debtStrategy debtStrategyService.Service,
	debtTradeoff debtTradeoffService.Service,
	budgetAllocation budgetAllocationService.Service,
	cashflowForecast cashflowForecastService.Service,
	dssCache *DSSCache,
	logger *zap.Logger,
) Service {
	return &monthService{
		repo:               repo,
		categoryService:    categoryService,
		incomeProfileRepo:  incomeProfileRepo,
		budgetService:      budgetService,
		goalPrioritization: goalPrioritization,
		debtStrategy:       debtStrategy,
		debtTradeoff:       debtTradeoff,
		budgetAllocation:   budgetAllocation,
		cashflowForecast:   cashflowForecast,
		incomeMapper:       cashflowForecastService.NewIncomeForecastMapper(),
		dssCache:           dssCache,
		logger:             logger,
	}
}
