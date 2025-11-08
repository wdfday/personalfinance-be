package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateIncomeProfile creates a new income profile
func (s *incomeProfileService) CreateIncomeProfile(ctx context.Context, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Check if income profile already exists for this period
	exists, err := s.repo.Exists(ctx, userUUID, req.Year, req.Month)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}
	if exists {
		return nil, shared.ErrConflict.
			WithDetails("field", "period").
			WithDetails("reason", "income profile already exists for this period")
	}

	// Convert request to entity
	ip, err := dto.FromCreateIncomeProfileRequest(req, userUUID)
	if err != nil {
		// Check for specific domain errors
		if err == domain.ErrInvalidYear {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "year").
				WithDetails("reason", "invalid year")
		}
		if err == domain.ErrInvalidMonth {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "month").
				WithDetails("reason", "invalid month")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create income profile in repository
	if err := s.repo.Create(ctx, ip); err != nil {
		s.logger.Error("failed to create income profile",
			zap.String("user_id", userID),
			zap.Int("year", req.Year),
			zap.Int("month", req.Month),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile created",
		zap.String("user_id", userID),
		zap.String("profile_id", ip.ID.String()),
		zap.Int("year", req.Year),
		zap.Int("month", req.Month))

	return ip, nil
}
