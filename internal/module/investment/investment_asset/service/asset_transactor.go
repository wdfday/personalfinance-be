package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
)

// BuyAsset adds more quantity to an existing asset
func (s *assetService) BuyAsset(ctx context.Context, userID string, assetID string, req dto.BuyAssetRequest) (*domain.InvestmentAsset, error) {
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

	// Check if watchlist item
	if asset.IsWatchlist {
		// Convert from watchlist to active holding
		asset.IsWatchlist = false
		asset.Status = domain.AssetStatusActive
	}

	// Add quantity and update cost basis
	asset.AddQuantity(req.Quantity, req.PricePerUnit)

	// Update current price
	asset.CurrentPrice = req.PricePerUnit
	now := time.Now().Format(time.RFC3339)
	asset.LastPriceUpdate = &now

	// Save updated asset
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to update asset after buy: %w", err)
	}

	return asset, nil
}

// SellAsset removes quantity from an existing asset
func (s *assetService) SellAsset(ctx context.Context, userID string, assetID string, req dto.SellAssetRequest) (*domain.InvestmentAsset, float64, error) {
	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Get existing asset
	asset, err := s.repo.GetByUserID(ctx, aid, uid)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get asset: %w", err)
	}

	// Check if watchlist item
	if asset.IsWatchlist {
		return nil, 0, fmt.Errorf("cannot sell a watchlist item")
	}

	// Check if sufficient quantity
	if req.Quantity > asset.Quantity {
		return nil, 0, fmt.Errorf("insufficient quantity: have %f, trying to sell %f", asset.Quantity, req.Quantity)
	}

	// Remove quantity and calculate realized gain
	realizedGain := asset.RemoveQuantity(req.Quantity, req.PricePerUnit)

	// Update current price
	asset.CurrentPrice = req.PricePerUnit
	now := time.Now().Format(time.RFC3339)
	asset.LastPriceUpdate = &now

	// If completely sold, update status
	if asset.Quantity == 0 {
		asset.Status = domain.AssetStatusSold
	}

	// Save updated asset
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, 0, fmt.Errorf("failed to update asset after sell: %w", err)
	}

	return asset, realizedGain, nil
}

// AddDividend adds a dividend payment to an asset
func (s *assetService) AddDividend(ctx context.Context, userID string, assetID string, req dto.AddDividendRequest) (*domain.InvestmentAsset, error) {
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

	// Add to total dividends
	asset.TotalDividends += req.Amount
	asset.LastDividendAmount = req.Amount
	asset.LastDividendDate = &req.Date

	// Recalculate dividend yield based on current price
	if asset.CurrentPrice > 0 {
		asset.DividendYield = (req.Amount / asset.CurrentPrice) * 100
	}

	// Save updated asset
	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to update asset after dividend: %w", err)
	}

	return asset, nil
}
