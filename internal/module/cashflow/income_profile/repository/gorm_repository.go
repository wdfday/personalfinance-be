package repository

import (
	"context"
	"errors"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
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

// Create creates a new income profile
func (r *gormRepository) Create(ctx context.Context, ip *domain.IncomeProfile) error {
	if err := r.db.WithContext(ctx).Create(ip).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves an income profile by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IncomeProfile, error) {
	var ip domain.IncomeProfile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ip).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &ip, nil
}

// GetByUserAndPeriod retrieves an income profile by user and period
func (r *gormRepository) GetByUserAndPeriod(ctx context.Context, userID uuid.UUID, year, month int) (*domain.IncomeProfile, error) {
	var ip domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND year = ? AND month = ?", userID, year, month).
		First(&ip).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &ip, nil
}

// GetByUser retrieves all income profiles for a user
func (r *gormRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("year DESC, month DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetByUserAndYear retrieves all income profiles for a user in a specific year
func (r *gormRepository) GetByUserAndYear(ctx context.Context, userID uuid.UUID, year int) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND year = ?", userID, year).
		Order("month DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// List retrieves income profiles with filters
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile

	db := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// Apply filters
	if query.Year != nil {
		db = db.Where("year = ?", *query.Year)
	}
	if query.IsActual != nil {
		db = db.Where("is_actual = ?", *query.IsActual)
	}

	// Order by year and month descending
	db = db.Order("year DESC, month DESC")

	if err := db.Find(&profiles).Error; err != nil {
		return nil, err
	}

	return profiles, nil
}

// Update updates an existing income profile
func (r *gormRepository) Update(ctx context.Context, ip *domain.IncomeProfile) error {
	if err := r.db.WithContext(ctx).Save(ip).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes an income profile
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.IncomeProfile{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// Exists checks if an income profile exists for user and period
func (r *gormRepository) Exists(ctx context.Context, userID uuid.UUID, year, month int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.IncomeProfile{}).
		Where("user_id = ? AND year = ? AND month = ?", userID, year, month).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
