package dto

import (
	"time"

	"github.com/google/uuid"
)

// SnapshotResponse represents a portfolio snapshot response
type SnapshotResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	SnapshotDate time.Time `json:"snapshot_date"`
	SnapshotType string    `json:"snapshot_type"`
	Period       string    `json:"period,omitempty"`

	TotalValue          float64 `json:"total_value"`
	TotalCost           float64 `json:"total_cost"`
	TotalUnrealizedGain float64 `json:"total_unrealized_gain"`
	TotalRealizedGain   float64 `json:"total_realized_gain"`
	TotalDividends      float64 `json:"total_dividends"`
	TotalReturn         float64 `json:"total_return"`
	TotalReturnPct      float64 `json:"total_return_pct"`
	DayChange           float64 `json:"day_change"`
	DayChangePct        float64 `json:"day_change_pct"`

	TotalAssets      int    `json:"total_assets"`
	ActiveAssets     int    `json:"active_assets"`
	AssetTypes       string `json:"asset_types,omitempty"`
	SectorAllocation string `json:"sector_allocation,omitempty"`

	CashInflow  float64 `json:"cash_inflow"`
	CashOutflow float64 `json:"cash_outflow"`
	NetCashFlow float64 `json:"net_cash_flow"`

	Volatility  *float64 `json:"volatility,omitempty"`
	SharpeRatio *float64 `json:"sharpe_ratio,omitempty"`
	Beta        *float64 `json:"beta,omitempty"`

	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SnapshotListResponse represents a paginated list of snapshots
type SnapshotListResponse struct {
	Snapshots []SnapshotResponse `json:"snapshots"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PerPage   int                `json:"per_page"`
}

// PerformanceMetrics represents portfolio performance over a time period
type PerformanceMetrics struct {
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	StartValue     float64   `json:"start_value"`
	EndValue       float64   `json:"end_value"`
	TotalReturn    float64   `json:"total_return"`
	TotalReturnPct float64   `json:"total_return_pct"`
}
