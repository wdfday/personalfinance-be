package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/investment_transaction/domain"
	"personalfinancedss/internal/module/investment/investment_transaction/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based investment transaction repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new investment transaction
func (r *gormRepository) Create(ctx context.Context, transaction *domain.InvestmentTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetByID retrieves an investment transaction by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentTransaction, error) {
	var transaction domain.InvestmentTransaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("investment transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

// GetByUserID retrieves an investment transaction by ID and user ID
func (r *gormRepository) GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.InvestmentTransaction, error) {
	var transaction domain.InvestmentTransaction
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("investment transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

// List retrieves investment transactions with filters and pagination
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) ([]*domain.InvestmentTransaction, int64, error) {
	var transactions []*domain.InvestmentTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.InvestmentTransaction{}).Where("user_id = ?", userID)

	// Apply filters
	if query.AssetID != "" {
		assetID, err := uuid.Parse(query.AssetID)
		if err == nil {
			db = db.Where("asset_id = ?", assetID)
		}
	}
	if query.TransactionType != "" {
		db = db.Where("transaction_type = ?", query.TransactionType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartDate != "" {
		db = db.Where("transaction_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("transaction_date <= ?", query.EndDate)
	}
	if query.MinAmount > 0 {
		db = db.Where("total_amount >= ?", query.MinAmount)
	}
	if query.MaxAmount > 0 {
		db = db.Where("total_amount <= ?", query.MaxAmount)
	}
	if query.Broker != "" {
		db = db.Where("broker = ?", query.Broker)
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "transaction_date"
	}
	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	if err := db.Limit(pageSize).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByAssetID retrieves all transactions for a specific asset
func (r *gormRepository) GetByAssetID(ctx context.Context, userID, assetID uuid.UUID) ([]*domain.InvestmentTransaction, error) {
	var transactions []*domain.InvestmentTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND asset_id = ?", userID, assetID).
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}

// GetByDateRange retrieves transactions within a date range
func (r *gormRepository) GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.InvestmentTransaction, error) {
	var transactions []*domain.InvestmentTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND transaction_date >= ? AND transaction_date <= ?", userID, startDate, endDate).
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}

// Update updates an investment transaction
func (r *gormRepository) Update(ctx context.Context, transaction *domain.InvestmentTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// Delete soft deletes an investment transaction
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.InvestmentTransaction{}, id).Error
}

// GetSummary calculates transaction summary for given filters
func (r *gormRepository) GetSummary(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error) {
	var summary dto.TransactionSummary

	db := r.db.WithContext(ctx).Model(&domain.InvestmentTransaction{}).Where("user_id = ?", userID)

	// Apply filters
	if query.AssetID != "" {
		assetID, err := uuid.Parse(query.AssetID)
		if err == nil {
			db = db.Where("asset_id = ?", assetID)
		}
	}
	if query.StartDate != "" {
		db = db.Where("transaction_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("transaction_date <= ?", query.EndDate)
	}

	// Get aggregate data
	err := db.Select(
		"COUNT(*) as total_transactions",
		"COALESCE(SUM(CASE WHEN transaction_type = 'buy' THEN total_amount ELSE 0 END), 0) as total_bought",
		"COALESCE(SUM(CASE WHEN transaction_type = 'sell' THEN total_amount ELSE 0 END), 0) as total_sold",
		"COALESCE(SUM(CASE WHEN transaction_type = 'dividend' THEN total_amount ELSE 0 END), 0) as total_dividends",
		"COALESCE(SUM(fees + commission + tax), 0) as total_fees",
		"COALESCE(SUM(CASE WHEN transaction_type = 'sell' THEN realized_gain ELSE 0 END), 0) as total_realized_gain",
	).Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	summary.NetInvested = summary.TotalBought - summary.TotalSold

	return &summary, nil
}
