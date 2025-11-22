package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvestmentAsset represents an investment asset in a portfolio
type InvestmentAsset struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Asset identification
	Symbol     string    `gorm:"type:varchar(20);not null;index;column:symbol" json:"symbol"`            // Stock ticker, crypto symbol, etc.
	Name       string    `gorm:"type:varchar(255);not null;column:name" json:"name"`                     // Full name of the asset
	AssetType  AssetType `gorm:"type:varchar(30);not null;index;column:asset_type" json:"asset_type"`    // stock, crypto, bond, etc.
	AssetClass string    `gorm:"type:varchar(50);column:asset_class" json:"asset_class,omitempty"`       // equity, fixed_income, commodity, etc.
	Sector     *string   `gorm:"type:varchar(100);column:sector" json:"sector,omitempty"`                // Technology, Healthcare, etc.
	Industry   *string   `gorm:"type:varchar(100);column:industry" json:"industry,omitempty"`            // Software, Pharma, etc.
	Exchange   *string   `gorm:"type:varchar(50);column:exchange" json:"exchange,omitempty"`             // NYSE, NASDAQ, Binance, etc.
	Currency   string    `gorm:"type:varchar(3);not null;default:'USD';column:currency" json:"currency"` // USD, VND, EUR, etc.

	// Holding information
	Quantity           float64 `gorm:"type:decimal(20,8);not null;default:0;column:quantity" json:"quantity"`                           // Number of shares/units owned
	AverageCostPerUnit float64 `gorm:"type:decimal(15,2);not null;default:0;column:average_cost_per_unit" json:"average_cost_per_unit"` // Average purchase price
	TotalCost          float64 `gorm:"type:decimal(15,2);not null;default:0;column:total_cost" json:"total_cost"`                       // Total amount invested

	// Current market information
	CurrentPrice      float64 `gorm:"type:decimal(15,2);not null;default:0;column:current_price" json:"current_price"`             // Latest market price
	CurrentValue      float64 `gorm:"type:decimal(15,2);not null;default:0;column:current_value" json:"current_value"`             // Quantity * CurrentPrice
	UnrealizedGain    float64 `gorm:"type:decimal(15,2);not null;default:0;column:unrealized_gain" json:"unrealized_gain"`         // CurrentValue - TotalCost
	UnrealizedGainPct float64 `gorm:"type:decimal(10,4);not null;default:0;column:unrealized_gain_pct" json:"unrealized_gain_pct"` // (UnrealizedGain / TotalCost) * 100

	// Realized gains (from sales)
	RealizedGain    float64 `gorm:"type:decimal(15,2);not null;default:0;column:realized_gain" json:"realized_gain"`         // Total profit/loss from sales
	RealizedGainPct float64 `gorm:"type:decimal(10,4);not null;default:0;column:realized_gain_pct" json:"realized_gain_pct"` // Average return percentage on sales

	// Portfolio allocation
	PortfolioWeight float64 `gorm:"type:decimal(10,4);not null;default:0;column:portfolio_weight" json:"portfolio_weight"` // Percentage of total portfolio value

	// Dividend/Income tracking
	TotalDividends     float64 `gorm:"type:decimal(15,2);not null;default:0;column:total_dividends" json:"total_dividends"`            // Total dividends received
	DividendYield      float64 `gorm:"type:decimal(10,4);not null;default:0;column:dividend_yield" json:"dividend_yield"`              // Annual dividend yield %
	LastDividendAmount float64 `gorm:"type:decimal(15,2);default:0;column:last_dividend_amount" json:"last_dividend_amount,omitempty"` // Most recent dividend
	LastDividendDate   *string `gorm:"type:date;column:last_dividend_date" json:"last_dividend_date,omitempty"`                        // Date of last dividend

	// Risk metrics
	Beta            *float64 `gorm:"type:decimal(10,4);column:beta" json:"beta,omitempty"`                           // Market volatility measure
	Volatility      *float64 `gorm:"type:decimal(10,4);column:volatility" json:"volatility,omitempty"`               // Standard deviation of returns
	SharpeRatio     *float64 `gorm:"type:decimal(10,4);column:sharpe_ratio" json:"sharpe_ratio,omitempty"`           // Risk-adjusted return
	MaxDrawdown     *float64 `gorm:"type:decimal(10,4);column:max_drawdown" json:"max_drawdown,omitempty"`           // Maximum observed loss
	OneYearReturn   *float64 `gorm:"type:decimal(10,4);column:one_year_return" json:"one_year_return,omitempty"`     // Return over last year
	ThreeYearReturn *float64 `gorm:"type:decimal(10,4);column:three_year_return" json:"three_year_return,omitempty"` // Return over last 3 years
	FiveYearReturn  *float64 `gorm:"type:decimal(10,4);column:five_year_return" json:"five_year_return,omitempty"`   // Return over last 5 years

	// Status and metadata
	Status       AssetStatus `gorm:"type:varchar(20);not null;default:'active';column:status" json:"status"`
	IsWatchlist  bool        `gorm:"default:false;column:is_watchlist" json:"is_watchlist"` // Track without owning
	Notes        *string     `gorm:"type:text;column:notes" json:"notes,omitempty"`
	Tags         *string     `gorm:"type:text;column:tags" json:"tags,omitempty"`                    // JSON array or comma-separated
	CustomFields *string     `gorm:"type:jsonb;column:custom_fields" json:"custom_fields,omitempty"` // Additional user-defined fields

	// External integration
	ExternalID     *string `gorm:"type:varchar(255);index;column:external_id" json:"external_id,omitempty"`
	ExternalSource *string `gorm:"type:varchar(100);column:external_source" json:"external_source,omitempty"` // broker name, exchange, etc.
	ISIN           *string `gorm:"type:varchar(12);column:isin" json:"isin,omitempty"`                        // International Securities ID
	CUSIP          *string `gorm:"type:varchar(9);column:cusip" json:"cusip,omitempty"`                       // US Securities ID

	// Auto-update settings
	AutoUpdatePrice bool    `gorm:"default:true;column:auto_update_price" json:"auto_update_price"`             // Whether to auto-fetch prices
	LastPriceUpdate *string `gorm:"type:timestamp;column:last_price_update" json:"last_price_update,omitempty"` // Last time price was updated

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the database table name
func (InvestmentAsset) TableName() string {
	return "investment_assets"
}

