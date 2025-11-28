package repository

import (
	"context"
	"errors"
	"time"

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

// GetByUser retrieves all active budget constraints for a user (not archived)
func (r *gormRepository) GetByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND archived_at IS NULL", userID).
		Order("priority ASC, minimum_amount DESC").
		Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

// GetActiveByUser retrieves all currently active budget constraints
func (r *gormRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			userID, domain.ConstraintStatusActive, now, now).
		Order("priority ASC").
		Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

// GetArchivedByUser retrieves all archived budget constraints
func (r *gormRepository) GetArchivedByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND archived_at IS NOT NULL", userID).
		Order("archived_at DESC").
		Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

// GetByStatus retrieves constraints by user and status
func (r *gormRepository) GetByStatus(ctx context.Context, userID uuid.UUID, status domain.ConstraintStatus) (domain.BudgetConstraints, error) {
	var constraints domain.BudgetConstraints
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("priority ASC").
		Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

// GetVersionHistory retrieves all versions of a constraint chain
func (r *gormRepository) GetVersionHistory(ctx context.Context, constraintID uuid.UUID) (domain.BudgetConstraints, error) {
	var versions domain.BudgetConstraints

	// Start with the given constraint
	current, err := r.GetByID(ctx, constraintID)
	if err != nil {
		return nil, err
	}

	// Traverse backwards to find all previous versions
	for current.PreviousVersionID != nil {
		var prev domain.BudgetConstraint
		if err := r.db.WithContext(ctx).
			Where("id = ?", *current.PreviousVersionID).
			First(&prev).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, err
		}
		versions = append(versions, &prev)
		current = &prev
	}

	// Also find any newer versions
	var newerVersions domain.BudgetConstraints
	tempID := constraintID
	for {
		var next domain.BudgetConstraint
		if err := r.db.WithContext(ctx).
			Where("previous_version_id = ?", tempID).
			First(&next).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, err
		}
		newerVersions = append(domain.BudgetConstraints{&next}, newerVersions...)
		tempID = next.ID
	}

	result := append(newerVersions, versions...)
	return result, nil
}

// GetLatestVersion retrieves the latest version of a constraint chain
func (r *gormRepository) GetLatestVersion(ctx context.Context, constraintID uuid.UUID) (*domain.BudgetConstraint, error) {
	current, err := r.GetByID(ctx, constraintID)
	if err != nil {
		return nil, err
	}

	// Find the latest version (no other constraint has this as previous_version_id)
	for {
		var next domain.BudgetConstraint
		if err := r.db.WithContext(ctx).
			Where("previous_version_id = ?", current.ID).
			First(&next).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return current, nil
			}
			return nil, err
		}
		current = &next
	}
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

// Archive archives a budget constraint
func (r *gormRepository) Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&domain.BudgetConstraint{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      domain.ConstraintStatusArchived,
			"archived_at": now,
			"archived_by": archivedBy,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}
	return nil
}

// GetTotalMandatory calculates total mandatory expenses for user
func (r *gormRepository) GetTotalMandatory(ctx context.Context, userID uuid.UUID) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).
		Model(&domain.BudgetConstraint{}).
		Where("user_id = ? AND archived_at IS NULL", userID).
		Select("COALESCE(SUM(minimum_amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
