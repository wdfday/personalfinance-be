package service

import (
	"context"

	"personalfinancedss/internal/module/investment/investment_asset/domain"
	"personalfinancedss/internal/module/investment/investment_asset/dto"
	"personalfinancedss/internal/module/investment/investment_asset/repository"
)

// AssetCreator defines asset creation operations
type AssetCreator interface {
	CreateAsset(ctx context.Context, userID string, req dto.CreateAssetRequest) (*domain.InvestmentAsset, error)
}

// AssetReader defines asset read operations
type AssetReader interface {
	GetAsset(ctx context.Context, userID string, assetID string) (*domain.InvestmentAsset, error)
	ListAssets(ctx context.Context, userID string, query dto.ListAssetsQuery) (*dto.AssetListResponse, error)
	GetPortfolioSummary(ctx context.Context, userID string) (*dto.PortfolioSummary, error)
	GetWatchlist(ctx context.Context, userID string) ([]*domain.InvestmentAsset, error)
	GetAssetsByType(ctx context.Context, userID string, assetType string) ([]*domain.InvestmentAsset, error)
}

// AssetUpdater defines asset update operations
type AssetUpdater interface {
	UpdateAsset(ctx context.Context, userID string, assetID string, req dto.UpdateAssetRequest) (*domain.InvestmentAsset, error)
	UpdatePrice(ctx context.Context, userID string, assetID string, req dto.UpdatePriceRequest) (*domain.InvestmentAsset, error)
	BulkUpdatePrices(ctx context.Context, req dto.BulkUpdatePricesRequest) error
}

// AssetDeleter defines asset delete operations
type AssetDeleter interface {
	DeleteAsset(ctx context.Context, userID string, assetID string) error
}

// AssetTransactor defines asset transaction operations
type AssetTransactor interface {
	BuyAsset(ctx context.Context, userID string, assetID string, req dto.BuyAssetRequest) (*domain.InvestmentAsset, error)
	SellAsset(ctx context.Context, userID string, assetID string, req dto.SellAssetRequest) (*domain.InvestmentAsset, float64, error)
	AddDividend(ctx context.Context, userID string, assetID string, req dto.AddDividendRequest) (*domain.InvestmentAsset, error)
}

// Service is the composite interface for all asset operations
type Service interface {
	AssetCreator
	AssetReader
	AssetUpdater
	AssetDeleter
	AssetTransactor
}

// assetService implements all asset use cases
type assetService struct {
	repo repository.Repository
}

// NewService creates a new asset service
func NewService(repo repository.Repository) Service {
	return &assetService{
		repo: repo,
	}
}
