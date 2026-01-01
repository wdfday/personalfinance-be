package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PortfolioSnapshot represents a point-in-time snapshot of a user's investment portfolio
type PortfolioSnapshot struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Snapshot metadata
	SnapshotDate time.Time       `gorm:"not null;index;column:snapshot_date" json:"snapshot_date"`
	SnapshotType SnapshotType    `gorm:"type:varchar(20);not null;index;column:snapshot_type" json:"snapshot_type"`
	Period       *SnapshotPeriod `gorm:"type:varchar(20);column:period" json:"period,omitempty"` // daily, weekly, monthly, yearly

	// Portfolio values
	TotalValue          float64 `gorm:"type:decimal(15,2);not null;column:total_value" json:"total_value"`
	TotalCost           float64 `gorm:"type:decimal(15,2);not null;column:total_cost" json:"total_cost"`
	TotalUnrealizedGain float64 `gorm:"type:decimal(15,2);not null;column:total_unrealized_gain" json:"total_unrealized_gain"`
	TotalRealizedGain   float64 `gorm:"type:decimal(15,2);not null;column:total_realized_gain" json:"total_realized_gain"`
	TotalDividends      float64 `gorm:"type:decimal(15,2);not null;column:total_dividends" json:"total_dividends"`

	// Returns and percentages
	TotalReturn    float64 `gorm:"type:decimal(15,2);not null;column:total_return" json:"total_return"`         // Unrealized + Realized
	TotalReturnPct float64 `gorm:"type:decimal(10,4);not null;column:total_return_pct" json:"total_return_pct"` // Percentage return
	DayChange      float64 `gorm:"type:decimal(15,2);default:0;column:day_change" json:"day_change"`            // Change from previous day
	DayChangePct   float64 `gorm:"type:decimal(10,4);default:0;column:day_change_pct" json:"day_change_pct"`    // Percentage change from previous day

	// Asset counts and allocation
	TotalAssets      int     `gorm:"not null;default:0;column:total_assets" json:"total_assets"`
	ActiveAssets     int     `gorm:"not null;default:0;column:active_assets" json:"active_assets"`
	AssetTypes       *string `gorm:"type:jsonb;column:asset_types" json:"asset_types,omitempty"`             // JSON object with type breakdown
	SectorAllocation *string `gorm:"type:jsonb;column:sector_allocation" json:"sector_allocation,omitempty"` // JSON object with sector breakdown

	// Cash flow for the period
	CashInflow  float64 `gorm:"type:decimal(15,2);default:0;column:cash_inflow" json:"cash_inflow"`     // Buys during period
	CashOutflow float64 `gorm:"type:decimal(15,2);default:0;column:cash_outflow" json:"cash_outflow"`   // Sells during period
	NetCashFlow float64 `gorm:"type:decimal(15,2);default:0;column:net_cash_flow" json:"net_cash_flow"` // Inflow - Outflow

	// Performance metrics
	Volatility  *float64 `gorm:"type:decimal(10,4);column:volatility" json:"volatility,omitempty"`     // Portfolio volatility
	SharpeRatio *float64 `gorm:"type:decimal(10,4);column:sharpe_ratio" json:"sharpe_ratio,omitempty"` // Risk-adjusted return
	Beta        *float64 `gorm:"type:decimal(10,4);column:beta" json:"beta,omitempty"`                 // Market correlation

	// Notes and metadata
	Notes *string `gorm:"type:text;column:notes" json:"notes,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the database table name
func (PortfolioSnapshot) TableName() string {
	return "portfolio_snapshots"
}

// CalculateMetrics recalculates derived metrics
func (s *PortfolioSnapshot) CalculateMetrics() {
	// Calculate total return
	s.TotalReturn = s.TotalUnrealizedGain + s.TotalRealizedGain

	// Calculate return percentage
	if s.TotalCost > 0 {
		s.TotalReturnPct = (s.TotalReturn / s.TotalCost) * 100
	} else {
		s.TotalReturnPct = 0
	}

	// Calculate net cash flow
	s.NetCashFlow = s.CashInflow - s.CashOutflow
}

// CalculateDayChange calculates change from previous snapshot
func (s *PortfolioSnapshot) CalculateDayChange(previousValue float64) {
	s.DayChange = s.TotalValue - previousValue
	if previousValue > 0 {
		s.DayChangePct = (s.DayChange / previousValue) * 100
	} else {
		s.DayChangePct = 0
	}
}
