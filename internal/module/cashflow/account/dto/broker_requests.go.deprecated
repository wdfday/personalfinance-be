package dto

import (
	"errors"
	"personalfinancedss/internal/module/cashflow/account/domain"
)

// CreateAccountWithBrokerRequest represents a request to create account with broker integration
type CreateAccountWithBrokerRequest struct {
	// Account info
	AccountName     string  `json:"account_name" binding:"required,min=1,max=255"`
	AccountType     string  `json:"account_type" binding:"required,oneof=investment crypto_wallet"`
	InstitutionName *string `json:"institution_name,omitempty" binding:"omitempty,max=255"`

	// Broker info
	BrokerType domain.BrokerType `json:"broker_type" binding:"required"`
	BrokerName string            `json:"broker_name,omitempty"`

	// SSI credentials
	ConsumerID     *string `json:"consumer_id,omitempty"`
	ConsumerSecret *string `json:"consumer_secret,omitempty"`
	OTPCode        *string `json:"otp_code,omitempty"`   // For initial authentication
	OTPMethod      *string `json:"otp_method,omitempty"` // PIN/SMS/EMAIL/SMART

	// OKX credentials
	APIKey     *string `json:"api_key,omitempty"`
	APISecret  *string `json:"api_secret,omitempty"`
	Passphrase *string `json:"passphrase,omitempty"`

	// Sync settings (optional)
	AutoSync         bool `json:"auto_sync"`
	SyncFrequency    int  `json:"sync_frequency,omitempty"` // minutes, default 60
	SyncAssets       bool `json:"sync_assets"`
	SyncTransactions bool `json:"sync_transactions"`
	SyncPrices       bool `json:"sync_prices"`
	SyncBalance      bool `json:"sync_balance"`
}

// Validate validates the create account with broker request
func (r *CreateAccountWithBrokerRequest) Validate() error {
	// Validate broker type
	if r.BrokerType != domain.BrokerTypeSSI && r.BrokerType != domain.BrokerTypeOKX {
		return errors.New("broker_type must be 'ssi' or 'okx'")
	}

	// Validate SSI credentials
	if r.BrokerType == domain.BrokerTypeSSI {
		if r.ConsumerID == nil || *r.ConsumerID == "" {
			return errors.New("consumer_id is required for SSI")
		}
		if r.ConsumerSecret == nil || *r.ConsumerSecret == "" {
			return errors.New("consumer_secret is required for SSI")
		}
	}

	// Validate OKX credentials
	if r.BrokerType == domain.BrokerTypeOKX {
		if r.APIKey == nil || *r.APIKey == "" {
			return errors.New("api_key is required for OKX")
		}
		if r.APISecret == nil || *r.APISecret == "" {
			return errors.New("api_secret is required for OKX")
		}
		if r.Passphrase == nil || *r.Passphrase == "" {
			return errors.New("passphrase is required for OKX")
		}
	}

	// Validate sync frequency
	if r.AutoSync && r.SyncFrequency < 1 {
		return errors.New("sync_frequency must be at least 1 minute when auto_sync is enabled")
	}

	return nil
}
