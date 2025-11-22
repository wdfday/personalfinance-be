package repository

import (
	"context"
	"errors"
	"fmt"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based investment asset repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new investment asset
func (r *gormRepository) Create(ctx context.Context, asset *domain.InvestmentAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// GetByID retrieves an investment asset by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentAsset, error) {
	var asset domain.InvestmentAsset
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("investment asset not found")
		}
		return nil, err
	}
	return &asset, nil
}

// GetByUserID retrieves an investment asset by ID and user ID
func (r *gormRepository) GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.InvestmentAsset, error) {
	var asset domain.InvestmentAsset
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("investment asset not found")
		}
		return nil, err
	}
	return &asset, nil
}

// GetBySymbol retrieves an investment asset by symbol and user ID
func (r *gormRepository) GetBySymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.InvestmentAsset, error) {
	var asset domain.InvestmentAsset
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND symbol = ? AND status = ?", userID, symbol, domain.AssetStatusActive).
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error if not found
		}
		return nil, err
	}
	return &asset, nil
}

// List retrieves investment assets with filters and pagination
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListAssetsQuery) ([]*domain.InvestmentAsset, int64, error) {
	var assets []*domain.InvestmentAsset
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.InvestmentAsset{}).Where("user_id = ?", userID)

	// Apply filters
	if query.AssetType != "" {
		db = db.Where("asset_type = ?", query.AssetType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Symbol != "" {
		db = db.Where("symbol ILIKE ?", "%"+query.Symbol+"%")
	}
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.Sector != "" {
		db = db.Where("sector = ?", query.Sector)
	}
	if query.Industry != "" {
		db = db.Where("industry = ?", query.Industry)
	}
	if query.Exchange != "" {
		db = db.Where("exchange = ?", query.Exchange)
	}
	if query.IsWatchlist != nil {
		db = db.Where("is_watchlist = ?", *query.IsWatchlist)
	}
	if query.MinValue > 0 {
		db = db.Where("current_value >= ?", query.MinValue)
	}
	if query.MaxValue > 0 {
		db = db.Where("current_value <= ?", query.MaxValue)
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "current_value" // Default sort by value
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
	if err := db.Limit(pageSize).Offset(offset).Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// Update updates an investment asset
func (r *gormRepository) Update(ctx context.Context, asset *domain.InvestmentAsset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

// UpdateColumns updates specific columns of an investment asset
func (r *gormRepository) UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&domain.InvestmentAsset{}).
		Where("id = ?", id).
		Updates(columns).Error
}

// Delete soft deletes an investment asset
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.InvestmentAsset{}, id).Error
}

// GetPortfolioSummary calculates portfolio summary for a user
func (r *gormRepository) GetPortfolioSummary(ctx context.Context, userID uuid.UUID) (*dto.PortfolioSummary, error) {
	var summary dto.PortfolioSummary

	// Get aggregate data
	err := r.db.WithContext(ctx).
		Model(&domain.InvestmentAsset{}).
		Where("user_id = ? AND status = ?", userID, domain.AssetStatusActive).
		Select(
			"COUNT(*) as total_assets",
			"COALESCE(SUM(total_cost), 0) as total_invested",
			"COALESCE(SUM(current_value), 0) as total_value",
			"COALESCE(SUM(unrealized_gain), 0) as total_unrealized_gain",
			"COALESCE(SUM(realized_gain), 0) as total_realized_gain",
			"COALESCE(SUM(total_dividends), 0) as total_dividends",
		).
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	// Calculate percentages
	if summary.TotalInvested > 0 {
		summary.TotalGain = summary.TotalUnrealizedGain + summary.TotalRealizedGain
		summary.TotalGainPct = (summary.TotalGain / summary.TotalInvested) * 100
	}

	// Get asset type breakdown
	type AssetTypeBreakdown struct {
		AssetType  domain.AssetType
		TotalValue float64
		TotalCost  float64
		AssetCount int
	}
	var breakdown []AssetTypeBreakdown

	err = r.db.WithContext(ctx).
		Model(&domain.InvestmentAsset{}).
		Where("user_id = ? AND status = ?", userID, domain.AssetStatusActive).
		Select(
			"asset_type",
			"COALESCE(SUM(current_value), 0) as total_value",
			"COALESCE(SUM(total_cost), 0) as total_cost",
			"COUNT(*) as asset_count",
		).
		Group("asset_type").
		Scan(&breakdown).Error

	if err != nil {
		return nil, err
	}

	summary.ByAssetType = make(map[string]dto.AssetTypeSummary)
	for _, b := range breakdown {
		percentage := 0.0
		if summary.TotalValue > 0 {
			percentage = (b.TotalValue / summary.TotalValue) * 100
		}
		summary.ByAssetType[string(b.AssetType)] = dto.AssetTypeSummary{
			AssetType:  string(b.AssetType),
			TotalValue: b.TotalValue,
			TotalCost:  b.TotalCost,
			AssetCount: b.AssetCount,
			Percentage: percentage,
		}
	}

	return &summary, nil
}

// GetAssetsByType retrieves all assets of a specific type for a user
func (r *gormRepository) GetAssetsByType(ctx context.Context, userID uuid.UUID, assetType domain.AssetType) ([]*domain.InvestmentAsset, error) {
	var assets []*domain.InvestmentAsset
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND asset_type = ? AND status = ?", userID, assetType, domain.AssetStatusActive).
		Order("current_value DESC").
		Find(&assets).Error
	return assets, err
}

// GetWatchlist retrieves all watchlist items for a user
func (r *gormRepository) GetWatchlist(ctx context.Context, userID uuid.UUID) ([]*domain.InvestmentAsset, error) {
	var assets []*domain.InvestmentAsset
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_watchlist = ?", userID, true).
		Order("symbol ASC").
		Find(&assets).Error
	return assets, err
}

// UpdatePrices bulk updates prices for multiple assets
func (r *gormRepository) UpdatePrices(ctx context.Context, updates map[uuid.UUID]float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for assetID, newPrice := range updates {
			var asset domain.InvestmentAsset
			if err := tx.Where("id = ?", assetID).First(&asset).Error; err != nil {
				continue // Skip if asset not found
			}

			asset.UpdatePrice(newPrice)
			if err := tx.Save(&asset).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetActiveAssets retrieves all active assets for a user
func (r *gormRepository) GetActiveAssets(ctx context.Context, userID uuid.UUID) ([]*domain.InvestmentAsset, error) {
	var assets []*domain.InvestmentAsset
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND quantity > 0", userID, domain.AssetStatusActive).
		Order("current_value DESC").
		Find(&assets).Error
	return assets, err
}
