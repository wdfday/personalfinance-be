package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"personalfinancedss/internal/module/cashflow/debt/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DebtCreator defines debt creation operations
type DebtCreator interface {
	CreateDebt(ctx context.Context, debt *domain.Debt) error
}

// DebtReader defines debt read operations
type DebtReader interface {
	GetDebtByID(ctx context.Context, debtID uuid.UUID) (*domain.Debt, error)
	GetUserDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)
	GetActiveDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)
	GetDebtsByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error)
	GetPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)
	GetDebtSummary(ctx context.Context, userID uuid.UUID) (*DebtSummary, error)
}

// DebtUpdater defines debt update operations
type DebtUpdater interface {
	UpdateDebt(ctx context.Context, debt *domain.Debt) error
	CalculateProgress(ctx context.Context, debtID uuid.UUID) error
	CheckOverdueDebts(ctx context.Context, userID uuid.UUID) error
}

// DebtDeleter defines debt delete operations
type DebtDeleter interface {
	DeleteDebt(ctx context.Context, debtID uuid.UUID) error
}

// DebtPaymentManager defines payment-related operations
type DebtPaymentManager interface {
	AddPayment(ctx context.Context, debtID uuid.UUID, amount float64) (*domain.Debt, error)
	MarkAsPaidOff(ctx context.Context, debtID uuid.UUID) error
}

// Service is the composite interface for all debt operations
type Service interface {
	DebtCreator
	DebtReader
	DebtUpdater
	DebtDeleter
	DebtPaymentManager
}

// debtService implements all debt use cases
type debtService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewService creates a new debt service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &debtService{
		repo:   repo,
		logger: logger,
	}
}

// DebtSummary represents a summary of user's debts
type DebtSummary struct {
	TotalDebts           int                     `json:"total_debts"`
	ActiveDebts          int                     `json:"active_debts"`
	PaidOffDebts         int                     `json:"paid_off_debts"`
	OverdueDebts         int                     `json:"overdue_debts"`
	TotalPrincipalAmount float64                 `json:"total_principal_amount"`
	TotalCurrentBalance  float64                 `json:"total_current_balance"`
	TotalPaid            float64                 `json:"total_paid"`
	TotalRemaining       float64                 `json:"total_remaining"`
	TotalInterestPaid    float64                 `json:"total_interest_paid"`
	AverageProgress      float64                 `json:"average_progress"`
	DebtsByType          map[string]*DebtTypeSum `json:"debts_by_type"`
	DebtsByStatus        map[string]int          `json:"debts_by_status"`
}

// DebtTypeSum represents summary for a debt type
type DebtTypeSum struct {
	Count           int     `json:"count"`
	PrincipalAmount float64 `json:"principal_amount"`
	CurrentBalance  float64 `json:"current_balance"`
	TotalPaid       float64 `json:"total_paid"`
	Progress        float64 `json:"progress"`
}
