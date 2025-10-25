package dto

import (
	"time"

	"personalfinancedss/internal/module/cashflow/account/domain"
)

// AccountResponse represents an account in API responses.
type AccountResponse struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	AccountName         string     `json:"account_name"`
	AccountType         string     `json:"account_type"`
	InstitutionName     *string    `json:"institution_name,omitempty"`
	CurrentBalance      float64    `json:"current_balance"`
	AvailableBalance    *float64   `json:"available_balance,omitempty"`
	Currency            string     `json:"currency"`
	AccountNumberMasked *string    `json:"account_number_masked,omitempty"`
	IsActive            bool       `json:"is_active"`
	IsPrimary           bool       `json:"is_primary"`
	IncludeInNetWorth   bool       `json:"include_in_net_worth"`
	LastSyncedAt        *time.Time `json:"last_synced_at,omitempty"`
	SyncStatus          *string    `json:"sync_status,omitempty"`
	SyncErrorMessage    *string    `json:"sync_error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ToResponse converts a domain account to DTO response.
func ToResponse(account domain.Account) AccountResponse {
	var syncStatus *string
	if account.SyncStatus != nil {
		status := string(*account.SyncStatus)
		syncStatus = &status
	}

	return AccountResponse{
		ID:                  account.ID.String(),
		UserID:              account.UserID.String(),
		AccountName:         account.AccountName,
		AccountType:         string(account.AccountType),
		InstitutionName:     account.InstitutionName,
		CurrentBalance:      account.CurrentBalance,
		AvailableBalance:    account.AvailableBalance,
		Currency:            string(account.Currency),
		AccountNumberMasked: account.AccountNumberMasked,
		IsActive:            account.IsActive,
		IsPrimary:           account.IsPrimary,
		IncludeInNetWorth:   account.IncludeInNetWorth,
		LastSyncedAt:        account.LastSyncedAt,
		SyncStatus:          syncStatus,
		SyncErrorMessage:    account.SyncErrorMessage,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}

// AccountsListResponse represents a list of accounts.
type AccountsListResponse struct {
	Items []AccountResponse `json:"items"`
	Total int64             `json:"total"`
}
