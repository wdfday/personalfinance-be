package service

import (
	"context"
	"math"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/shared"
)

// GetTransaction retrieves a single transaction by ID
func (s *transactionService) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Parse transaction ID
	transactionUUID, err := parseUUID(transactionID, "transaction_id")
	if err != nil {
		return nil, err
	}

	// Retrieve transaction
	transaction, err := s.repo.GetByUserID(ctx, transactionUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return transaction, nil
}

// ListTransactions retrieves a list of transactions with filters and pagination
func (s *transactionService) ListTransactions(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionListResponse, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Set default pagination
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	} else if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Set default sorting (use booking_date instead of transaction_date)
	if query.SortBy == "" {
		query.SortBy = "booking_date"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	// Retrieve transactions from repository
	transactions, total, err := s.repo.List(ctx, userUUID, query)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))
	pagination := dto.PaginationInfo{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
		TotalCount: total,
	}

	// Get summary
	summary, err := s.repo.GetSummary(ctx, userUUID, query)
	if err != nil {
		// Log error but don't fail the request
		summary = nil
	}

	// Convert to response
	response := dto.ToTransactionListResponse(transactions, pagination, summary)

	return response, nil
}

// GetTransactionSummary retrieves transaction summary for given filters
func (s *transactionService) GetTransactionSummary(ctx context.Context, userID string, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Get summary from repository
	summary, err := s.repo.GetSummary(ctx, userUUID, query)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return summary, nil
}
