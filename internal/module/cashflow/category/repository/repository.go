package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"

	"github.com/google/uuid"
)

// Repository defines category data access operations
type Repository interface {
	// Create creates a new category
	Create(ctx context.Context, category *domain.Category) error

	// GetByID retrieves a category by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)

	// GetByUserID retrieves a category by ID and user ID (or system category)
	GetByUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Category, error)

	// List retrieves categories with filters
	List(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error)

	// ListWithChildren retrieves categories with their children (hierarchical)
	ListWithChildren(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error)

	// GetChildren retrieves child categories of a parent
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Category, error)

	// GetByType retrieves categories by type
	GetByType(ctx context.Context, userID uuid.UUID, categoryType domain.CategoryType) ([]*domain.Category, error)

	// Update updates a category
	Update(ctx context.Context, category *domain.Category) error

	// UpdateColumns updates specific columns of a category
	UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]any) error

	// Delete soft deletes a category
	Delete(ctx context.Context, id uuid.UUID) error

	// GetCategoryStats retrieves transaction statistics for a category
	GetCategoryStats(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID) (*domain.Category, error)

	// CheckNameExists checks if a category name already exists for a user
	CheckNameExists(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)

	// GetSystemCategories retrieves all system-provided categories
	GetSystemCategories(ctx context.Context) ([]*domain.Category, error)

	// BulkCreate creates multiple categories at once (for default categories)
	BulkCreate(ctx context.Context, categories []*domain.Category) error
}
