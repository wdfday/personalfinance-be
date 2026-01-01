package dto

import (
	"personalfinancedss/internal/module/identify/broker/domain"

	"github.com/google/uuid"
)

// ============================================================================
// Base Common Fields
// ============================================================================

// BaseBrokerConnection contains common fields for all broker connections
type BaseBrokerConnection struct {
	BrokerName string `json:"broker_name" binding:"required" example:"SSI Securities"`

	// Sync settings (optional, defaults will be applied)
	AutoSync         *bool `json:"auto_sync,omitempty" example:"true"`
	SyncFrequency    *int  `json:"sync_frequency,omitempty" example:"60"` // minutes
	SyncAssets       *bool `json:"sync_assets,omitempty" example:"true"`
	SyncTransactions *bool `json:"sync_transactions,omitempty" example:"true"`
	SyncPrices       *bool `json:"sync_prices,omitempty" example:"true"`
	SyncBalance      *bool `json:"sync_balance,omitempty" example:"true"`

	Notes *string `json:"notes,omitempty" example:"My broker account"`
}

// ApplyDefaults applies default values to sync settings
func (b *BaseBrokerConnection) ApplyDefaults() {
	if b.AutoSync == nil {
		autoSync := true
		b.AutoSync = &autoSync
	}
	if b.SyncFrequency == nil {
		syncFreq := 60
		b.SyncFrequency = &syncFreq
	}
	if b.SyncAssets == nil {
		syncAssets := true
		b.SyncAssets = &syncAssets
	}
	if b.SyncTransactions == nil {
		syncTx := true
		b.SyncTransactions = &syncTx
	}
	if b.SyncPrices == nil {
		syncPrices := true
		b.SyncPrices = &syncPrices
	}
	if b.SyncBalance == nil {
		syncBalance := true
		b.SyncBalance = &syncBalance
	}
}

// ============================================================================
// SSI Broker
// ============================================================================

// CreateSSIConnectionRequest represents request to create SSI broker connection
type CreateSSIConnectionRequest struct {
	BaseBrokerConnection

	// SSI-specific credentials
	ConsumerID     string  `json:"consumer_id" binding:"required" example:"your-consumer-id"`
	ConsumerSecret string  `json:"consumer_secret" binding:"required" example:"your-consumer-secret"`
	OTPCode        *string `json:"otp_code,omitempty" example:"123456"`
	OTPMethod      *string `json:"otp_method,omitempty" example:"SMART"` // PIN/SMS/EMAIL/SMART
}

// ============================================================================
// OKX Broker
// ============================================================================

// CreateOKXConnectionRequest represents request to create OKX broker connection
type CreateOKXConnectionRequest struct {
	BaseBrokerConnection

	// OKX-specific credentials
	APIKey     string `json:"api_key" binding:"required" example:"your-api-key"`
	APISecret  string `json:"api_secret" binding:"required" example:"your-api-secret"`
	Passphrase string `json:"passphrase" binding:"required" example:"your-passphrase"`
}

// ============================================================================
// SePay Broker
// ============================================================================

// CreateSepayConnectionRequest represents request to create SePay broker connection
type CreateSepayConnectionRequest struct {
	BaseBrokerConnection

	// SePay-specific credentials
	APIKey string `json:"api_key" binding:"required" example:"your-sepay-api-key"`
}

// ============================================================================
// Update Requests (keep unified for simplicity)
// ============================================================================

