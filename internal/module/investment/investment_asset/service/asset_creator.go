package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
)

// CreateAsset creates a new investment asset
func (s *assetService) CreateAsset(ctx context.Context, userID string, req dto.CreateAssetRequest) (*domain.InvestmentAsset, error) {
	// Validate asset type
	if !req.AssetType.IsValid() {
		return nil, fmt.Errorf("invalid asset type: %s", req.AssetType)
	}

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if asset already exists (for non-watchlist items)
	if !req.IsWatchlist && req.Quantity > 0 {
		existing, err := s.repo.GetBySymbol(ctx, uid, req.Symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing asset: %w", err)
		}
		if existing != nil && existing.Status == domain.AssetStatusActive {
			return nil, fmt.Errorf("asset with symbol %s already exists", req.Symbol)
		}
	}

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Create asset entity
	asset := &domain.InvestmentAsset{
		UserID:             uid,
		Symbol:             req.Symbol,
		Name:               req.Name,
		AssetType:          req.AssetType,
		AssetClass:         req.AssetClass,
		Sector:             req.Sector,
		Industry:           req.Industry,
		Exchange:           req.Exchange,
		Currency:           currency,
		Quantity:           req.Quantity,
		AverageCostPerUnit: req.PricePerUnit,
		TotalCost:          req.Quantity * req.PricePerUnit,
		CurrentPrice:       req.PricePerUnit,
		IsWatchlist:        req.IsWatchlist,
		Notes:              req.Notes,
		Tags:               req.Tags,
		ISIN:               req.ISIN,
		CUSIP:              req.CUSIP,
		AutoUpdatePrice:    true,
	}

	// Set status
	if req.IsWatchlist {
		asset.Status = domain.AssetStatusWatchlist
	} else {
		asset.Status = domain.AssetStatusActive
	}

	// Calculate initial metrics
	asset.CalculateMetrics()

	// Save to repository
	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}
