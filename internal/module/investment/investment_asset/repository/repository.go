package repository

import (
	"context"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"

	"github.com/google/uuid"
)

// Repository defines investment asset data access operations
type Repository interface {
	// Create creates a new investment asset
	Create(ctx context.Context, asset *domain.InvestmentAsset) error

	// GetByID retrieves an investment asset by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentAsset, error)

	// GetByUserID retrieves an investment asset by ID and user ID
	GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.InvestmentAsset, error)

	// GetBySymbol retrieves an investment asset by symbol and user ID
	GetBySymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.InvestmentAsset, error)

	// List retrieves investment assets with filters and pagination
	List(ctx context.Context, userID uuid.UUID, query dto.ListAssetsQuery) ([]*domain.InvestmentAsset, int64, error)

	// Update updates an investment asset
	Update(ctx context.Context, asset *domain.InvestmentAsset) error

	// UpdateColumns updates specific columns of an investment asset
	UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]any) error

	// Delete soft deletes an investment asset
	Delete(ctx context.Context, id uuid.UUID) error

	// GetPortfolioSummary calculates portfolio summary for a user
	GetPortfolioSummary(ctx context.Context, userID uuid.UUID) (*dto.PortfolioSummary, error)

	// GetAssetsByType retrieves all assets of a specific type for a user
	GetAssetsByType(ctx context.Context, userID uuid.UUID, assetType domain.AssetType) ([]*domain.InvestmentAsset, error)

	// GetWatchlist retrieves all watchlist items for a user
	GetWatchlist(ctx context.Context, userID uuid.UUID) ([]*domain.InvestmentAsset, error)

	// UpdatePrices bulk updates prices for multiple assets
	UpdatePrices(ctx context.Context, updates map[uuid.UUID]float64) error

	// GetActiveAssets retrieves all active assets for a user
	GetActiveAssets(ctx context.Context, userID uuid.UUID) ([]*domain.InvestmentAsset, error)
}
