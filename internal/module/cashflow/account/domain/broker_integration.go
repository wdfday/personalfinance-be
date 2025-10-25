package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// BrokerType represents the type of broker
type BrokerType string

const (
	BrokerTypeSSI BrokerType = "ssi" // SSI Securities (Vietnam stocks)
	BrokerTypeOKX BrokerType = "okx" // OKX Exchange (Crypto)
)

// BrokerIntegration represents broker integration configuration
// Single struct for all broker types (SSI, OKX, etc.)
type BrokerIntegration struct {
	// Broker information
	BrokerType BrokerType `json:"broker_type"`
	BrokerName string     `json:"broker_name,omitempty"`
	IsActive   bool       `json:"is_active"`

	// Token management (encrypted)
	AccessToken     *string    `json:"access_token,omitempty"`
	RefreshToken    *string    `json:"refresh_token,omitempty"`
	TokenExpiresAt  *time.Time `json:"token_expires_at,omitempty"`
	LastRefreshedAt *time.Time `json:"last_refreshed_at,omitempty"`

	// Auto-sync settings
	AutoSync      bool `json:"auto_sync"`
	SyncFrequency int  `json:"sync_frequency"` // minutes (default: 60)

	// Sync statistics
	TotalSyncs      int     `json:"total_syncs"`
	SuccessfulSyncs int     `json:"successful_syncs"`
	FailedSyncs     int     `json:"failed_syncs"`
	LastSyncDetails *string `json:"last_sync_details,omitempty"`

	// Sync options
	SyncAssets       bool `json:"sync_assets"`       // Sync to investment_assets
	SyncTransactions bool `json:"sync_transactions"` // Sync to investment_transactions
	SyncPrices       bool `json:"sync_prices"`       // Update market prices
	SyncBalance      bool `json:"sync_balance"`      // Update account balance

	// OKX-specific credentials (encrypted, only used when BrokerType=okx)
	OKXAPIKey     *string `json:"okx_api_key,omitempty"`
	OKXAPISecret  *string `json:"okx_api_secret,omitempty"`
	OKXPassphrase *string `json:"okx_passphrase,omitempty"`

	// SSI-specific credentials (encrypted, only used when BrokerType=ssi)
	SSIConsumerID     *string `json:"ssi_consumer_id,omitempty"`
	SSIConsumerSecret *string `json:"ssi_consumer_secret,omitempty"`
	SSIOTPMethod      *string `json:"ssi_otp_method,omitempty"` // PIN/SMS/EMAIL/SMART
}

// NewBrokerIntegration creates a new broker integration
func NewBrokerIntegration(brokerType BrokerType) *BrokerIntegration {
	return &BrokerIntegration{
		BrokerType:    brokerType,
		IsActive:      true,
		SyncFrequency: 60, // Default 60 minutes
	}
}

// GetBrokerIntegration parses the JSONB and returns BrokerIntegration
func (a *Account) GetBrokerIntegration() (*BrokerIntegration, error) {
	if a.BrokerIntegration == nil || len(a.BrokerIntegration) == 0 {
		return nil, nil
	}

	var integration BrokerIntegration
	if err := json.Unmarshal(a.BrokerIntegration, &integration); err != nil {
		return nil, err
	}

	return &integration, nil
}

// SetBrokerIntegration sets the broker integration configuration
func (a *Account) SetBrokerIntegration(integration *BrokerIntegration) error {
	if integration == nil {
		a.BrokerIntegration = nil
		return nil
	}

	data, err := json.Marshal(integration)
	if err != nil {
		return err
	}

	a.BrokerIntegration = data
	return nil
}

// HasBrokerIntegration checks if account has broker integration configured
func (a *Account) HasBrokerIntegration() bool {
	return a.BrokerIntegration != nil && len(a.BrokerIntegration) > 0
}

// IsBrokerActive checks if broker integration is active
func (a *Account) IsBrokerActive() bool {
	integration, err := a.GetBrokerIntegration()
	if err != nil || integration == nil {
		return false
	}
	return integration.IsActive
}

