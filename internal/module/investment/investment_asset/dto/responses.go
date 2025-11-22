package dto

import (
	"time"

	"github.com/google/uuid"
)

// AssetResponse represents an investment asset response
type AssetResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Asset identification
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	AssetType  string `json:"asset_type"`
	AssetClass string `json:"asset_class,omitempty"`
	Sector     string `json:"sector,omitempty"`
	Industry   string `json:"industry,omitempty"`
	Exchange   string `json:"exchange,omitempty"`
	Currency   string `json:"currency"`

	// Holding information
	Quantity           float64 `json:"quantity"`
	AverageCostPerUnit float64 `json:"average_cost_per_unit"`
	TotalCost          float64 `json:"total_cost"`

	// Current market information
	CurrentPrice      float64 `json:"current_price"`
	CurrentValue      float64 `json:"current_value"`
	UnrealizedGain    float64 `json:"unrealized_gain"`
	UnrealizedGainPct float64 `json:"unrealized_gain_pct"`

	// Realized gains
	RealizedGain    float64 `json:"realized_gain"`
	RealizedGainPct float64 `json:"realized_gain_pct"`

	// Portfolio allocation
	PortfolioWeight float64 `json:"portfolio_weight"`

	// Dividends
	TotalDividends     float64 `json:"total_dividends"`
	DividendYield      float64 `json:"dividend_yield"`
	LastDividendAmount float64 `json:"last_dividend_amount,omitempty"`
	LastDividendDate   string  `json:"last_dividend_date,omitempty"`

	// Risk metrics
	Beta            *float64 `json:"beta,omitempty"`
	Volatility      *float64 `json:"volatility,omitempty"`
	SharpeRatio     *float64 `json:"sharpe_ratio,omitempty"`
	MaxDrawdown     *float64 `json:"max_drawdown,omitempty"`
	OneYearReturn   *float64 `json:"one_year_return,omitempty"`
	ThreeYearReturn *float64 `json:"three_year_return,omitempty"`
	FiveYearReturn  *float64 `json:"five_year_return,omitempty"`

	// Status and metadata
	Status          string `json:"status"`
	IsWatchlist     bool   `json:"is_watchlist"`
	Notes           string `json:"notes,omitempty"`
	Tags            string `json:"tags,omitempty"`
	AutoUpdatePrice bool   `json:"auto_update_price"`
	LastPriceUpdate string `json:"last_price_update,omitempty"`

	// External identifiers
	ISIN  string `json:"isin,omitempty"`
	CUSIP string `json:"cusip,omitempty"`

	// Calculated fields
	TotalReturn    float64 `json:"total_return"`
	TotalReturnPct float64 `json:"total_return_pct"`
}

// AssetListResponse represents a paginated list of assets
type AssetListResponse struct {
	Assets  []AssetResponse `json:"assets"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
}

// PortfolioSummary represents an overview of the entire portfolio
type PortfolioSummary struct {
	TotalAssets         int                         `json:"total_assets"`
	TotalInvested       float64                     `json:"total_invested"`
	TotalValue          float64                     `json:"total_value"`
	TotalUnrealizedGain float64                     `json:"total_unrealized_gain"`
	TotalRealizedGain   float64                     `json:"total_realized_gain"`
	TotalGain           float64                     `json:"total_gain"`
	TotalGainPct        float64                     `json:"total_gain_pct"`
	TotalDividends      float64                     `json:"total_dividends"`
	ByAssetType         map[string]AssetTypeSummary `json:"by_asset_type"`
}

// AssetTypeSummary represents summary for a specific asset type
type AssetTypeSummary struct {
	AssetType  string  `json:"asset_type"`
	TotalValue float64 `json:"total_value"`
	TotalCost  float64 `json:"total_cost"`
	AssetCount int     `json:"asset_count"`
	Percentage float64 `json:"percentage"`
}

// TransactionResponse represents the response after a buy/sell transaction
type TransactionResponse struct {
	Asset         AssetResponse `json:"asset"`
	TransactionID uuid.UUID     `json:"transaction_id,omitempty"`
	RealizedGain  *float64      `json:"realized_gain,omitempty"` // Only for sell transactions
	Message       string        `json:"message"`
}
