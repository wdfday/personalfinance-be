package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"
)

// GetIncomeProfile retrieves an income profile by ID
func (s *incomeProfileService) GetIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error) {
	// Parse IDs
	profileUUID, err := parseProfileID(profileID)
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
	userUUID, _ := parseUserID(userID)
	if ip.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	return ip, nil
}

// GetIncomeProfileWithHistory retrieves an income profile with its version history
func (s *incomeProfileService) GetIncomeProfileWithHistory(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, []*domain.IncomeProfile, error) {
	// Parse IDs
	profileUUID, err := parseProfileID(profileID)
	if err != nil {
		return nil, nil, shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get current profile
	ip, err := s.repo.GetByID(ctx, profileUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, nil, shared.ErrNotFound.
				WithDetails("resource", "income_profile").
				WithDetails("id", profileID)
		}
		return nil, nil, shared.ErrInternal.WithError(err)
	}

	// Verify it belongs to the user
	userUUID, _ := parseUserID(userID)
	if ip.UserID != userUUID {
		return nil, nil, shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	// Get version history
	history, err := s.repo.GetVersionHistory(ctx, profileUUID)
	if err != nil {
		return nil, nil, shared.ErrInternal.WithError(err)
	}

	return ip, history, nil
}

// ListIncomeProfiles retrieves income profiles with filters
func (s *incomeProfileService) ListIncomeProfiles(ctx context.Context, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
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

// GetActiveIncomes retrieves all currently active income profiles for a user
func (s *incomeProfileService) GetActiveIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get active incomes from repository
	profiles, err := s.repo.GetActiveByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return profiles, nil
}

// GetArchivedIncomes retrieves all archived income profiles for a user
func (s *incomeProfileService) GetArchivedIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get archived incomes from repository
	profiles, err := s.repo.GetArchivedByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return profiles, nil
}

// GetRecurringIncomes retrieves all recurring income profiles for a user
func (s *incomeProfileService) GetRecurringIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get recurring incomes from repository
	profiles, err := s.repo.GetRecurringByUser(ctx, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return profiles, nil
}