// CalculateMetrics recalculates derived metrics
func (a *InvestmentAsset) CalculateMetrics() {
	// Calculate current value
	a.CurrentValue = a.Quantity * a.CurrentPrice

	// Calculate unrealized gain
	a.UnrealizedGain = a.CurrentValue - a.TotalCost

	// Calculate unrealized gain percentage
	if a.TotalCost > 0 {
		a.UnrealizedGainPct = (a.UnrealizedGain / a.TotalCost) * 100
	} else {
		a.UnrealizedGainPct = 0
	}
}

// AddQuantity adds to the holding and updates average cost
func (a *InvestmentAsset) AddQuantity(quantity, pricePerUnit float64) {
	newTotalCost := a.TotalCost + (quantity * pricePerUnit)
	newQuantity := a.Quantity + quantity

	if newQuantity > 0 {
		a.AverageCostPerUnit = newTotalCost / newQuantity
	}

	a.Quantity = newQuantity
	a.TotalCost = newTotalCost
	a.CalculateMetrics()
}

// RemoveQuantity removes from the holding and calculates realized gain
func (a *InvestmentAsset) RemoveQuantity(quantity, pricePerUnit float64) float64 {
	if quantity > a.Quantity {
		quantity = a.Quantity
	}

	// Calculate realized gain from this sale
	costBasis := quantity * a.AverageCostPerUnit
	saleProceeds := quantity * pricePerUnit
	realizedGain := saleProceeds - costBasis

	// Update totals
	a.Quantity -= quantity
	a.TotalCost -= costBasis
	a.RealizedGain += realizedGain

	// Update realized gain percentage
	if a.TotalCost > 0 {
		a.RealizedGainPct = (a.RealizedGain / a.TotalCost) * 100
	}

	a.CalculateMetrics()

	return realizedGain
}

// UpdatePrice updates the current price and recalculates metrics
func (a *InvestmentAsset) UpdatePrice(newPrice float64) {
	a.CurrentPrice = newPrice
	a.CalculateMetrics()
}

// IsActive returns true if the asset is active
func (a *InvestmentAsset) IsActive() bool {
	return a.Status == AssetStatusActive
}

// IsSold returns true if the asset has been completely sold
func (a *InvestmentAsset) IsSold() bool {
	return a.Status == AssetStatusSold || a.Quantity == 0
}

// TotalReturn calculates the total return (realized + unrealized)
func (a *InvestmentAsset) TotalReturn() float64 {
	return a.RealizedGain + a.UnrealizedGain
}

// TotalReturnPct calculates the total return percentage
func (a *InvestmentAsset) TotalReturnPct() float64 {
	if a.TotalCost > 0 {
		return (a.TotalReturn() / a.TotalCost) * 100
	}
	return 0
}
