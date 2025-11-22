package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
)

// UpdateAsset updates an existing asset
func (s *assetService) UpdateAsset(ctx context.Context, userID string, assetID string, req dto.UpdateAssetRequest) (*domain.InvestmentAsset, error) {
	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Get existing asset
	asset, err := s.repo.GetByUserID(ctx, aid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		asset.Name = *req.Name
	}
	if req.AssetClass != nil {
		asset.AssetClass = *req.AssetClass
	}
	if req.Sector != nil {
		asset.Sector = req.Sector
	}
	if req.Industry != nil {
		asset.Industry = req.Industry
	}
	if req.Exchange != nil {
		asset.Exchange = req.Exchange
	}
	if req.Currency != nil {
		asset.Currency = *req.Currency
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, fmt.Errorf("invalid asset status: %s", *req.Status)
		}
		asset.Status = *req.Status
	}
	if req.IsWatchlist != nil {
		asset.IsWatchlist = *req.IsWatchlist
	}
	if req.Notes != nil {
		asset.Notes = req.Notes
	}
	if req.Tags != nil {
		asset.Tags = req.Tags
	}
	if req.AutoUpdatePrice != nil {
		asset.AutoUpdatePrice = *req.AutoUpdatePrice
	}
	if req.DividendYield != nil {
		asset.DividendYield = *req.DividendYield
	}

	// Save updated asset
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	return asset, nil
}

// UpdatePrice updates the current price of an asset
func (s *assetService) UpdatePrice(ctx context.Context, userID string, assetID string, req dto.UpdatePriceRequest) (*domain.InvestmentAsset, error) {
	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Get existing asset
	asset, err := s.repo.GetByUserID(ctx, aid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Update price and recalculate metrics
	asset.UpdatePrice(req.CurrentPrice)

	// Update timestamp
	now := time.Now().Format(time.RFC3339)
	asset.LastPriceUpdate = &now

	// Save updated asset
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to update asset price: %w", err)
	}

	return asset, nil
}

// BulkUpdatePrices updates prices for multiple assets
func (s *assetService) BulkUpdatePrices(ctx context.Context, req dto.BulkUpdatePricesRequest) error {
	// Build update map
	updates := make(map[uuid.UUID]float64)

	for _, update := range req.Updates {
		assetID, err := uuid.Parse(update.AssetID)
		if err != nil {
			return fmt.Errorf("invalid asset ID %s: %w", update.AssetID, err)
		}
		updates[assetID] = update.Price
	}

	// Perform bulk update
	if err := s.repo.UpdatePrices(ctx, updates); err != nil {
		return fmt.Errorf("failed to bulk update prices: %w", err)
	}

	return nil
}
