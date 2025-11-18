package dto

import (
	"personalfinancedss/internal/module/identify/broker/domain"
)

// CreateBrokerConnectionRequest represents the API request to create a broker connection
type CreateBrokerConnectionRequest struct {
	BrokerType string `json:"broker_type" binding:"required" example:"ssi"` // ssi, okx, sepay
	BrokerName string `json:"broker_name" binding:"required" example:"SSI Securities"`

	// Credentials (broker-specific)
	APIKey         string  `json:"api_key,omitempty"`
	APISecret      string  `json:"api_secret,omitempty"`
	Passphrase     *string `json:"passphrase,omitempty"`      // For OKX
	ConsumerID     *string `json:"consumer_id,omitempty"`     // For SSI
	ConsumerSecret *string `json:"consumer_secret,omitempty"` // For SSI
	OTPCode        *string `json:"otp_code,omitempty"`        // For SSI initial auth
	OTPMethod      *string `json:"otp_method,omitempty"`      // For SSI: PIN/SMS/EMAIL/SMART

	// Sync settings (optional, defaults will be applied)
	AutoSync         *bool `json:"auto_sync,omitempty" example:"true"`
	SyncFrequency    *int  `json:"sync_frequency,omitempty" example:"60"` // minutes
	SyncAssets       *bool `json:"sync_assets,omitempty" example:"true"`
	SyncTransactions *bool `json:"sync_transactions,omitempty" example:"true"`
	SyncPrices       *bool `json:"sync_prices,omitempty" example:"true"`
	SyncBalance      *bool `json:"sync_balance,omitempty" example:"true"`

	// Optional
	Notes *string `json:"notes,omitempty"`
}

// UpdateBrokerConnectionRequest represents the API request to update a broker connection
type UpdateBrokerConnectionRequest struct {
	BrokerName *string `json:"broker_name,omitempty"`

	// Update credentials (optional)
	APIKey         *string `json:"api_key,omitempty"`
	APISecret      *string `json:"api_secret,omitempty"`
	Passphrase     *string `json:"passphrase,omitempty"`
	ConsumerID     *string `json:"consumer_id,omitempty"`
	ConsumerSecret *string `json:"consumer_secret,omitempty"`
	OTPMethod      *string `json:"otp_method,omitempty"`

	// Sync settings
	AutoSync         *bool `json:"auto_sync,omitempty"`
	SyncFrequency    *int  `json:"sync_frequency,omitempty"`
	SyncAssets       *bool `json:"sync_assets,omitempty"`
	SyncTransactions *bool `json:"sync_transactions,omitempty"`
	SyncPrices       *bool `json:"sync_prices,omitempty"`
	SyncBalance      *bool `json:"sync_balance,omitempty"`

	Notes *string `json:"notes,omitempty"`
}

// ListBrokerConnectionsQuery represents query parameters for listing connections
type ListBrokerConnectionsQuery struct {
	BrokerType      *string `form:"broker_type"`       // Filter by broker type
	Status          *string `form:"status"`            // Filter by status
	AutoSyncOnly    bool    `form:"auto_sync_only"`    // Only return auto-sync enabled
	ActiveOnly      bool    `form:"active_only"`       // Only return active connections
	NeedingSyncOnly bool    `form:"needing_sync_only"` // Only return connections needing sync
}

// Validate validates the create request
func (r *CreateBrokerConnectionRequest) Validate() error {
	// Convert string to BrokerType enum
	brokerType := domain.BrokerType(r.BrokerType)

	switch brokerType {
	case domain.BrokerTypeSSI:
		if r.ConsumerID == nil || r.ConsumerSecret == nil {
			return ErrSSICredentialsRequired
		}
	case domain.BrokerTypeOKX:
		if r.APIKey == "" || r.APISecret == "" || r.Passphrase == nil {
			return ErrOKXCredentialsRequired
		}
	case domain.BrokerTypeSePay:
		if r.APIKey == "" {
			return ErrSepayCredentialsRequired
		}
	default:
		return ErrInvalidBrokerType
	}

	return nil
}

// ApplyDefaults applies default values to optional fields
func (r *CreateBrokerConnectionRequest) ApplyDefaults() {
	if r.AutoSync == nil {
		autoSync := true
		r.AutoSync = &autoSync
	}

	if r.SyncFrequency == nil {
		syncFreq := 60 // 60 minutes
		r.SyncFrequency = &syncFreq
	}

	if r.SyncAssets == nil {
		syncAssets := true
		r.SyncAssets = &syncAssets
	}

	if r.SyncTransactions == nil {
		syncTx := true
		r.SyncTransactions = &syncTx
	}

	if r.SyncPrices == nil {
		syncPrices := true
		r.SyncPrices = &syncPrices
	}

	if r.SyncBalance == nil {
		syncBalance := true
		r.SyncBalance = &syncBalance
	}
}
