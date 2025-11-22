package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
)

// GetAsset retrieves a single asset by ID
func (s *assetService) GetAsset(ctx context.Context, userID string, assetID string) (*domain.InvestmentAsset, error) {
	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Get asset
	asset, err := s.repo.GetByUserID(ctx, aid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

// ListAssets retrieves a list of assets with pagination and filters
func (s *assetService) ListAssets(ctx context.Context, userID string, query dto.ListAssetsQuery) (*dto.AssetListResponse, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Get assets from repository
	assets, total, err := s.repo.List(ctx, uid, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	// Convert to response
	response := dto.ToAssetListResponse(assets, total, query.Page, query.PageSize)

	return &response, nil
}

// GetPortfolioSummary retrieves portfolio summary for a user
func (s *assetService) GetPortfolioSummary(ctx context.Context, userID string) (*dto.PortfolioSummary, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get summary from repository
	summary, err := s.repo.GetPortfolioSummary(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio summary: %w", err)
	}

	return summary, nil
}

// GetWatchlist retrieves all watchlist items for a user
func (s *assetService) GetWatchlist(ctx context.Context, userID string) ([]*domain.InvestmentAsset, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get watchlist from repository
	assets, err := s.repo.GetWatchlist(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	return assets, nil
}

// GetAssetsByType retrieves all assets of a specific type for a user
func (s *assetService) GetAssetsByType(ctx context.Context, userID string, assetType string) ([]*domain.InvestmentAsset, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate asset type
	aType := domain.AssetType(assetType)
	if !aType.IsValid() {
		return nil, fmt.Errorf("invalid asset type: %s", assetType)
	}

	// Get assets from repository
	assets, err := s.repo.GetAssetsByType(ctx, uid, aType)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets by type: %w", err)
	}

	return assets, nil
}
