package repository

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for debt data access
type Repository interface {
	// Create creates a new debt
	Create(ctx context.Context, debt *domain.Debt) error

	// FindByID retrieves a debt by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Debt, error)

	// FindByUserID retrieves all debts for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)

	// FindActiveByUserID retrieves all active debts for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)

	// FindByType retrieves debts of a specific type
	FindByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error)

	// FindByStatus retrieves debts with a specific status
	FindByStatus(ctx context.Context, userID uuid.UUID, status domain.DebtStatus) ([]domain.Debt, error)

	// FindPaidOffDebts retrieves paid off debts for a user
	FindPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)

	// FindOverdueDebts retrieves overdue debts
	FindOverdueDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error)

	// Update updates an existing debt
	Update(ctx context.Context, debt *domain.Debt) error

	// Delete soft deletes a debt
	Delete(ctx context.Context, id uuid.UUID) error

	// AddPayment adds a payment amount to a debt
	AddPayment(ctx context.Context, id uuid.UUID, amount float64) error
}
