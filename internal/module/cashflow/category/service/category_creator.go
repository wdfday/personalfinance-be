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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.
			WithDetails("field", "user_id").
			WithDetails("reason", "invalid UUID format")
	}

	// Skip if user already has categories
	if hasCategories, err := s.userHasCategories(ctx, userUUID); err != nil {
		return err
	} else if hasCategories {
		s.logger.Info("user already has categories, skipping initialization",
			zap.String("user_id", userID))
		return nil
	}

	// Fetch system categories
	systemCategories, err := s.repo.GetSystemCategories(ctx)
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	if len(systemCategories) == 0 {
		s.logger.Warn("no system categories found to initialize",
			zap.String("user_id", userID))
		return nil
	}

	// Clone categories for user with updated IDs and parent references
	userCategories := s.cloneCategoriesForUser(systemCategories, userUUID)

	s.logger.Info("initializing default categories for user",
		zap.String("user_id", userID),
		zap.Int("count", len(userCategories)))

	// Bulk create user categories
	if err := s.repo.BulkCreate(ctx, userCategories); err != nil {
		s.logger.Error("failed to bulk create default categories",
			zap.String("user_id", userID),
			zap.Error(err))
		return shared.ErrInternal.WithError(err)
	}

	s.logger.Info("successfully initialized default categories",
		zap.String("user_id", userID),
		zap.Int("count", len(userCategories)))

	return nil
}

// userHasCategories checks if user already has any categories
func (s *categoryService) userHasCategories(ctx context.Context, userID uuid.UUID) (bool, error) {
	categories, err := s.repo.List(ctx, userID, dto.ListCategoriesQuery{
		IsRootOnly: true,
	})
	if err != nil {
		return false, shared.ErrInternal.WithError(err)
	}
	return len(categories) > 0, nil
}

// cloneCategoriesForUser creates copies of system categories for a specific user
// It maintains the hierarchical structure by remapping parent IDs
func (s *categoryService) cloneCategoriesForUser(systemCategories []*domain.Category, userID uuid.UUID) []*domain.Category {
	idMap := make(map[uuid.UUID]uuid.UUID, len(systemCategories))
	userCategories := make([]*domain.Category, 0, len(systemCategories))

	// First pass: clone categories and create ID mapping
	for _, systemCat := range systemCategories {
		newCat := s.cloneSingleCategory(systemCat, userID)
		idMap[systemCat.ID] = newCat.ID
		userCategories = append(userCategories, newCat)
	}

	// Second pass: remap parent IDs to new user-specific IDs
	for _, userCat := range userCategories {
		if userCat.ParentID != nil {
			if newParentID, exists := idMap[*userCat.ParentID]; exists {
				userCat.ParentID = &newParentID
			} else {
				// Orphaned category - reset to root level
				s.logger.Warn("parent category not found, resetting to root",
					zap.String("category_id", userCat.ID.String()))
				userCat.ParentID = nil
				userCat.Level = 0
			}
		}
	}

	return userCategories
}

// cloneSingleCategory creates a copy of a category for a specific user
func (s *categoryService) cloneSingleCategory(systemCat *domain.Category, userID uuid.UUID) *domain.Category {
	return &domain.Category{
		ID:               uuid.New(),
		UserID:           &userID,
		Name:             systemCat.Name,
		Description:      systemCat.Description,
		Type:             systemCat.Type,
		Icon:             systemCat.Icon,
		Color:            systemCat.Color,
		IsDefault:        false, // User copy is not a system default
		IsActive:         true,
		Level:            systemCat.Level,
		DisplayOrder:     systemCat.DisplayOrder,
		ParentID:         systemCat.ParentID, // Will be remapped in second pass
		Parent:           nil,
		Children:         nil,
		TransactionCount: 0,
		TotalAmount:      0,
	}
}
