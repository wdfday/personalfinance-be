package dto

import (
	"time"

	"github.com/google/uuid"
)

// BrokerConnectionResponse represents a broker connection in API responses
type BrokerConnectionResponse struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	BrokerType string    `json:"broker_type" example:"ssi"` // ssi, okx, sepay
	BrokerName string    `json:"broker_name" example:"SSI Securities"`
	Status     string    `json:"status" example:"active"` // active, disconnected, error, pending

	// Token info (NOT sensitive values, just metadata)
	TokenExpiresAt  *time.Time `json:"token_expires_at,omitempty"`
	LastRefreshedAt *time.Time `json:"last_refreshed_at,omitempty"`
	IsTokenValid    bool       `json:"is_token_valid"`

	// Sync settings
	AutoSync         bool `json:"auto_sync" example:"true"`
	SyncFrequency    int  `json:"sync_frequency" example:"60"` // minutes
	SyncAssets       bool `json:"sync_assets" example:"true"`
	SyncTransactions bool `json:"sync_transactions" example:"true"`
	SyncPrices       bool `json:"sync_prices" example:"true"`
	SyncBalance      bool `json:"sync_balance" example:"true"`

	// Sync status
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus *string    `json:"last_sync_status,omitempty"` // success, failed
	LastSyncError  *string    `json:"last_sync_error,omitempty"`

	// Sync statistics
	TotalSyncs      int `json:"total_syncs" example:"10"`
	SuccessfulSyncs int `json:"successful_syncs" example:"9"`
	FailedSyncs     int `json:"failed_syncs" example:"1"`

	// External account info
	ExternalAccountID     *string `json:"external_account_id,omitempty"`
	ExternalAccountNumber *string `json:"external_account_number,omitempty"`
	ExternalAccountName   *string `json:"external_account_name,omitempty"`

	// Metadata
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BrokerConnectionListResponse represents a list of broker connections
type BrokerConnectionListResponse struct {
	Connections []*BrokerConnectionResponse `json:"connections"`
	Total       int                         `json:"total"`
}

// SyncResultResponse represents the result of a sync operation
type SyncResultResponse struct {
	Success            bool                   `json:"success"`
	SyncedAt           time.Time              `json:"synced_at"`
	AssetsCount        int                    `json:"assets_count"`
	TransactionsCount  int                    `json:"transactions_count"`
	UpdatedPricesCount int                    `json:"updated_prices_count"`
	BalanceUpdated     bool                   `json:"balance_updated"`
	Error              *string                `json:"error,omitempty"`
	Details            map[string]interface{} `json:"details,omitempty"`
}

// BrokerProviderInfo represents information about a broker provider
type BrokerProviderInfo struct {
	BrokerType        string   `json:"broker_type" example:"ssi"`
	DisplayName       string   `json:"display_name" example:"SSI Securities"`
	Description       string   `json:"description" example:"Vietnam stock trading platform"`
	RequiredFields    []string `json:"required_fields" example:"consumer_id,consumer_secret,otp_code"`
	SupportedFeatures []string `json:"supported_features" example:"portfolio,transactions,market_prices"`
	Logo              string   `json:"logo,omitempty"`
}

// ListBrokerProvidersResponse represents available broker providers
type ListBrokerProvidersResponse struct {
	Providers []BrokerProviderInfo `json:"providers"`
}
