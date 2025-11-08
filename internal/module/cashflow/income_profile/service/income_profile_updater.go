package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateIncomeProfile updates an existing income profile
func (s *incomeProfileService) UpdateIncomeProfile(ctx context.Context, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	// Parse profile ID
	profileUUID, err := uuid.Parse(profileID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing income profile
	ip, err := s.repo.GetByID(ctx, profileUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrNotFound.
				WithDetails("resource", "income_profile").
				WithDetails("id", profileID)
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := uuid.Parse(userID)
	if ip.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	// Apply updates
	if err := dto.FromUpdateIncomeProfileRequest(req, ip); err != nil {
		if err == domain.ErrNegativeAmount {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "amount").
				WithDetails("reason", "amount cannot be negative")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, ip); err != nil {
		s.logger.Error("failed to update income profile",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile updated",
		zap.String("user_id", userID),
		zap.String("profile_id", profileID))

	return ip, nil
}
