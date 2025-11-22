package dto

import (
	"time"

	"github.com/google/uuid"
)

// TransactionResponse represents an investment transaction response
type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	AssetID   uuid.UUID `json:"asset_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Transaction details
	TransactionType string  `json:"transaction_type"`
	Quantity        float64 `json:"quantity"`
	PricePerUnit    float64 `json:"price_per_unit"`
	TotalAmount     float64 `json:"total_amount"`
	Currency        string  `json:"currency"`

	// Fees and costs
	Fees       float64 `json:"fees"`
	Commission float64 `json:"commission"`
	Tax        float64 `json:"tax"`
	TotalCost  float64 `json:"total_cost"`

	// For sell transactions
	RealizedGain    *float64 `json:"realized_gain,omitempty"`
	RealizedGainPct *float64 `json:"realized_gain_pct,omitempty"`

	// Transaction metadata
	TransactionDate time.Time  `json:"transaction_date"`
	SettlementDate  *time.Time `json:"settlement_date,omitempty"`
	Status          string     `json:"status"`

	// Description and notes
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`

	// Broker/Exchange information
	Broker   string `json:"broker,omitempty"`
	Exchange string `json:"exchange,omitempty"`
	OrderID  string `json:"order_id,omitempty"`

	// Additional fields
	Tags string `json:"tags,omitempty"`
}

// TransactionListResponse represents a paginated list of transactions
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	PerPage      int                   `json:"per_page"`
}

// TransactionSummary represents aggregate transaction statistics
type TransactionSummary struct {
	TotalTransactions int     `json:"total_transactions"`
	TotalBought       float64 `json:"total_bought"`
	TotalSold         float64 `json:"total_sold"`
	TotalDividends    float64 `json:"total_dividends"`
	TotalFees         float64 `json:"total_fees"`
	TotalRealizedGain float64 `json:"total_realized_gain"`
	NetInvested       float64 `json:"net_invested"`
}
