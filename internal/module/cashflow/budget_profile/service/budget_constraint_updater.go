package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"

	"go.uber.org/zap"
)

// UpdateBudgetConstraint creates a NEW version and archives the old one (versioning pattern)
func (s *budgetConstraintService) UpdateBudgetConstraint(ctx context.Context, userID string, constraintID string, req dto.UpdateBudgetConstraintRequest) (*domain.BudgetConstraint, error) {
	// Parse IDs
	constraintUUID, err := parseConstraintID(constraintID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing budget constraint
	existing, err := s.repo.GetByID(ctx, constraintUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("id", constraintID)
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := parseUserID(userID)
	if existing.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Check if already archived
	if existing.IsArchived() {
		return nil, shared.ErrBadRequest.
			WithDetails("reason", "cannot update archived budget constraint")
	}

	// Create new version with updates applied
	newVersion, err := dto.ApplyUpdateBudgetConstraintRequest(req, existing)
	if err != nil {
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

	// Archive the old version
	existing.Archive(userUUID)
	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to archive old version",
			zap.String("user_id", userID),
			zap.String("constraint_id", constraintID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create the new version
	if err := s.repo.Create(ctx, newVersion); err != nil {
		s.logger.Error("failed to create new version",
			zap.String("user_id", userID),
			zap.String("old_constraint_id", constraintID),
			zap.String("new_constraint_id", newVersion.ID.String()),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint updated (new version created)",
		zap.String("user_id", userID),
		zap.String("old_constraint_id", constraintID),
		zap.String("new_constraint_id", newVersion.ID.String()))

	return newVersion, nil
}

// ArchiveBudgetConstraint manually archives a budget constraint (creates new version)
func (s *budgetConstraintService) ArchiveBudgetConstraint(ctx context.Context, userID string, constraintID string) error {
	// Parse IDs
	constraintUUID, err := parseConstraintID(constraintID)
	if err != nil {
		return shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing budget constraint to verify ownership
	existing, err := s.repo.GetByID(ctx, constraintUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.ErrNotFound.
				WithDetails("resource", "budget_constraint").
				WithDetails("id", constraintID)
		}
		return shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := parseUserID(userID)
	if existing.UserID != userUUID {
		return shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Check if already archived
	if existing.IsArchived() {
		return shared.ErrBadRequest.
			WithDetails("reason", "budget constraint is already archived")
	}

	// Archive the old version
	existing.Archive(userUUID)
	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to archive old version",
			zap.String("user_id", userID),
			zap.String("constraint_id", constraintID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint archived",
		zap.String("user_id", userID),
		zap.String("constraint_id", constraintID))

	return nil
}

// CheckAndArchiveEnded checks and archives ended budget constraints automatically
func (s *budgetConstraintService) CheckAndArchiveEnded(ctx context.Context, userID string) (int, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return 0, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get all active budget constraints
	constraints, err := s.repo.GetActiveByUser(ctx, userUUID)
	if err != nil {
		return 0, shared.ErrInternal.WithError(err)
	}

	endedCount := 0

	// Check each constraint and mark as ended if end_date has passed
	for _, bc := range constraints {
		if bc.CheckAndMarkAsEnded() {
			if err := s.repo.Update(ctx, bc); err != nil {
				s.logger.Error("failed to mark ended budget constraint",
					zap.String("user_id", userID),
					zap.String("constraint_id", bc.ID.String()),
					zap.Error(err))
				continue // Continue with other constraints
			}
			endedCount++
		}
	}

	if endedCount > 0 {
		s.logger.Info("marked ended budget constraints",
			zap.String("user_id", userID),
			zap.Int("count", endedCount))
	}

	return endedCount, nil
}

// EndBudgetConstraint marks a budget constraint as ended
func (s *budgetConstraintService) EndBudgetConstraint(ctx context.Context, userID string, constraintID string) (*domain.BudgetConstraint, error) {
	// Parse IDs
	constraintUUID, err := parseConstraintID(constraintID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "constraint_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing budget constraint to verify ownership
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
	userUUID, _ := parseUserID(userID)
	if bc.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "budget constraint does not belong to user")
	}

	// Check if already ended or archived
	if bc.Status == domain.ConstraintStatusEnded || bc.Status == domain.ConstraintStatusArchived {
		return nil, shared.ErrBadRequest.
			WithDetails("reason", "budget constraint is already ended or archived")
	}

	// Mark as ended
	bc.MarkAsEnded()

	// Update in repository
	if err := s.repo.Update(ctx, bc); err != nil {
		s.logger.Error("failed to end budget constraint",
			zap.String("user_id", userID),
			zap.String("constraint_id", constraintID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("budget constraint marked as ended",
		zap.String("user_id", userID),
		zap.String("constraint_id", constraintID))

	return bc, nil
}
