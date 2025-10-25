package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/dto"

	"github.com/google/uuid"
)

// ImportJSONTransactions imports bank transactions from JSON format
func (s *transactionService) ImportJSONTransactions(ctx context.Context, userID string, req dto.ImportJSONRequest) (*dto.ImportJSONResponse, error) {
	// Parse and validate user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Parse and validate account ID
	accountUUID, err := parseUUID(req.AccountID, "accountId")
	if err != nil {
		return nil, err
	}

	response := &dto.ImportJSONResponse{
		TotalReceived: len(req.Transactions),
		ImportedIDs:   make([]string, 0),
		SkippedIDs:    make([]string, 0),
		Errors:        make([]dto.ImportError, 0),
	}

	var lastRunningBalance *int64
	var processedCount int

	// Process each transaction
	for _, bankTxn := range req.Transactions {
		// Check if transaction already exists by external ID
		if bankTxn.Additions.ExternalID != "" {
			existing, err := s.repo.GetByExternalID(ctx, userUUID, bankTxn.Additions.ExternalID)
			if err == nil && existing != nil {
				// Transaction already exists, skip it
				response.SkippedCount++
				response.SkippedIDs = append(response.SkippedIDs, bankTxn.ID)
				continue
			}
		}

		// Convert bank transaction to domain model
		transaction, err := bankTxn.ToBankTransactionDomain(userID, req.AccountID, req.BankCode)
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, dto.ImportError{
				BankTransactionID: bankTxn.ID,
				Error:             fmt.Sprintf("conversion error: %v", err),
			})
			continue
		}

		// Set user and account IDs
		transaction.UserID = userUUID
		transaction.AccountID = accountUUID
		transaction.ID = uuid.New()

		// Ensure timestamps
		ensureTimestamps(transaction)

		// Create transaction in repository
		if err := s.repo.Create(ctx, transaction); err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, dto.ImportError{
				BankTransactionID: bankTxn.ID,
				Error:             fmt.Sprintf("database error: %v", err),
			})
			continue
		}

		// Track the last running balance
		if transaction.RunningBalance != nil {
			lastRunningBalance = transaction.RunningBalance
		}

		response.SuccessCount++
		response.ImportedIDs = append(response.ImportedIDs, transaction.ID.String())
		processedCount++
	}

	// Sync account balance if we have a running balance from the last transaction
	if lastRunningBalance != nil && processedCount > 0 {
		// Get current account balance (if available from account module)
		// For now, we'll just return the sync info
		response.AccountBalance = &dto.AccountBalanceSync{
			AccountID:    req.AccountID,
			NewBalance:   *lastRunningBalance,
			LastSyncedAt: time.Now(),
		}
	}

	return response, nil
}