// UpdateBrokerConnectionRequest represents the API request to update a broker connection
type UpdateBrokerConnectionRequest struct {
	BrokerName *string `json:"broker_name,omitempty"`

	// Update credentials (optional, broker-specific)
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

// ============================================================================
// Query Parameters
// ============================================================================

// ListBrokerConnectionsQuery represents query parameters for listing connections
type ListBrokerConnectionsQuery struct {
	BrokerType      *string `form:"broker_type"`       // Filter by broker type
	Status          *string `form:"status"`            // Filter by status
	AutoSyncOnly    bool    `form:"auto_sync_only"`    // Only return auto-sync enabled
	ActiveOnly      bool    `form:"active_only"`       // Only return active connections
	NeedingSyncOnly bool    `form:"needing_sync_only"` // Only return connections needing sync
}

// ToServiceFilters converts query to service filters - defined in responses.go conversion functions

// ============================================================================
// Conversion to Service Layer Request
// ============================================================================

// ToServiceRequest converts SSI DTO to service request
func (r *CreateSSIConnectionRequest) ToServiceRequest(userID uuid.UUID) *CreateBrokerConnectionServiceRequest {
	r.ApplyDefaults()
	return &CreateBrokerConnectionServiceRequest{
		UserID:     userID,
		BrokerType: domain.BrokerTypeSSI,
		BrokerName: r.BrokerName,

		// Credentials
		ConsumerID:     &r.ConsumerID,
		ConsumerSecret: &r.ConsumerSecret,
		OTPCode:        r.OTPCode,
		OTPMethod:      r.OTPMethod,

		// Sync settings
		AutoSync:         *r.AutoSync,
		SyncFrequency:    *r.SyncFrequency,
		SyncAssets:       *r.SyncAssets,
		SyncTransactions: *r.SyncTransactions,
		SyncPrices:       *r.SyncPrices,
		SyncBalance:      *r.SyncBalance,

		Notes: r.Notes,
	}
}

// ToServiceRequest converts OKX DTO to service request
func (r *CreateOKXConnectionRequest) ToServiceRequest(userID uuid.UUID) *CreateBrokerConnectionServiceRequest {
	r.ApplyDefaults()
	return &CreateBrokerConnectionServiceRequest{
		UserID:     userID,
		BrokerType: domain.BrokerTypeOKX,
		BrokerName: r.BrokerName,

		// Credentials
		APIKey:     r.APIKey,
		APISecret:  r.APISecret,
		Passphrase: &r.Passphrase,

		// Sync settings
		AutoSync:         *r.AutoSync,
		SyncFrequency:    *r.SyncFrequency,
		SyncAssets:       *r.SyncAssets,
		SyncTransactions: *r.SyncTransactions,
		SyncPrices:       *r.SyncPrices,
		SyncBalance:      *r.SyncBalance,

		Notes: r.Notes,
	}
}

// ToServiceRequest converts SePay DTO to service request
func (r *CreateSepayConnectionRequest) ToServiceRequest(userID uuid.UUID) *CreateBrokerConnectionServiceRequest {
	r.ApplyDefaults()
	return &CreateBrokerConnectionServiceRequest{
		UserID:     userID,
		BrokerType: domain.BrokerTypeSePay,
		BrokerName: r.BrokerName,

		// Credentials
		APIKey: r.APIKey,

		// Sync settings
		AutoSync:         *r.AutoSync,
		SyncFrequency:    *r.SyncFrequency,
		SyncAssets:       *r.SyncAssets,
		SyncTransactions: *r.SyncTransactions,
		SyncPrices:       *r.SyncPrices,
		SyncBalance:      *r.SyncBalance,

		Notes: r.Notes,
	}
}

// ============================================================================
// Service Layer Request (unified internal representation)
// ============================================================================

// CreateBrokerConnectionServiceRequest is the internal service layer request
type CreateBrokerConnectionServiceRequest struct {
	UserID     uuid.UUID
	BrokerType domain.BrokerType
	BrokerName string

	// All possible credentials (broker-specific fields will be nil/empty)
	APIKey         string
	APISecret      string
	Passphrase     *string
	ConsumerID     *string
	ConsumerSecret *string
	OTPCode        *string
	OTPMethod      *string

	// Sync settings
	AutoSync         bool
	SyncFrequency    int
	SyncAssets       bool
	SyncTransactions bool
	SyncPrices       bool
	SyncBalance      bool

	Notes *string
}
