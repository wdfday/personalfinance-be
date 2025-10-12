package service

import (
	"context"

	"personalfinancedss/internal/module/identify/profile/domain"
	profiledto "personalfinancedss/internal/module/identify/profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// CreateProfile creates a new profile for a user
func (s *profileService) CreateProfile(ctx context.Context, userID string, req profiledto.CreateProfileRequest) (*domain.UserProfile, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Check if profile already exists
	if _, err := s.repo.GetByUserID(ctx, userID); err == nil {
		return nil, shared.ErrConflict.WithDetails("resource", "profile").WithDetails("reason", "profile already exists")
	} else if err != shared.ErrNotFound {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Convert request to entity using conversion function
	profile, err := profiledto.FromCreateProfileRequest(req, userUUID)
	if err != nil {
		// Map domain errors to shared errors
		if err == domain.ErrInvalidRiskTolerance {
			return nil, shared.ErrBadRequest.WithDetails("field", "risk_tolerance").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidInvestmentHorizon {
			return nil, shared.ErrBadRequest.WithDetails("field", "investment_horizon").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidInvestmentExperience {
			return nil, shared.ErrBadRequest.WithDetails("field", "investment_experience").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidBudgetMethod {
			return nil, shared.ErrBadRequest.WithDetails("field", "budget_method").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidNotificationChannel {
			return nil, shared.ErrBadRequest.WithDetails("field", "notification_channels").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidReportFrequency {
			return nil, shared.ErrBadRequest.WithDetails("field", "report_frequency").WithDetails("reason", "invalid value")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create profile in repository
	if err := s.repo.Create(ctx, profile); err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return s.GetProfile(ctx, userID)
}
