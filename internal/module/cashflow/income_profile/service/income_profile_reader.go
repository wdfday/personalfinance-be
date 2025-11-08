package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// GetIncomeProfile retrieves an income profile by ID
func (s *incomeProfileService) GetIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error) {
	// Parse profile ID
	profileUUID, err := uuid.Parse(profileID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get income profile from repository
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

	return ip, nil
}

// GetIncomeProfileByPeriod retrieves an income profile by period
func (s *incomeProfileService) GetIncomeProfileByPeriod(ctx context.Context, userID string, year, month int) (*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get income profile from repository
	ip, err := s.repo.GetByUserAndPeriod(ctx, userUUID, year, month)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.ErrNotFound.
				WithDetails("resource", "income_profile").
				WithDetails("year", year).
				WithDetails("month", month)
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return ip, nil
}

// ListIncomeProfiles retrieves income profiles with filters
func (s *incomeProfileService) ListIncomeProfiles(ctx context.Context, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get income profiles from repository
	profiles, err := s.repo.List(ctx, userUUID, query)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return profiles, nil
}
