package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetCategory retrieves a single category by ID
func (s *categoryService) GetCategory(ctx context.Context, userID string, categoryID string) (*domain.Category, error) {
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

	// Retrieve category (includes system categories)
	category, err := s.repo.GetByUserID(ctx, categoryUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return category, nil
}

// ListCategories retrieves a list of categories with filters
func (s *categoryService) ListCategories(ctx context.Context, userID string, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("invalid user ID format in ListCategories",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	s.logger.Info("listing categories for user",
		zap.String("user_id", userID),
		zap.Any("query", query))

	// Retrieve categories from repository
	categories, err := s.repo.List(ctx, userUUID, query)
	if err != nil {
		s.logger.Error("failed to list categories from repository",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Debug("retrieved categories from repository",
		zap.String("user_id", userID),
		zap.Int("count", len(categories)))

	// If stats are requested, enrich each category with statistics
	if query.IncludeStats {
		s.logger.Debug("enriching categories with statistics",
			zap.String("user_id", userID),
			zap.Int("count", len(categories)))

		for i, cat := range categories {
			statsCategory, err := s.repo.GetCategoryStats(ctx, cat.ID, userUUID)
			if err != nil {
				s.logger.Warn("failed to get stats for category",
					zap.String("user_id", userID),
					zap.String("category_id", cat.ID.String()),
					zap.String("category_name", cat.Name),
					zap.Error(err))
				// Continue with next category, stats will remain 0
				continue
			}

			if statsCategory != nil {
				categories[i].TransactionCount = statsCategory.TransactionCount
				categories[i].TotalAmount = statsCategory.TotalAmount
			}
		}
	}

	s.logger.Info("successfully listed categories for user",
		zap.String("user_id", userID),
		zap.Int("count", len(categories)),
		zap.Bool("include_stats", query.IncludeStats))

	return categories, nil
}

// ListCategoriesWithChildren retrieves categories with their children (hierarchical)
func (s *categoryService) ListCategoriesWithChildren(ctx context.Context, userID string, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Retrieve categories with children from repository
	categories, err := s.repo.ListWithChildren(ctx, userUUID, query)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// If stats are requested, enrich each category and its children with statistics
	if query.IncludeStats {
		for i, cat := range categories {
			// Get stats for parent
			statsCategory, err := s.repo.GetCategoryStats(ctx, cat.ID, userUUID)
			if err == nil && statsCategory != nil {
				categories[i].TransactionCount = statsCategory.TransactionCount
				categories[i].TotalAmount = statsCategory.TotalAmount
			}

			// Get stats for children
			if len(cat.Children) > 0 {
				for j, child := range cat.Children {
					childStats, err := s.repo.GetCategoryStats(ctx, child.ID, userUUID)
					if err == nil && childStats != nil {
						categories[i].Children[j].TransactionCount = childStats.TransactionCount
						categories[i].Children[j].TotalAmount = childStats.TotalAmount
					}
				}
			}
		}
	}

	return categories, nil
}

// GetCategoryStats retrieves transaction statistics for a category
func (s *categoryService) GetCategoryStats(ctx context.Context, userID string, categoryID string) (*domain.Category, error) {
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

	// Get category with stats
	category, err := s.repo.GetCategoryStats(ctx, categoryUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return category, nil
}

func (s *categoryService) GetSystemCategories(ctx context.Context) ([]*domain.Category, error) {
	// Retrieve system categories from repository
	categories, err := s.repo.GetSystemCategories(ctx)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return categories, nil
}
