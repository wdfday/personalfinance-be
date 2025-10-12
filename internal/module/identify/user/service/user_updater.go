package service

import (
	"context"
	"personalfinancedss/internal/module/identify/user/domain"
	"personalfinancedss/internal/shared"
	"time"
)

// Update updates a user
func (s *UserService) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		if err == shared.ErrUserNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}
	return nil
}

// UpdateColumns performs partial update
func (s *UserService) UpdateColumns(ctx context.Context, id string, cols map[string]any) error {
	if err := s.repo.UpdateColumns(ctx, id, cols); err != nil {
		if err == shared.ErrUserNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}
	return nil
}

// UpdatePassword updates user password
func (s *UserService) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	cols := map[string]any{
		"password":            passwordHash,
		"password_changed_at": time.Now(),
	}
	return s.UpdateColumns(ctx, id, cols)
}

// UpdateLastLogin updates last login info
func (s *UserService) UpdateLastLogin(ctx context.Context, id string, at time.Time, ip *string) error {
	return s.repo.UpdateLastLogin(ctx, id, at, ip)
}
