package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Account represents a user's financial account.
type Account struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;column:user_id" json:"user_id"`

	AccountName     string      `gorm:"type:varchar(255);not null;column:account_name" json:"account_name"`
	AccountType     AccountType `gorm:"type:varchar(50);not null;column:account_type" json:"account_type"`
	InstitutionName *string     `gorm:"type:varchar(255);column:institution_name" json:"institution_name,omitempty"`

	CurrentBalance   float64  `gorm:"type:decimal(15,2);not null;default:0;column:current_balance" json:"current_balance"`
	AvailableBalance *float64 `gorm:"type:decimal(15,2);column:available_balance" json:"available_balance,omitempty"`
	Currency         Currency `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	AccountNumberMasked    *string `gorm:"type:varchar(50);column:account_number_masked" json:"account_number_masked,omitempty"`
	AccountNumberEncrypted *string `gorm:"type:text;column:account_number_encrypted" json:"-"`

	IsActive          bool `gorm:"default:true;column:is_active" json:"is_active"`
	IsPrimary         bool `gorm:"default:false;column:is_primary" json:"is_primary"`
	IncludeInNetWorth bool `gorm:"default:true;column:include_in_net_worth" json:"include_in_net_worth"`

	// IsAutoSync, true for investment - crypto - which has api, false for others
	IsAutoSync       bool        `gorm:"default:false;column:is_auto_sync" json:"is_auto_sync"`
	LastSyncedAt     *time.Time  `gorm:"column:last_synced_at" json:"last_synced_at,omitempty"`
	SyncStatus       *SyncStatus `gorm:"type:varchar(20);column:sync_status" json:"sync_status,omitempty"`
	SyncErrorMessage *string     `gorm:"type:text;column:sync_error_message" json:"sync_error_message,omitempty"`

	// Broker Integration - New approach: Reference to broker_connections table
	BrokerConnectionID *uuid.UUID `gorm:"type:uuid;column:broker_connection_id;index" json:"broker_connection_id,omitempty"`

	// Broker Integration (DEPRECATED - for backward compatibility)
	// Will be migrated to broker_connections table
	BrokerIntegration datatypes.JSON `gorm:"type:jsonb;column:broker_integration" json:"broker_integration,omitempty"`

	// DSS Metadata for Analytics & Forecasting
	DSSMetadata datatypes.JSON `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`
	// Structure:
	// {
	//   "opening_balance": 50000000,          // Initial balance
	//   "opening_balance_date": "2024-01-01",
	//   "highest_balance": 75000000,          // Highest balance reached
	//   "highest_balance_date": "2024-06-15",
	//   "lowest_balance": 30000000,           // Lowest balance reached
	//   "lowest_balance_date": "2024-03-10",
	//   "avg_monthly_balance": 55000000,      // Average monthly balance
	//   "balance_volatility": 0.15,           // Std dev / mean
	//   "balance_trend": "increasing",        // increasing, stable, decreasing
	//   "transaction_frequency": 45,          // Avg transactions per month
	//   "avg_transaction_amount": 500000,     // Average transaction size
	//   "inflow_total": 60000000,             // Total inflows (monthly)
	//   "outflow_total": 55000000,            // Total outflows (monthly)
	//   "net_cashflow": 5000000,              // Net monthly cashflow
	//   "cashflow_trend": "positive",         // positive, negative, neutral
	//   "days_to_zero": 180,                  // Days until balance reaches zero (forecast)
	//   "risk_level": "low",                  // low, medium, high
	//   "health_score": 0.85,                 // Overall account health (0-1)
	//   "last_analyzed": "2024-01-15T10:00:00Z"
	// }

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName matches the database table.
func (Account) TableName() string {
	return "accounts"
}
