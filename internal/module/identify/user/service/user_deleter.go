package service

import (
	"context"
	"personalfinancedss/internal/shared"
)

// SoftDelete soft deletes a user
func (s *UserService) SoftDelete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if err == shared.ErrUserNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}
	return nil
}

// HardDelete permanently deletes a user
func (s *UserService) HardDelete(ctx context.Context, id string) error {
	if err := s.repo.HardDelete(ctx, id); err != nil {
		if err == shared.ErrUserNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}
	return nil
}

// Restore restores a soft deleted user
func (s *UserService) Restore(ctx context.Context, id string) error {
	if err := s.repo.Restore(ctx, id); err != nil {
		if err == shared.ErrUserNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}
	return nil
}
