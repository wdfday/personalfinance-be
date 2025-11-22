package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/investment_transaction/domain"
	"personalfinancedss/internal/module/investment/investment_transaction/dto"
	"personalfinancedss/internal/module/investment/investment_transaction/repository"

	"github.com/google/uuid"
)

// Service defines all investment transaction operations
type Service interface {
	CreateTransaction(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*domain.InvestmentTransaction, error)
	GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.InvestmentTransaction, error)
	ListTransactions(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionListResponse, error)
	GetByAsset(ctx context.Context, userID string, assetID string) ([]*domain.InvestmentTransaction, error)
	UpdateTransaction(ctx context.Context, userID string, transactionID string, req dto.UpdateTransactionRequest) (*domain.InvestmentTransaction, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) error
	GetSummary(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error)
}

// transactionService implements the Service interface
type transactionService struct {
	repo repository.Repository
}

// NewService creates a new transaction service
func NewService(repo repository.Repository) Service {
	return &transactionService{
		repo: repo,
	}
}

// CreateTransaction creates a new investment transaction
func (s *transactionService) CreateTransaction(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*domain.InvestmentTransaction, error) {
	// Validate transaction type
	if !req.TransactionType.IsValid() {
		return nil, fmt.Errorf("invalid transaction type: %s", req.TransactionType)
	}

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse asset ID
	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Set default currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Parse transaction date
	var transactionDate time.Time
	if req.TransactionDate != "" {
		transactionDate, err = time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction date format: %w", err)
		}
	} else {
		transactionDate = time.Now()
	}

	// Parse settlement date if provided
	var settlementDate *time.Time
	if req.SettlementDate != nil {
		sd, err := time.Parse("2006-01-02", *req.SettlementDate)
		if err != nil {
			return nil, fmt.Errorf("invalid settlement date format: %w", err)
		}
		settlementDate = &sd
	}

	// Set default status
	status := domain.TransactionStatusCompleted
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, fmt.Errorf("invalid transaction status: %s", *req.Status)
		}
		status = *req.Status
	}

	// Create transaction entity
	transaction := &domain.InvestmentTransaction{
		UserID:          uid,
		AssetID:         assetID,
		TransactionType: req.TransactionType,
		Quantity:        req.Quantity,
		PricePerUnit:    req.PricePerUnit,
		TotalAmount:     req.Quantity * req.PricePerUnit,
		Currency:        currency,
		Fees:            req.Fees,
		Commission:      req.Commission,
		Tax:             req.Tax,
		TransactionDate: transactionDate,
		SettlementDate:  settlementDate,
		Status:          status,
		Description:     req.Description,
		Notes:           req.Notes,
		Broker:          req.Broker,
		Exchange:        req.Exchange,
		OrderID:         req.OrderID,
		Tags:            req.Tags,
	}

	// Calculate total cost
	transaction.CalculateTotalCost()

	// Save to repository
	if err := s.repo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction, nil
}

// GetTransaction retrieves a single transaction by ID
func (s *transactionService) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.InvestmentTransaction, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	tid, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID: %w", err)
	}

	transaction, err := s.repo.GetByUserID(ctx, tid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return transaction, nil
}

// ListTransactions retrieves a list of transactions with pagination and filters
func (s *transactionService) ListTransactions(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	transactions, total, err := s.repo.List(ctx, uid, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	response := dto.ToTransactionListResponse(transactions, total, query.Page, query.PageSize)
	return &response, nil
}

// GetByAsset retrieves all transactions for a specific asset
func (s *transactionService) GetByAsset(ctx context.Context, userID string, assetID string) ([]*domain.InvestmentTransaction, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	transactions, err := s.repo.GetByAssetID(ctx, uid, aid)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by asset: %w", err)
	}

	return transactions, nil
}

// UpdateTransaction updates an existing transaction
func (s *transactionService) UpdateTransaction(ctx context.Context, userID string, transactionID string, req dto.UpdateTransactionRequest) (*domain.InvestmentTransaction, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	tid, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID: %w", err)
	}

	transaction, err := s.repo.GetByUserID(ctx, tid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Update fields if provided
	if req.Quantity != nil {
		transaction.Quantity = *req.Quantity
		transaction.TotalAmount = transaction.Quantity * transaction.PricePerUnit
	}
	if req.PricePerUnit != nil {
		transaction.PricePerUnit = *req.PricePerUnit
		transaction.TotalAmount = transaction.Quantity * transaction.PricePerUnit
	}
	if req.Fees != nil {
		transaction.Fees = *req.Fees
	}
	if req.Commission != nil {
		transaction.Commission = *req.Commission
	}
	if req.Tax != nil {
		transaction.Tax = *req.Tax
	}
	if req.TransactionDate != nil {
		td, err := time.Parse("2006-01-02", *req.TransactionDate)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction date format: %w", err)
		}
		transaction.TransactionDate = td
	}
	if req.SettlementDate != nil {
		sd, err := time.Parse("2006-01-02", *req.SettlementDate)
		if err != nil {
			return nil, fmt.Errorf("invalid settlement date format: %w", err)
		}
		transaction.SettlementDate = &sd
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, fmt.Errorf("invalid transaction status: %s", *req.Status)
		}
		transaction.Status = *req.Status
	}
	if req.Description != nil {
		transaction.Description = *req.Description
	}
	if req.Notes != nil {
		transaction.Notes = req.Notes
	}
	if req.Broker != nil {
		transaction.Broker = req.Broker
	}
	if req.Exchange != nil {
		transaction.Exchange = req.Exchange
	}
	if req.Tags != nil {
		transaction.Tags = req.Tags
	}

	// Recalculate total cost
	transaction.CalculateTotalCost()

	if err := s.repo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	return transaction, nil
}

// DeleteTransaction soft deletes a transaction
func (s *transactionService) DeleteTransaction(ctx context.Context, userID string, transactionID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	tid, err := uuid.Parse(transactionID)
	if err != nil {
		return fmt.Errorf("invalid transaction ID: %w", err)
	}

	// Verify transaction exists and belongs to user
	_, err = s.repo.GetByUserID(ctx, tid, uid)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	if err := s.repo.Delete(ctx, tid); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

// GetSummary retrieves transaction summary for a user
func (s *transactionService) GetSummary(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	summary, err := s.repo.GetSummary(ctx, uid, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	return summary, nil
}
