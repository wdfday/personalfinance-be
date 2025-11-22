package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/portfolio_snapshot/domain"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	return r.db.WithContext(ctx).Create(snapshot).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PortfolioSnapshot, error) {
	var snapshot domain.PortfolioSnapshot
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("portfolio snapshot not found")
		}
		return nil, err
	}
	return &snapshot, nil
}

func (r *gormRepository) GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.PortfolioSnapshot, error) {
	var snapshot domain.PortfolioSnapshot
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("portfolio snapshot not found")
		}
		return nil, err
	}
	return &snapshot, nil
}

func (r *gormRepository) GetLatest(ctx context.Context, userID uuid.UUID) (*domain.PortfolioSnapshot, error) {
	var snapshot domain.PortfolioSnapshot
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("snapshot_date DESC").
		First(&snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error if no snapshots exist
		}
		return nil, err
	}
	return &snapshot, nil
}

func (r *gormRepository) GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.PortfolioSnapshot, error) {
	var snapshots []*domain.PortfolioSnapshot
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND snapshot_date >= ? AND snapshot_date <= ?", userID, startDate, endDate).
		Order("snapshot_date ASC").
		Find(&snapshots).Error
	return snapshots, err
}

func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListSnapshotsQuery) ([]*domain.PortfolioSnapshot, int64, error) {
	var snapshots []*domain.PortfolioSnapshot
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PortfolioSnapshot{}).Where("user_id = ?", userID)

	if query.SnapshotType != "" {
		db = db.Where("snapshot_type = ?", query.SnapshotType)
	}
	if query.Period != "" {
		db = db.Where("period = ?", query.Period)
	}
	if query.StartDate != "" {
		db = db.Where("snapshot_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("snapshot_date <= ?", query.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 30
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	err := db.Order("snapshot_date DESC").Limit(pageSize).Offset(offset).Find(&snapshots).Error
	return snapshots, total, err
}

func (r *gormRepository) GetByPeriod(ctx context.Context, userID uuid.UUID, period domain.SnapshotPeriod, limit int) ([]*domain.PortfolioSnapshot, error) {
	var snapshots []*domain.PortfolioSnapshot
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND period = ?", userID, period).
		Order("snapshot_date DESC").
		Limit(limit).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *gormRepository) Update(ctx context.Context, snapshot *domain.PortfolioSnapshot) error {
	return r.db.WithContext(ctx).Save(snapshot).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PortfolioSnapshot{}, id).Error
}

func (r *gormRepository) GetPerformanceMetrics(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*dto.PerformanceMetrics, error) {
	snapshots, err := r.GetByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return &dto.PerformanceMetrics{}, nil
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]

	metrics := &dto.PerformanceMetrics{
		StartDate:  first.SnapshotDate,
		EndDate:    last.SnapshotDate,
		StartValue: first.TotalValue,
		EndValue:   last.TotalValue,
	}

	if first.TotalValue > 0 {
		metrics.TotalReturn = last.TotalValue - first.TotalValue
		metrics.TotalReturnPct = (metrics.TotalReturn / first.TotalValue) * 100
	}

	return metrics, nil
}
