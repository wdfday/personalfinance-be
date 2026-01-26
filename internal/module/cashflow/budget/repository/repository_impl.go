package repository

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// New creates a new budget repository
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, budget *domain.Budget) error {
	if err := r.db.WithContext(ctx).Create(budget).Error; err != nil {
		return err
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	var budget domain.Budget
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&budget).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return &budget, nil
}

// FindByIDAndUserID retrieves a budget by ID and verifies user ownership
func (r *repository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error) {
	var budget domain.Budget
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&budget).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return &budget, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN (?)", userID, []string{
			string(domain.BudgetStatusActive),
			string(domain.BudgetStatusWarning),
		}).
		Order("created_at DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) FindByUserIDAndCategory(ctx context.Context, userID, categoryID uuid.UUID) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND category_id = ?", userID, categoryID).
		Order("created_at DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) FindByConstraintID(ctx context.Context, userID, constraintID uuid.UUID) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND constraint_id = ?", userID, constraintID).
		Order("created_at DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) FindByPeriod(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]domain.Budget, error) {
	var budgets []domain.Budget
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND start_date >= ? AND (end_date IS NULL OR end_date <= ?)", userID, startDate, endDate).
		Order("start_date DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) Update(ctx context.Context, budget *domain.Budget) error {
	return r.db.WithContext(ctx).Save(budget).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Budget{}, id).Error
}

func (r *repository) UpdateSpentAmount(ctx context.Context, id uuid.UUID, spentAmount float64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Budget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"spent_amount":       spentAmount,
			"remaining_amount":   gorm.Expr("amount - ?", spentAmount),
			"percentage_spent":   gorm.Expr("(? / amount) * 100", spentAmount),
			"last_calculated_at": time.Now(),
		}).Error
}

func (r *repository) FindExpiredBudgets(ctx context.Context) ([]domain.Budget, error) {
	var budgets []domain.Budget
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("end_date IS NOT NULL AND end_date < ? AND status != ?", now, domain.BudgetStatusExpired).
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) FindBudgetsNeedingRecalculation(ctx context.Context, threshold time.Duration) ([]domain.Budget, error) {
	var budgets []domain.Budget
	cutoffTime := time.Now().Add(-threshold)
	err := r.db.WithContext(ctx).
		Where("last_calculated_at IS NULL OR last_calculated_at < ?", cutoffTime).
		Where("status IN (?)", []string{
			string(domain.BudgetStatusActive),
			string(domain.BudgetStatusWarning),
		}).
		Find(&budgets).Error
	return budgets, err
}

// FindByUserIDPaginated retrieves budgets for a user with pagination
func (r *repository) FindByUserIDPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult, error) {
	var budgets []domain.Budget
	var total int64

	// Set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&domain.Budget{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated data
	offset := (params.Page - 1) * params.PageSize
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&budgets).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Data:       budgets,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteByIDAndUserID deletes a budget with ownership verification
func (r *repository) DeleteByIDAndUserID(ctx context.Context, id, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Budget{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrBudgetNotFound
	}

	return nil
}

// ExistsByUserIDAndName checks if a budget with the same name exists for user
func (r *repository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Budget{}).
		Where("user_id = ? AND name = ?", userID, name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
