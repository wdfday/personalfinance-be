package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"go.uber.org/zap"
)

// UpdateIncomeProfile creates a NEW version and archives the old one (versioning pattern)
func (s *incomeProfileService) UpdateIncomeProfile(ctx context.Context, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	// Parse IDs
	profileUUID, err := parseProfileID(profileID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing income profile
	existing, err := s.repo.GetByID(ctx, profileUUID)
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
	if existing.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	// Check if already archived
	if existing.IsArchived() {
		return nil, shared.ErrBadRequest.
			WithDetails("reason", "cannot update archived income profile")
	}

	// Create new version with updates applied
	newVersion, err := dto.ApplyUpdateIncomeProfileRequest(req, existing)
	if err != nil {
		if err == domain.ErrNegativeAmount {
			return nil, shared.ErrBadRequest.
				WithDetails("field", "amount").
				WithDetails("reason", "amount cannot be negative")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Archive the old version
	existing.Archive(userUUID)
	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to archive old version",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create the new version
	if err := s.repo.Create(ctx, newVersion); err != nil {
		s.logger.Error("failed to create new version",
			zap.String("user_id", userID),
			zap.String("old_profile_id", profileID),
			zap.String("new_profile_id", newVersion.ID.String()),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile updated (new version created)",
		zap.String("user_id", userID),
		zap.String("old_profile_id", profileID),
		zap.String("new_profile_id", newVersion.ID.String()))

	return newVersion, nil
}

// UpdateDSSMetadata updates DSS analysis metadata
func (s *incomeProfileService) UpdateDSSMetadata(ctx context.Context, userID string, profileID string, req dto.UpdateDSSMetadataRequest) (*domain.IncomeProfile, error) {
	// Parse IDs
	profileUUID, err := parseProfileID(profileID)
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
	userUUID, _ := parseUserID(userID)
	if ip.UserID != userUUID {
		return nil, shared.ErrForbidden.
			WithDetails("reason", "income profile does not belong to user")
	}

	// Convert request to metadata map
	metadata := dto.FromUpdateDSSMetadataRequest(req)

	// Update DSS metadata
	if err := ip.UpdateDSSMetadata(metadata); err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, ip); err != nil {
		s.logger.Error("failed to update DSS metadata",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("DSS metadata updated",
		zap.String("user_id", userID),
		zap.String("profile_id", profileID))

	return ip, nil
}

// ArchiveIncomeProfile manually archives an income profile
func (s *incomeProfileService) ArchiveIncomeProfile(ctx context.Context, userID string, profileID string) error {
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

	// Check if already archived
	if ip.IsArchived() {
		return shared.ErrBadRequest.
			WithDetails("reason", "income profile is already archived")
	}

	// Archive in repository
	if err := s.repo.Archive(ctx, profileUUID, userUUID); err != nil {
		s.logger.Error("failed to archive income profile",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile archived",
		zap.String("user_id", userID),
		zap.String("profile_id", profileID))

	return nil
}

// EndIncomeProfile marks an income profile as ended
func (s *incomeProfileService) EndIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error) {
	// Parse and validate IDs
	profileUUID, err := parseProfileID(profileID)
	if err != nil {
		return nil, shared.ErrBadRequest.
			WithDetails("field", "profile_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get existing income profile to verify ownership
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

	// Check if already ended or archived
	if ip.Status == domain.IncomeStatusEnded || ip.Status == domain.IncomeStatusArchived {
		return nil, shared.ErrBadRequest.
			WithDetails("reason", "income profile is already ended or archived")
	}

	// Mark as ended
	ip.MarkAsEnded()

	// Update in repository
	if err := s.repo.Update(ctx, ip); err != nil {
		s.logger.Error("failed to end income profile",
			zap.String("user_id", userID),
			zap.String("profile_id", profileID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("income profile marked as ended",
		zap.String("user_id", userID),
		zap.String("profile_id", profileID))

	return ip, nil
}

// CheckAndArchiveEnded checks and archives ended income profiles automatically
func (s *incomeProfileService) CheckAndArchiveEnded(ctx context.Context, userID string) (int, error) {
	// Parse and validate user ID
	userUUID, err := parseUserID(userID)
	if err != nil {
		return 0, shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Get all active income profiles
	profiles, err := s.repo.GetActiveByUser(ctx, userUUID)
	if err != nil {
		return 0, shared.ErrInternal.WithError(err)
	}

	endedCount := 0

	// Check each profile and mark as ended if end_date has passed
	for _, ip := range profiles {
		if ip.CheckAndMarkAsEnded() {
			if err := s.repo.Update(ctx, ip); err != nil {
				s.logger.Error("failed to mark ended income profile",
					zap.String("user_id", userID),
					zap.String("profile_id", ip.ID.String()),
					zap.Error(err))
				continue // Continue with other profiles
			}
			endedCount++
		}
	}

	if endedCount > 0 {
		s.logger.Info("marked ended income profiles",
			zap.String("user_id", userID),
			zap.Int("count", endedCount))
	}

	return endedCount, nil
}
