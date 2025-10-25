package dto

// CreateAccountRequest represents data for creating a new account.
type CreateAccountRequest struct {
	AccountName         string   `json:"account_name" binding:"required,min=1,max=255"`
	AccountType         string   `json:"account_type" binding:"required,oneof=cash bank savings credit_card investment crypto_wallet"`
	InstitutionName     *string  `json:"institution_name,omitempty" binding:"omitempty,max=255"`
	CurrentBalance      *float64 `json:"current_balance,omitempty"`
	AvailableBalance    *float64 `json:"available_balance,omitempty"`
	Currency            *string  `json:"currency,omitempty" binding:"omitempty,len=3"`
	AccountNumberMasked *string  `json:"account_number_masked,omitempty" binding:"omitempty,max=50"`
	IsActive            *bool    `json:"is_active,omitempty"`
	IsPrimary           *bool    `json:"is_primary,omitempty"`
	IncludeInNetWorth   *bool    `json:"include_in_net_worth,omitempty"`
}

// UpdateAccountRequest represents data for updating an account.
type UpdateAccountRequest struct {
	AccountName         *string  `json:"account_name,omitempty" binding:"omitempty,min=1,max=255"`
	AccountType         *string  `json:"account_type,omitempty" binding:"omitempty,oneof=cash bank savings credit_card investment crypto_wallet"`
	InstitutionName     *string  `json:"institution_name,omitempty" binding:"omitempty,max=255"`
	CurrentBalance      *float64 `json:"current_balance,omitempty"`
	AvailableBalance    *float64 `json:"available_balance,omitempty"`
	Currency            *string  `json:"currency,omitempty" binding:"omitempty,len=3"`
	AccountNumberMasked *string  `json:"account_number_masked,omitempty" binding:"omitempty,max=50"`
	IsActive            *bool    `json:"is_active,omitempty"`
	IsPrimary           *bool    `json:"is_primary,omitempty"`
	IncludeInNetWorth   *bool    `json:"include_in_net_worth,omitempty"`
	SyncStatus          *string  `json:"sync_status,omitempty" binding:"omitempty,oneof=active error disconnected"`
	SyncErrorMessage    *string  `json:"sync_error_message,omitempty"`
}

// ListAccountsRequest represents query parameters for listing accounts.
type ListAccountsRequest struct {
	AccountType    *string `form:"account_type" binding:"omitempty,oneof=cash bank savings credit_card investment crypto_wallet"`
	IsActive       *bool   `form:"is_active"`
	IsPrimary      *bool   `form:"is_primary"`
	IncludeDeleted *bool   `form:"include_deleted"`
}
