package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvestmentTransaction represents a transaction for buying/selling investment assets
type InvestmentTransaction struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	UserID  uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`
	AssetID uuid.UUID `gorm:"type:uuid;not null;index;column:asset_id" json:"asset_id"`

	// Transaction details
	TransactionType TransactionType `gorm:"type:varchar(20);not null;index;column:transaction_type" json:"transaction_type"`
	Quantity        float64         `gorm:"type:decimal(20,8);not null;column:quantity" json:"quantity"`
	PricePerUnit    float64         `gorm:"type:decimal(15,2);not null;column:price_per_unit" json:"price_per_unit"`
	TotalAmount     float64         `gorm:"type:decimal(15,2);not null;column:total_amount" json:"total_amount"`
	Currency        string          `gorm:"type:varchar(3);not null;default:'USD';column:currency" json:"currency"`

	// Fees and costs
	Fees       float64 `gorm:"type:decimal(15,2);default:0;column:fees" json:"fees"`
	Commission float64 `gorm:"type:decimal(15,2);default:0;column:commission" json:"commission"`
	Tax        float64 `gorm:"type:decimal(15,2);default:0;column:tax" json:"tax"`
	TotalCost  float64 `gorm:"type:decimal(15,2);not null;column:total_cost" json:"total_cost"` // TotalAmount + Fees + Commission + Tax

	// For sell transactions
	RealizedGain    *float64 `gorm:"type:decimal(15,2);column:realized_gain" json:"realized_gain,omitempty"`
	RealizedGainPct *float64 `gorm:"type:decimal(10,4);column:realized_gain_pct" json:"realized_gain_pct,omitempty"`

	// Transaction metadata
	TransactionDate time.Time         `gorm:"not null;index;column:transaction_date" json:"transaction_date"`
	SettlementDate  *time.Time        `gorm:"column:settlement_date" json:"settlement_date,omitempty"`
	Status          TransactionStatus `gorm:"type:varchar(20);not null;default:'completed';column:status" json:"status"`

	// Description and notes
	Description string  `gorm:"type:varchar(500);column:description" json:"description,omitempty"`
	Notes       *string `gorm:"type:text;column:notes" json:"notes,omitempty"`

	// Broker/Exchange information
	Broker     *string `gorm:"type:varchar(100);column:broker" json:"broker,omitempty"`
	Exchange   *string `gorm:"type:varchar(50);column:exchange" json:"exchange,omitempty"`
	OrderID    *string `gorm:"type:varchar(100);column:order_id" json:"order_id,omitempty"`
	ExternalID *string `gorm:"type:varchar(255);index;column:external_id" json:"external_id,omitempty"`

	// Additional fields
	Tags *string `gorm:"type:text;column:tags" json:"tags,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the database table name
func (InvestmentTransaction) TableName() string {
	return "investment_transactions"
}

// CalculateTotalCost calculates the total cost including fees
func (t *InvestmentTransaction) CalculateTotalCost() {
	t.TotalCost = t.TotalAmount + t.Fees + t.Commission + t.Tax
}

// IsBuy returns true if this is a buy transaction
func (t *InvestmentTransaction) IsBuy() bool {
	return t.TransactionType == TransactionTypeBuy
}

// IsSell returns true if this is a sell transaction
func (t *InvestmentTransaction) IsSell() bool {
	return t.TransactionType == TransactionTypeSell
}

// IsDividend returns true if this is a dividend transaction
func (t *InvestmentTransaction) IsDividend() bool {
	return t.TransactionType == TransactionTypeDividend
}

// IsCompleted returns true if the transaction is completed
func (t *InvestmentTransaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}
