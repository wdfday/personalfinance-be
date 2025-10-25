package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(ctx context.Context, userID string, req dto.CreateCategoryRequest) (*domain.Category, error) {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Convert request to entity using conversion function
	category, err := dto.FromCreateCategoryRequest(req, userUUID)
	if err != nil {
		if err == domain.ErrInvalidCategoryType {
			return nil, shared.ErrBadRequest.WithDetails("field", "type").WithDetails("reason", "invalid value")
		}
		if err == domain.ErrInvalidParentID {
			return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "invalid UUID format")
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Check if name already exists for this user
	exists, err := s.repo.CheckNameExists(ctx, userUUID, category.Name, nil)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}
	if exists {
		return nil, shared.ErrConflict.WithDetails("field", "name").WithDetails("reason", "category name already exists")
	}

	// Handle parent category if specified
	if category.ParentID != nil {
		// Verify parent exists and belongs to user
		parent, err := s.repo.GetByUserID(ctx, *category.ParentID, userUUID)
		if err != nil {
			if err == shared.ErrNotFound {
				return nil, shared.ErrBadRequest.WithDetails("field", "parent_id").WithDetails("reason", "parent category not found")
			}
			return nil, shared.ErrInternal.WithError(err)
		}

		// Calculate level based on parent
		category.Level = parent.Level + 1
	}

	// Create category in repository
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Retrieve the created category
	created, err := s.repo.GetByID(ctx, category.ID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return created, nil
}

// InitializeDefaultCategories creates default system categories for a new user
func (s *categoryService) InitializeDefaultCategories(ctx context.Context, userID string) error {
	// Parse and validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Kiểm tra xem user đã có default categories chưa (tránh duplicate)
	existingCategories, err := s.repo.List(ctx, userUUID, dto.ListCategoriesQuery{
		IsRootOnly: true,
	})
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}
	if len(existingCategories) > 0 {
		s.logger.Info("user already has categories, skipping initialization",
			zap.String("user_id", userID),
			zap.Int("existing_count", len(existingCategories)))
		return nil
	}

	// Lấy system categories (user_id = NULL)
	categories, err := s.repo.GetSystemCategories(ctx)
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	if len(categories) == 0 {
		s.logger.Warn("no system categories found to initialize",
			zap.String("user_id", userID))
		return nil
	}

	// Map để lưu old ID -> new ID để xử lý ParentID
	idMap := make(map[uuid.UUID]uuid.UUID, len(categories))
	clones := make([]*domain.Category, 0, len(categories))

	// Pass 1: Clone tất cả categories và tạo ID mapping
	for _, cat := range categories {
		// Copy value ra biến mới
		newCategory := *cat

		// Lưu mapping old ID -> new ID
		oldID := cat.ID
		newID := uuid.New()
		idMap[oldID] = newID

		// Gán ID mới + UserID mới
		newCategory.ID = newID
		newCategory.UserID = &userUUID

		// Reset runtime-only fields
		newCategory.Parent = nil
		newCategory.Children = nil
		newCategory.TransactionCount = 0
		newCategory.TotalAmount = 0

		// ParentID sẽ được update ở pass 2
		clones = append(clones, &newCategory)
	}

	// Pass 2: Update ParentID với new IDs từ mapping
	for _, clone := range clones {
		if clone.ParentID != nil {
			if newParentID, exists := idMap[*clone.ParentID]; exists {
				clone.ParentID = &newParentID
			} else {
				// Parent không tồn tại trong system categories, reset ParentID
				s.logger.Warn("parent category not found in system categories, resetting ParentID",
					zap.String("category_id", clone.ID.String()),
					zap.String("old_parent_id", clone.ParentID.String()))
				clone.ParentID = nil
				clone.Level = 0
			}
		}
	}

	s.logger.Info("initializing default categories for user",
		zap.String("user_id", userID),
		zap.Int("count", len(clones)))

	// Bulk create clones, không phải system categories
	if err := s.repo.BulkCreate(ctx, clones); err != nil {
		s.logger.Error("failed to bulk create default categories",
			zap.String("user_id", userID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("successfully initialized default categories",
		zap.String("user_id", userID),
		zap.Int("count", len(clones)))

	return nil
}
