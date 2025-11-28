package service

import (
	"context"
	"personalfinancedss/internal/shared"

	"go.uber.org/zap"
)

// DeleteIncomeProfile soft deletes an income profile
func (s *incomeProfileService) DeleteIncomeProfile(ctx context.Context, userID string, profileID string) error {
	// Parse IDs
	profileUUID, err := parseProfileID(profileID)
	if err != nil {
		return shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing income profile to verify ownership
	ip, err := s.repo.GetByID(ctx, profileUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.ErrNotFound.
				WithDetails("resource", "income_profile").
				WithDetails("id", profileID)
		}
		return shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := parseUserID(userID)
	if ip.UserID != userUUID {
		return shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	// Soft delete from repository
	if err := s.repo.Delete(ctx, profileUUID); err != nil {
		s.logger.Error("failed to delete income profile",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile deleted",
		zap.String("user_id", userID),
		zap.String("profile_id", profileID))

	return nil
}
