package dto

import "personalfinancedss/internal/module/investment/investment_asset/domain"

// CreateAssetRequest represents a request to create a new investment asset
type CreateAssetRequest struct {
	Symbol       string           `json:"symbol" binding:"required,max=20"`
	Name         string           `json:"name" binding:"required,max=255"`
	AssetType    domain.AssetType `json:"asset_type" binding:"required"`
	AssetClass   string           `json:"asset_class,omitempty" binding:"max=50"`
	Sector       *string          `json:"sector,omitempty" binding:"omitempty,max=100"`
	Industry     *string          `json:"industry,omitempty" binding:"omitempty,max=100"`
	Exchange     *string          `json:"exchange,omitempty" binding:"omitempty,max=50"`
	Currency     string           `json:"currency,omitempty" binding:"omitempty,len=3"`
	Quantity     float64          `json:"quantity" binding:"required,min=0"`
	PricePerUnit float64          `json:"price_per_unit" binding:"required,min=0"`
	IsWatchlist  bool             `json:"is_watchlist"`
	Notes        *string          `json:"notes,omitempty"`
	Tags         *string          `json:"tags,omitempty"`
	ISIN         *string          `json:"isin,omitempty" binding:"omitempty,len=12"`
	CUSIP        *string          `json:"cusip,omitempty" binding:"omitempty,len=9"`
}

// UpdateAssetRequest represents a request to update an investment asset
type UpdateAssetRequest struct {
	Name            *string             `json:"name,omitempty" binding:"omitempty,max=255"`
	AssetClass      *string             `json:"asset_class,omitempty" binding:"omitempty,max=50"`
	Sector          *string             `json:"sector,omitempty" binding:"omitempty,max=100"`
	Industry        *string             `json:"industry,omitempty" binding:"omitempty,max=100"`
	Exchange        *string             `json:"exchange,omitempty" binding:"omitempty,max=50"`
	Currency        *string             `json:"currency,omitempty" binding:"omitempty,len=3"`
	Status          *domain.AssetStatus `json:"status,omitempty"`
	IsWatchlist     *bool               `json:"is_watchlist,omitempty"`
	Notes           *string             `json:"notes,omitempty"`
	Tags            *string             `json:"tags,omitempty"`
	AutoUpdatePrice *bool               `json:"auto_update_price,omitempty"`
	DividendYield   *float64            `json:"dividend_yield,omitempty" binding:"omitempty,min=0,max=100"`
}

// BuyAssetRequest represents a request to buy more units of an asset
type BuyAssetRequest struct {
	Quantity     float64 `json:"quantity" binding:"required,min=0.00000001"`
	PricePerUnit float64 `json:"price_per_unit" binding:"required,min=0"`
	Notes        *string `json:"notes,omitempty"`
}

// SellAssetRequest represents a request to sell units of an asset
type SellAssetRequest struct {
	Quantity     float64 `json:"quantity" binding:"required,min=0.00000001"`
	PricePerUnit float64 `json:"price_per_unit" binding:"required,min=0"`
	Notes        *string `json:"notes,omitempty"`
}

// UpdatePriceRequest represents a request to update the current price of an asset
type UpdatePriceRequest struct {
	CurrentPrice float64 `json:"current_price" binding:"required,min=0"`
}

// BulkUpdatePricesRequest represents a request to update prices for multiple assets
type BulkUpdatePricesRequest struct {
	Updates []PriceUpdate `json:"updates" binding:"required,min=1,dive"`
}

// PriceUpdate represents a single price update
type PriceUpdate struct {
	AssetID string  `json:"asset_id" binding:"required,uuid"`
	Price   float64 `json:"price" binding:"required,min=0"`
}

// AddDividendRequest represents a request to add a dividend payment
type AddDividendRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0"`
	Date   string  `json:"date" binding:"required"` // YYYY-MM-DD format
	Notes  *string `json:"notes,omitempty"`
}

// ListAssetsQuery represents query parameters for listing assets
type ListAssetsQuery struct {
	AssetType   string  `form:"asset_type"`
	Status      string  `form:"status"`
	Symbol      string  `form:"symbol"`
	Name        string  `form:"name"`
	Sector      string  `form:"sector"`
	Industry    string  `form:"industry"`
	Exchange    string  `form:"exchange"`
	IsWatchlist *bool   `form:"is_watchlist"`
	MinValue    float64 `form:"min_value" binding:"omitempty,min=0"`
	MaxValue    float64 `form:"max_value" binding:"omitempty,min=0"`
	Page        int     `form:"page" binding:"omitempty,min=1"`
	PageSize    int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy      string  `form:"sort_by"` // current_value, unrealized_gain, symbol, etc.
	SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
