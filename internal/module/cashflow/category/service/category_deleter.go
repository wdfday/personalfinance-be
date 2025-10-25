package service

import (
	"context"

	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// DeleteCategory soft deletes a category
func (s *categoryService) DeleteCategory(ctx context.Context, userID string, categoryID string) error {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Parse category ID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "category_id").WithDetails("reason", "invalid UUID format")
	}

	// Verify category exists and belongs to user
	category, err := s.repo.GetByUserID(ctx, categoryUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}

	// Prevent deleting system categories
	if category.IsSystemCategory() {
		return shared.ErrForbidden.WithDetails("reason", "cannot delete system categories")
	}

	// Check if category has children
	children, err := s.repo.GetChildren(ctx, categoryUUID)
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	if len(children) > 0 {
		return shared.ErrBadRequest.WithDetails("reason", "cannot delete category with sub-categories")
	}

	// Delete category
	if err := s.repo.Delete(ctx, categoryUUID); err != nil {
		if err == shared.ErrNotFound {
			return err
		}
		return shared.ErrInternal.WithError(err)
	}

	return nil
}
