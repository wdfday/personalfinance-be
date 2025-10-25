package repository

import (
	"context"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based category repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new category
func (r *gormRepository) Create(ctx context.Context, category *domain.Category) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a category by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

// GetByUserID retrieves a category by ID and user ID (or system category)
func (r *gormRepository) GetByUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).
		Where("id = ? AND (user_id = ? OR user_id IS NULL)", id, userID).
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

// List retrieves categories with filters
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	var categories []*domain.Category

	db := r.db.WithContext(ctx).
		Where("user_id = ?", userID) // User's categories

	// Apply filters
	db = r.applyFilters(db, query)

	// Order by display order and name
	if err := db.Order("display_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// ListWithChildren retrieves categories with their children (hierarchical)
func (r *gormRepository) ListWithChildren(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	var categories []*domain.Category

	db := r.db.WithContext(ctx).
		Where("user_id = ? OR user_id IS NULL", userID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("display_order ASC, name ASC")
		})

	// Apply filters
	db = r.applyFilters(db, query)

	// Order by display order and name
	if err := db.Order("display_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// applyFilters applies query filters to the database query
func (r *gormRepository) applyFilters(db *gorm.DB, query dto.ListCategoriesQuery) *gorm.DB {
	if query.Type != nil {
		db = db.Where("type = ? OR type = ?", *query.Type, domain.CategoryTypeBoth)
	}

	if query.ParentID != nil {
		parentUUID, err := uuid.Parse(*query.ParentID)
		if err == nil {
			db = db.Where("parent_id = ?", parentUUID)
		}
	}

	if query.IsRootOnly {
		db = db.Where("parent_id IS NULL")
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	return db
}

// GetChildren retrieves child categories of a parent
func (r *gormRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Category, error) {
	var categories []*domain.Category

	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("display_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// GetByType retrieves categories by type
func (r *gormRepository) GetByType(ctx context.Context, userID uuid.UUID, categoryType domain.CategoryType) ([]*domain.Category, error) {
	var categories []*domain.Category

	if err := r.db.WithContext(ctx).
		Where("(user_id = ? OR user_id IS NULL) AND (type = ? OR type = ?)",
			userID, categoryType, domain.CategoryTypeBoth).
		Where("is_active = ?", true).
		Order("display_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// Update updates a category
func (r *gormRepository) Update(ctx context.Context, category *domain.Category) error {
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		return err
	}
	return nil
}

// UpdateColumns updates specific columns of a category
func (r *gormRepository) UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]any) error {
	result := r.db.WithContext(ctx).Model(&domain.Category{}).Where("id = ?", id).Updates(columns)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// Delete soft deletes a category
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetCategoryStats retrieves transaction statistics for a category
func (r *gormRepository) GetCategoryStats(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID) (*domain.Category, error) {
	var category domain.Category

	// First get the category
	if err := r.db.WithContext(ctx).
		Where("id = ? AND (user_id = ? OR user_id IS NULL)", categoryID, userID).
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	// Calculate statistics from transactions
	type stats struct {
		Count int
		Total float64
	}
	var result stats

	if err := r.db.WithContext(ctx).
		Table("transactions").
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Where("category_id = ? AND user_id = ? AND status = ? AND deleted_at IS NULL",
			categoryID, userID, "completed").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	category.TransactionCount = result.Count
	category.TotalAmount = result.Total

	return &category, nil
}

// CheckNameExists checks if a category name already exists for a user
func (r *gormRepository) CheckNameExists(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&domain.Category{}).
		Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, name)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetSystemCategories retrieves all system-provided categories
func (r *gormRepository) GetSystemCategories(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category

	systemUUID := uuid.Nil // 00000000-0000-0000-0000-000000000000

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", systemUUID, true).
		Order("type ASC, display_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// BulkCreate creates multiple categories at once
func (r *gormRepository) BulkCreate(ctx context.Context, categories []*domain.Category) error {
	if len(categories) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Create(&categories).Error; err != nil {
		return err
	}

	return nil
}
