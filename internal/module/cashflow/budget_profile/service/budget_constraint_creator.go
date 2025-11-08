package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateBudgetConstraint creates a new budget constraint
func (s *budgetConstraintService) CreateBudgetConstraint(ctx context.Context, userID string, req dto.CreateBudgetConstraintRequest) (*domain.BudgetConstraint, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "category_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Check if budget constraint already exists for this category
	exists, err := s.repo.Exists(ctx, userUUID, categoryID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}
	if exists {
		return nil, shared.ErrConflict.
			WithDetails("field", "category_id").
			WithDetails("reason", "budget constraint already exists for this category")
	}

	// Convert request to entity
	bc, err := dto.FromCreateBudgetConstraintRequest(req, userUUID)
	if err != nil {
		// Check for specific domain errors
		if err == domain.ErrInvalidCategoryID {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "category_id").
				WithDetails("reason", "invalid category ID")
		}
		if err == domain.ErrMaximumBelowMinimum {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "maximum_amount").
				WithDetails("reason", "maximum amount must be greater than or equal to minimum")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create budget constraint in repository
	if err := s.repo.Create(ctx, bc); err != nil {
		s.logger.Error("failed to create budget constraint",
			zap.String("user_id", userID),
			zap.String("category_id", req.CategoryID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint created",
		zap.String("user_id", userID),
		zap.String("constraint_id", bc.ID.String()),
		zap.String("category_id", req.CategoryID))

	return bc, nil
}
