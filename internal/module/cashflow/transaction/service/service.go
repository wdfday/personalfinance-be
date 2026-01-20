package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/module/cashflow/transaction/repository"
)

// TransactionCreator defines transaction creation operations
type TransactionCreator interface {
	CreateTransaction(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*domain.Transaction, error)
}

// TransactionReader defines transaction read operations
type TransactionReader interface {
	GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error)
	ListTransactions(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionListResponse, error)
	GetTransactionSummary(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error)
}

// TransactionUpdater defines transaction update operations
type TransactionUpdater interface {
	UpdateTransaction(ctx context.Context, userID string, transactionID string, req dto.UpdateTransactionRequest) (*domain.Transaction, error)
}

// TransactionDeleter defines transaction delete operations
type TransactionDeleter interface {
	DeleteTransaction(ctx context.Context, userID string, transactionID string) error
}

// Service is the composite interface for all transaction operations
type Service interface {
	TransactionCreator
	TransactionReader
	TransactionUpdater
	TransactionDeleter

	// ImportJSONTransactions imports bank transactions from JSON format
	ImportJSONTransactions(ctx context.Context, userID string, req dto.ImportJSONRequest) (*dto.ImportJSONResponse, error)
}

// transactionService implements all transaction use cases
type transactionService struct {
	repo          repository.Repository
	linkProcessor *LinkProcessor
}

// NewService creates a new transaction service
func NewService(repo repository.Repository, linkProcessor *LinkProcessor) Service {
	return &transactionService{
		repo:          repo,
		linkProcessor: linkProcessor,
	}
}
