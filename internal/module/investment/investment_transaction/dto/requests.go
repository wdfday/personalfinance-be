package dto

import "personalfinancedss/internal/module/investment/investment_transaction/domain"

// CreateTransactionRequest represents a request to create a new investment transaction
type CreateTransactionRequest struct {
	AssetID         string                    `json:"asset_id" binding:"required,uuid"`
	TransactionType domain.TransactionType    `json:"transaction_type" binding:"required"`
	Quantity        float64                   `json:"quantity" binding:"required,min=0.00000001"`
	PricePerUnit    float64                   `json:"price_per_unit" binding:"required,min=0"`
	Currency        string                    `json:"currency,omitempty" binding:"omitempty,len=3"`
	Fees            float64                   `json:"fees,omitempty" binding:"omitempty,min=0"`
	Commission      float64                   `json:"commission,omitempty" binding:"omitempty,min=0"`
	Tax             float64                   `json:"tax,omitempty" binding:"omitempty,min=0"`
	TransactionDate string                    `json:"transaction_date,omitempty"` // YYYY-MM-DD format
	SettlementDate  *string                   `json:"settlement_date,omitempty"`  // YYYY-MM-DD format
	Status          *domain.TransactionStatus `json:"status,omitempty"`
	Description     string                    `json:"description,omitempty" binding:"omitempty,max=500"`
	Notes           *string                   `json:"notes,omitempty"`
	Broker          *string                   `json:"broker,omitempty" binding:"omitempty,max=100"`
	Exchange        *string                   `json:"exchange,omitempty" binding:"omitempty,max=50"`
	OrderID         *string                   `json:"order_id,omitempty" binding:"omitempty,max=100"`
	Tags            *string                   `json:"tags,omitempty"`
}

// UpdateTransactionRequest represents a request to update an investment transaction
type UpdateTransactionRequest struct {
	Quantity        *float64                  `json:"quantity,omitempty" binding:"omitempty,min=0.00000001"`
	PricePerUnit    *float64                  `json:"price_per_unit,omitempty" binding:"omitempty,min=0"`
	Fees            *float64                  `json:"fees,omitempty" binding:"omitempty,min=0"`
	Commission      *float64                  `json:"commission,omitempty" binding:"omitempty,min=0"`
	Tax             *float64                  `json:"tax,omitempty" binding:"omitempty,min=0"`
	TransactionDate *string                   `json:"transaction_date,omitempty"`
	SettlementDate  *string                   `json:"settlement_date,omitempty"`
	Status          *domain.TransactionStatus `json:"status,omitempty"`
	Description     *string                   `json:"description,omitempty" binding:"omitempty,max=500"`
	Notes           *string                   `json:"notes,omitempty"`
	Broker          *string                   `json:"broker,omitempty" binding:"omitempty,max=100"`
	Exchange        *string                   `json:"exchange,omitempty" binding:"omitempty,max=50"`
	Tags            *string                   `json:"tags,omitempty"`
}

// ListTransactionsQuery represents query parameters for listing transactions
type ListTransactionsQuery struct {
	AssetID         string  `form:"asset_id"`
	TransactionType string  `form:"transaction_type"`
	Status          string  `form:"status"`
	StartDate       string  `form:"start_date"` // YYYY-MM-DD format
	EndDate         string  `form:"end_date"`   // YYYY-MM-DD format
	MinAmount       float64 `form:"min_amount" binding:"omitempty,min=0"`
	MaxAmount       float64 `form:"max_amount" binding:"omitempty,min=0"`
	Broker          string  `form:"broker"`
	Page            int     `form:"page" binding:"omitempty,min=1"`
	PageSize        int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy          string  `form:"sort_by"` // transaction_date, total_amount, etc.
	SortOrder       string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
