package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// UpdateCategory updates an existing category
func (s *categoryService) UpdateCategory(ctx context.Context, userID string, categoryID string, req dto.UpdateCategoryRequest) (*domain.Category, error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Parse category ID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "category_id").WithDetails("reason", "invalid UUID format")
	}

	// Retrieve existing category and verify ownership
	category, err := s.repo.GetByUserID(ctx, categoryUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Prevent updating system categories
	if category.IsSystemCategory() {
		return nil, shared.ErrForbidden.WithDetails("reason", "cannot update system categories")
	}

	// Convert request to updates using conversion function
	updates, err := dto.ApplyUpdateCategoryRequest(req)
	if err != nil {
		if err == domain.ErrInvalidCategoryType {
			return nil, shared.ErrBadRequest.WithDetails("field", "type").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidParentID {
			return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "invalid UUID format")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Additional validations that require database access
	if req.Name != nil {
		exists, err := s.repo.CheckNameExists(ctx, userUUID, *req.Name, &categoryUUID)
		if err != nil {
			return nil, shared.ErrInternal.WithError(err)
		}
		if exists {
			return nil, shared.ErrConflict.WithDetails("field", "name").WithDetails("reason", "category name already exists")
		}
	}

	if req.ParentID != nil {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "invalid UUID format")
		}

		// Prevent circular reference
		if parentUUID == categoryUUID {
			return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "category cannot be its own parent")
		}

		// Verify parent exists
		parent, err := s.repo.GetByUserID(ctx, parentUUID, userUUID)
		if err != nil {
			if err == shared.ErrNotFound {
				return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "parent category not found")
			}
			return nil, shared.ErrInternal.WithError(err)
		}

		// Calculate and update level
		updates["level"] = parent.Level + 1
	}

	// Apply updates if any
	if len(updates) > 0 {
		if err := s.repo.UpdateColumns(ctx, categoryUUID, updates); err != nil {
			if err == shared.ErrNotFound {
				return nil, err
			}
			return nil, shared.ErrInternal.WithError(err)
		}
	}

	// Retrieve updated category
	updated, err := s.repo.GetByUserID(ctx, categoryUUID, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return updated, nil
}
