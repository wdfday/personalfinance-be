package repository

import (
	"context"
	"errors"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormRepository implements Repository using GORM
type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new budget constraint
func (r *gormRepository) Create(ctx context.Context, bc *domain.BudgetConstraint) error {
	if err := r.db.WithContext(ctx).Create(bc).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a budget constraint by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetConstraint, error) {
	var bc domain.BudgetConstraint
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&bc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &bc, nil
}

// GetByUser retrieves all budget constraints for a user
func (r *gormRepository) GetByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("priority ASC, minimum_amount DESC").
		Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

// GetByUserAndCategory retrieves a budget constraint by user and category
func (r *gormRepository) GetByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*domain.BudgetConstraint, error) {
	var bc domain.BudgetConstraint
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		First(&bc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &bc, nil
}

// List retrieves budget constraints with filters
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints

	db := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// Apply filters
	if query.CategoryID != nil {
		categoryID, err := uuid.Parse(*query.CategoryID)
		if err == nil {
			db = db.Where("category_id = ?", categoryID)
		}
	}
	if query.IsFlexible != nil {
		db = db.Where("is_flexible = ?", *query.IsFlexible)
	}

	// Order by priority and minimum amount
	db = db.Order("priority ASC, minimum_amount DESC")

	if err := db.Find(&constraints).Error; err != nil {
		return nil, err
	}

	return constraints, nil
}

// Update updates an existing budget constraint
func (r *gormRepository) Update(ctx context.Context, bc *domain.BudgetConstraint) error {
	if err := r.db.WithContext(ctx).Save(bc).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a budget constraint
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.BudgetConstraint{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// DeleteByUserAndCategory deletes a budget constraint by user and category
func (r *gormRepository) DeleteByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Delete(&domain.BudgetConstraint{}, "user_id = ? AND category_id = ?", userID, categoryID).Error; err != nil {
		return err
	}
	return nil
}

// Exists checks if a budget constraint exists for user and category
func (r *gormRepository) Exists(ctx context.Context, userID, categoryID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.BudgetConstraint{}).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTotalMandatory calculates total mandatory expenses for user
func (r *gormRepository) GetTotalMandatory(ctx context.Context, userID uuid.UUID) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).
		Model(&domain.BudgetConstraint{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(minimum_amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
