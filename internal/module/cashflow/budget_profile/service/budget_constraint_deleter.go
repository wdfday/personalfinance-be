package service

import (
	"context"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteBudgetConstraint deletes a budget constraint
func (s *budgetConstraintService) DeleteBudgetConstraint(ctx context.Context, userID string, constraintID string) error {
	// Parse constraint ID
	constraintUUID, err := uuid.Parse(constraintID)
	if err != nil {
		return shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing budget constraint to verify ownership
	bc, err := s.repo.GetByID(ctx, constraintUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("id", constraintID)
		}
		return shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := uuid.Parse(userID)
	if bc.UserID != userUUID {
		return shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Delete from repository
	if err := s.repo.Delete(ctx, constraintUUID); err != nil {
		s.logger.Error("failed to delete budget constraint",
			zap.String("user_id", userID),
			zap.String("constraint_id", constraintID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint deleted",
		zap.String("user_id", userID),
		zap.String("constraint_id", constraintID))

	return nil
}
