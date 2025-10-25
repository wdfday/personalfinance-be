package repository

import (
	"context"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"

	"github.com/google/uuid"
)

// Repository defines transaction data access operations
type Repository interface {
	// Create creates a new transaction
	Create(ctx context.Context, transaction *domain.Transaction) error

	// GetByID retrieves a transaction by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)

	// GetByUserID retrieves a transaction by ID and user ID
	GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error)

	// GetByExternalID retrieves a transaction by external ID for a user (for import deduplication)
	GetByExternalID(ctx context.Context, userID uuid.UUID, externalID string) (*domain.Transaction, error)

	// List retrieves transactions with filters and pagination
	List(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) ([]*domain.Transaction, int64, error)

	// Update updates a transaction
	Update(ctx context.Context, transaction *domain.Transaction) error

	// UpdateColumns updates specific columns of a transaction
	UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]interface{}) error

	// Delete soft deletes a transaction
	Delete(ctx context.Context, id uuid.UUID) error

	// GetAccountBalance calculates the current balance for an account based on transactions
	GetAccountBalance(ctx context.Context, accountID uuid.UUID) (int64, error)

	// GetTransactionsByDateRange gets transactions within a date range
	GetTransactionsByDateRange(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, startDate, endDate time.Time) ([]*domain.Transaction, error)

	// GetSummary calculates transaction summary for given filters
	GetSummary(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error)

	// GetRecurringTransactions gets all manual recurring transaction templates (future use)
	GetRecurringTransactions(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error)
}
