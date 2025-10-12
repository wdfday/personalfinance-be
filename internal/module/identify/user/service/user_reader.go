package service

import (
	"context"
	"personalfinancedss/internal/module/identify/user/domain"
	"personalfinancedss/internal/shared"
	"strings"
)

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == shared.ErrUserNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == shared.ErrUserNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}
	return user, nil
}

// List retrieves user with pagination
func (s *UserService) List(ctx context.Context, filter domain.ListUsersFilter, pagination shared.Pagination) (shared.Page[domain.User], error) {
	page, err := s.repo.List(ctx, filter, pagination)
	if err != nil {
		return shared.Page[domain.User]{}, shared.ErrInternal.WithError(err)
	}
	return page, nil
}