// NeedsSync checks if the account needs syncing
func (a *Account) NeedsSync() bool {
	integration, err := a.GetBrokerIntegration()
	if err != nil || integration == nil {
		return false
	}

	if !integration.AutoSync || !integration.IsActive {
		return false
	}

	if a.LastSyncedAt == nil {
		return true
	}

	nextSync := a.LastSyncedAt.Add(time.Duration(integration.SyncFrequency) * time.Minute)
	return time.Now().After(nextSync)
}

// UpdateSyncStatus updates the sync status after a sync operation
func (a *Account) UpdateSyncStatus(success bool, errorMsg *string) error {
	integration, err := a.GetBrokerIntegration()
	if err != nil || integration == nil {
		return err
	}

	now := time.Now()
	a.LastSyncedAt = &now
	integration.TotalSyncs++

	if success {
		integration.SuccessfulSyncs++
		status := SyncStatusActive
		a.SyncStatus = &status
		a.SyncErrorMessage = nil
	} else {
		integration.FailedSyncs++
		status := SyncStatusError
		a.SyncStatus = &status
		a.SyncErrorMessage = errorMsg

		// Deactivate after 3 consecutive failures
		if integration.FailedSyncs >= 3 {
			integration.IsActive = false
			disconnected := SyncStatusDisconnected
			a.SyncStatus = &disconnected
		}
	}

	return a.SetBrokerIntegration(integration)
}

// RefreshAccessToken updates the access token
func (a *Account) RefreshAccessToken(token string, expiresIn int) error {
	integration, err := a.GetBrokerIntegration()
	if err != nil || integration == nil {
		return err
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(expiresIn) * time.Second)
	integration.AccessToken = &token
	integration.TokenExpiresAt = &expiresAt
	integration.LastRefreshedAt = &now

	return a.SetBrokerIntegration(integration)
}

// IsTokenValid checks if the access token is still valid
func (a *Account) IsTokenValid() bool {
	integration, err := a.GetBrokerIntegration()
	if err != nil || integration == nil {
		return false
	}

	if integration.TokenExpiresAt == nil {
		return false
	}

	return time.Now().Before(*integration.TokenExpiresAt)
}

// GetCredentials returns credentials based on broker type (for internal use)
func (i *BrokerIntegration) GetCredentials() map[string]*string {
	credentials := make(map[string]*string)

	switch i.BrokerType {
	case BrokerTypeOKX:
		credentials["api_key"] = i.OKXAPIKey
		credentials["api_secret"] = i.OKXAPISecret
		credentials["passphrase"] = i.OKXPassphrase
	case BrokerTypeSSI:
		credentials["consumer_id"] = i.SSIConsumerID
		credentials["consumer_secret"] = i.SSIConsumerSecret
		credentials["otp_method"] = i.SSIOTPMethod
	}

	return credentials
}

// Validate checks if the integration has required fields
func (i *BrokerIntegration) Validate() error {
	switch i.BrokerType {
	case BrokerTypeOKX:
		if i.OKXAPIKey == nil || *i.OKXAPIKey == "" {
			return fmt.Errorf("OKX API key is required")
		}
		if i.OKXAPISecret == nil || *i.OKXAPISecret == "" {
			return fmt.Errorf("OKX API secret is required")
		}
		if i.OKXPassphrase == nil || *i.OKXPassphrase == "" {
			return fmt.Errorf("OKX passphrase is required")
		}
	case BrokerTypeSSI:
		if i.SSIConsumerID == nil || *i.SSIConsumerID == "" {
			return fmt.Errorf("SSI consumer ID is required")
		}
		if i.SSIConsumerSecret == nil || *i.SSIConsumerSecret == "" {
			return fmt.Errorf("SSI consumer secret is required")
		}
	default:
		return fmt.Errorf("unsupported broker type: %s", i.BrokerType)
	}

	return nil
}
