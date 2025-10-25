package service

import (
	"context"

	"personalfinancedss/internal/shared"
)

// DeleteTransaction soft deletes a transaction
func (s *transactionService) DeleteTransaction(ctx context.Context, userID string, transactionID string) error {
	// Parse user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return err
	}

	// Parse transaction ID
	transactionUUID, err := parseUUID(transactionID, "transaction_id")
	if err != nil {
		return err
	}

	// Verify transaction belongs to user
	_, err = s.repo.GetByUserID(ctx, transactionUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}

	// Delete transaction
	if err := s.repo.Delete(ctx, transactionUUID); err != nil {
		if err == shared.ErrNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}

	return nil
}
