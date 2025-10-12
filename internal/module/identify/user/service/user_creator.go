package service

import (
	"context"
	"personalfinancedss/internal/module/identify/user/domain"
	"personalfinancedss/internal/shared"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Create creates a new user
func (s *UserService) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Normalize email
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	// Check if email already exists
	if _, err := s.repo.GetByEmail(ctx, user.Email); err == nil {
		return nil, shared.ErrConflict.WithDetails("field", "email")
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.LastActiveAt = now

	// Create user
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create default profile for the new user
	// This ensures every user has a profile with sensible defaults
	if s.profileService != nil {
		if _, err := s.profileService.CreateDefaultProfile(ctx, user.ID.String()); err != nil {
			// Log error but don't fail user creation
			// Profile can be created later if needed
			s.logger.Warn(
				"Failed to create default profile for user",
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
		}
	}

	if s.categoryService != nil {
		if err := s.categoryService.InitializeDefaultCategories(ctx, user.ID.String()); err != nil {
			// Log error but don't fail user creation
			// Categories can be initialized later if needed
			s.logger.Warn(
				"Failed to initialize default categories for user",
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
		}
	}

	if s.accountService != nil {
		if err := s.accountService.CreateDefaultCashAccount(ctx, user.ID.String()); err != nil {
			// Log error but don't fail user creation
			// Account can be created later if needed
			s.logger.Warn(
				"Failed to create default cash account for user",
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
		}

	}

	return user, nil
}
