package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateBudgetConstraint updates an existing budget constraint
func (s *budgetConstraintService) UpdateBudgetConstraint(ctx context.Context, userID string, constraintID string, req dto.UpdateBudgetConstraintRequest) (*domain.BudgetConstraint, error) {
	// Parse constraint ID
	constraintUUID, err := uuid.Parse(constraintID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing budget constraint
	bc, err := s.repo.GetByID(ctx, constraintUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("id", constraintID)
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := uuid.Parse(userID)
	if bc.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Apply updates
	if err := dto.FromUpdateBudgetConstraintRequest(req, bc); err != nil {
		if err == domain.ErrNegativeAmount {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "amount").
				WithDetails("reason", "amount cannot be negative")
		}
		if err == domain.ErrMaximumBelowMinimum || err == domain.ErrMinimumExceedsMaximum {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "maximum_amount").
				WithDetails("reason", "maximum amount must be greater than or equal to minimum")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, bc); err != nil {
		s.logger.Error("failed to update budget constraint",
			zap.String("user_id", userID),
			zap.String("constraint_id", constraintID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint updated",
		zap.String("user_id", userID),
		zap.String("constraint_id", constraintID))

	return bc, nil
}
