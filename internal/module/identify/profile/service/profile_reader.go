package service

import (
	"context"
	"errors"

	"personalfinancedss/internal/module/identify/profile/domain"
	"personalfinancedss/internal/shared"
)

// GetProfile retrieves a user's profile
func (s *profileService) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}
	return profile, nil
}
