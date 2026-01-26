package repository

import (
	"context"
	"errors"
	"time"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormRepository implements Repository using GORM with versioning support
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

// GetByUser retrieves all active income profiles for a user (not archived)
func (r *gormRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND archived_at IS NULL", userID).
		Order("start_date DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetActiveByUser retrieves all currently active income profiles for a user
func (r *gormRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			userID, domain.IncomeStatusActive, now, now).
		Order("start_date DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetArchivedByUser retrieves all archived income profiles for a user
func (r *gormRepository) GetArchivedByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND archived_at IS NOT NULL", userID).
		Order("archived_at DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetByStatus retrieves income profiles by user and status
func (r *gormRepository) GetByStatus(ctx context.Context, userID uuid.UUID, status domain.IncomeStatus) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("start_date DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetVersionHistory retrieves all versions of an income profile chain
func (r *gormRepository) GetVersionHistory(ctx context.Context, profileID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var versions []*domain.IncomeProfile

	// Start with the given profile
	current, err := r.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// Traverse backwards to find all previous versions
	for current.PreviousVersionID != nil {
		var prev domain.IncomeProfile
		if err := r.db.WithContext(ctx).
			Where("id = ?", *current.PreviousVersionID).
			First(&prev).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // Stop if previous version not found
			}
			return nil, err
		}
		versions = append(versions, &prev)
		current = &prev
	}

	// Also find any newer versions (where previous_version_id = profileID)
	var newerVersions []*domain.IncomeProfile
	tempID := profileID
	for {
		var next domain.IncomeProfile
		if err := r.db.WithContext(ctx).
			Where("previous_version_id = ?", tempID).
			First(&next).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // No more newer versions
			}
			return nil, err
		}
		newerVersions = append([]*domain.IncomeProfile{&next}, newerVersions...)
		tempID = next.ID
	}

	// Combine: newer versions + original profile + older versions
	result := append(newerVersions, versions...)
	return result, nil
}

// GetLatestVersion retrieves the latest version of an income profile chain
func (r *gormRepository) GetLatestVersion(ctx context.Context, profileID uuid.UUID) (*domain.IncomeProfile, error) {
	current, err := r.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// Find the latest version (no other profile has this as previous_version_id)
	for {
		var next domain.IncomeProfile
		if err := r.db.WithContext(ctx).
			Where("previous_version_id = ?", current.ID).
			First(&next).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No newer version found, current is the latest
				return current, nil
			}
			return nil, err
		}
		current = &next
	}
}

// GetBySource retrieves income profiles by user and source
func (r *gormRepository) GetBySource(ctx context.Context, userID uuid.UUID, source string) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND source = ?", userID, source).
		Order("start_date DESC").
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetRecurringByUser retrieves all recurring income profiles for a user
func (r *gormRepository) GetRecurringByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	var profiles []*domain.IncomeProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_recurring = ? AND archived_at IS NULL", userID, true).
		Order("start_date DESC").
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
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.IsRecurring != nil {
		db = db.Where("is_recurring = ?", *query.IsRecurring)
	}
	if query.Source != nil {
		db = db.Where("source = ?", *query.Source)
	}

	// Include archived filter
	if !query.IncludeArchived {
		db = db.Where("archived_at IS NULL")
	}

	// Order by start date descending
	db = db.Order("start_date DESC")

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

// Delete soft deletes an income profile
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.IncomeProfile{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// Archive archives an income profile
func (r *gormRepository) Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&domain.IncomeProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      domain.IncomeStatusArchived,
			"archived_at": now,
			"archived_by": archivedBy,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}
	return nil
}
