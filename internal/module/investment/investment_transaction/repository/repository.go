package repository

import (
	"context"
	"time"

	"personalfinancedss/internal/module/investment/investment_transaction/domain"
	"personalfinancedss/internal/module/investment/investment_transaction/dto"

	"github.com/google/uuid"
)

// Repository defines investment transaction data access operations
type Repository interface {
	// Create creates a new investment transaction
	Create(ctx context.Context, transaction *domain.InvestmentTransaction) error

	// GetByID retrieves an investment transaction by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentTransaction, error)

	// GetByUserID retrieves an investment transaction by ID and user ID
	GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.InvestmentTransaction, error)

	// List retrieves investment transactions with filters and pagination
	List(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) ([]*domain.InvestmentTransaction, int64, error)

	// GetByAssetID retrieves all transactions for a specific asset
	GetByAssetID(ctx context.Context, userID, assetID uuid.UUID) ([]*domain.InvestmentTransaction, error)

	// GetByDateRange retrieves transactions within a date range
	GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.InvestmentTransaction, error)

	// Update updates an investment transaction
	Update(ctx context.Context, transaction *domain.InvestmentTransaction) error

	// Delete soft deletes an investment transaction
	Delete(ctx context.Context, id uuid.UUID) error

	// GetSummary calculates transaction summary for given filters
	GetSummary(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error)
}
