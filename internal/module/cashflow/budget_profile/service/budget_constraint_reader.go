package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// GetBudgetConstraint retrieves a budget constraint by ID
func (s *budgetConstraintService) GetBudgetConstraint(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, error) {
	// Parse constraint ID
	constraintUUID, err := uuid.Parse(constraintID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get budget constraint from repository
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

	return bc, nil
}

// GetBudgetConstraintWithHistory retrieves a budget constraint with its version history
func (s *budgetConstraintService) GetBudgetConstraintWithHistory(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, domain.BudgetConstraints, error) {
	// Parse IDs
	constraintUUID, err := parseConstraintID(constraintID)
	if err != nil {
		return nil, nil, shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get current constraint
	bc, err := s.repo.GetByID(ctx, constraintUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, nil, shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("id", constraintID)
		}
		return nil, nil, shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := parseUserID(userID)
	if bc.UserID != userUUID {
		return nil, nil, shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Get version history
	history, err := s.repo.GetVersionHistory(ctx, constraintUUID)
	if err != nil {
		return nil, nil, shared.ErrInternal.WithError(err)
	}

	return bc, history, nil
}

// GetBudgetConstraintByCategory retrieves a budget constraint by category
func (s *budgetConstraintService) GetBudgetConstraintByCategory(ctx context.Context, userID string, categoryID string) (*domain.BudgetConstraint, error) {
	// Parse user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Parse category ID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "category_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get budget constraint from repository
	bc, err := s.repo.GetByUserAndCategory(ctx, userUUID, categoryUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("category_id", categoryID)
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return bc, nil
}

// ListBudgetConstraints retrieves budget constraints with filters
func (s *budgetConstraintService) ListBudgetConstraints(ctx context.Context, userID string, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get budget constraints from repository
	constraints, err := s.repo.List(ctx, userUUID, query)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return constraints, nil
}

// GetActiveConstraints retrieves all currently active budget constraints for a user
func (s *budgetConstraintService) GetActiveConstraints(ctx context.Context, userID string) (domain.BudgetConstraints, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get active constraints from repository
	constraints, err := s.repo.GetActiveByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return constraints, nil
}

// GetArchivedConstraints retrieves all archived budget constraints for a user
func (s *budgetConstraintService) GetArchivedConstraints(ctx context.Context, userID string) (domain.BudgetConstraints, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get archived constraints from repository
	constraints, err := s.repo.GetArchivedByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return constraints, nil
}

// GetBudgetConstraintSummary retrieves summary of budget constraints for a user
func (s *budgetConstraintService) GetBudgetConstraintSummary(ctx context.Context, userID string) (*dto.BudgetConstraintSummaryResponse, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get all budget constraints for user
	constraints, err := s.repo.GetByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Calculate summary
	summary := &dto.BudgetConstraintSummaryResponse{
		TotalMandatoryExpenses: constraints.TotalMandatoryExpenses(),
		TotalFlexible:          len(constraints.GetFlexible()),
		TotalFixed:             len(constraints.GetFixed()),
		Count:                  len(constraints),
	}

	// Count active constraints
	activeCount := 0
	for _, bc := range constraints {
		if bc.IsActive() {
			activeCount++
		}
	}
	summary.ActiveCount = activeCount

	return summary, nil
}
