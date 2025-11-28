package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"go.uber.org/zap"
)

// CreateIncomeProfile creates a new income profile
func (s *incomeProfileService) CreateIncomeProfile(ctx context.Context, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Convert request to entity
	ip, err := dto.FromCreateIncomeProfileRequest(req, userUUID)
	if err != nil {
		// Check for specific domain errors
		if err == domain.ErrInvalidSource {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "source").
				WithDetails("reason", "source cannot be empty")
		}
		if err == domain.ErrInvalidFrequency {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "frequency").
				WithDetails("reason", "invalid frequency")
		}
		if err == domain.ErrInvalidStartDate {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "start_date").
				WithDetails("reason", "start date is required")
		}
		if err == domain.ErrEndDateBeforeStartDate {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "end_date").
				WithDetails("reason", "end date cannot be before start date")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// TODO: AI Tagging - Analyze income source and automatically tag
	// Example tags: ["primary", "stable", "taxable", "freelance", "passive"]
	// This would be implemented by AI service that analyzes the source text
	// and income patterns to automatically categorize and tag the income

	// Create income profile in repository
	if err := s.repo.Create(ctx, ip); err != nil {
		s.logger.Error("failed to create income profile",
			zap.String("user_id", userID),
			zap.String("source", req.Source),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile created",
		zap.String("user_id", userID),
		zap.String("profile_id", ip.ID.String()),
		zap.String("source", req.Source),
		zap.String("frequency", req.Frequency))

	return ip, nil
}
