package dto

import "time"

// CreateTransactionRequest represents request to create a new transaction
type CreateTransactionRequest struct {
	// Core transaction fields
	AccountID  string `json:"accountId" binding:"required,uuid"` // FK to account (bank account / e-wallet / cash account)
	Direction  string `json:"direction" binding:"required,oneof=DEBIT CREDIT"`
	Instrument string `json:"instrument" binding:"required,oneof=CASH BANK_ACCOUNT DEBIT_CARD CREDIT_CARD E_WALLET CRYPTO UNKNOWN"`
	Source     string `json:"source" binding:"required,oneof=BANK_API CSV_IMPORT JSON_IMPORT MANUAL"`

	// Amount (in smallest currency unit, e.g., VND = dong)
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"omitempty,len=3"` // Default: VND

	// Timestamps
	BookingDate time.Time  `json:"bookingDate" binding:"required"`
	ValueDate   *time.Time `json:"valueDate,omitempty"` // If not provided, defaults to BookingDate

	// Description fields
	Description string `json:"description,omitempty" binding:"omitempty,max=500"` // Technical description from bank/import
	UserNote    string `json:"userNote,omitempty" binding:"omitempty,max=1000"`   // User's personal note
	Reference   string `json:"reference,omitempty"`                               // Bank reference code

	// Bank-specific fields (optional, mainly for BANK_API or imports)
	BankCode   string `json:"bankCode,omitempty"`   // e.g., "TCB", "VCB"
	ExternalID string `json:"externalId,omitempty"` // External transaction ID
	Channel    string `json:"channel,omitempty" binding:"omitempty,oneof=MOBILE_APP INTERNET_BANKING ATM POS UNKNOWN"`

	// Running balance after transaction (optional, usually from bank)
	RunningBalance *int64 `json:"runningBalance,omitempty"`

	// Counterparty information (optional)
	CounterpartyName          string `json:"counterpartyName,omitempty"`
	CounterpartyAccountNumber string `json:"counterpartyAccountNumber,omitempty"`
	CounterpartyBankName      string `json:"counterpartyBankName,omitempty"`
	CounterpartyType          string `json:"counterpartyType,omitempty" binding:"omitempty,oneof=MERCHANT PERSON INTERNAL UNKNOWN"`

	// Classification (optional)
	UserCategoryID string `json:"userCategoryId,omitempty" binding:"omitempty,uuid"`

	// Links to other entities (optional)
	Links []TransactionLinkDTO `json:"links,omitempty"`

	// Metadata (optional)
	CheckImageAvailability string `json:"checkImageAvailability,omitempty"`
}

// UpdateTransactionRequest represents request to update an existing transaction
type UpdateTransactionRequest struct {
	// Core transaction fields (all optional for updates)
	AccountID  *string `json:"accountId,omitempty" binding:"omitempty,uuid"`
	Direction  *string `json:"direction,omitempty" binding:"omitempty,oneof=DEBIT CREDIT"`
	Instrument *string `json:"instrument,omitempty" binding:"omitempty,oneof=CASH BANK_ACCOUNT DEBIT_CARD CREDIT_CARD E_WALLET CRYPTO UNKNOWN"`
	Source     *string `json:"source,omitempty" binding:"omitempty,oneof=BANK_API CSV_IMPORT JSON_IMPORT MANUAL"`

	// Amount
	Amount   *int64  `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Currency *string `json:"currency,omitempty" binding:"omitempty,len=3"`

	// Timestamps
	BookingDate *time.Time `json:"bookingDate,omitempty"`
	ValueDate   *time.Time `json:"valueDate,omitempty"`

	// Description fields
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	UserNote    *string `json:"userNote,omitempty" binding:"omitempty,max=1000"`
	Reference   *string `json:"reference,omitempty"`

	// Bank-specific fields
	BankCode   *string `json:"bankCode,omitempty"`
	ExternalID *string `json:"externalId,omitempty"`
	Channel    *string `json:"channel,omitempty" binding:"omitempty,oneof=MOBILE_APP INTERNET_BANKING ATM POS UNKNOWN"`

	// Running balance
	RunningBalance *int64 `json:"runningBalance,omitempty"`

	// Counterparty information
	CounterpartyName          *string `json:"counterpartyName,omitempty"`
	CounterpartyAccountNumber *string `json:"counterpartyAccountNumber,omitempty"`
	CounterpartyBankName      *string `json:"counterpartyBankName,omitempty"`
	CounterpartyType          *string `json:"counterpartyType,omitempty" binding:"omitempty,oneof=MERCHANT PERSON INTERNAL UNKNOWN"`

	// Classification
	UserCategoryID *string `json:"userCategoryId,omitempty" binding:"omitempty,uuid"`

	// Links
	Links *[]TransactionLinkDTO `json:"links,omitempty"`

	// Metadata
	CheckImageAvailability *string `json:"checkImageAvailability,omitempty"`
}

// TransactionLinkDTO represents a link to another financial entity
type TransactionLinkDTO struct {
	Type string `json:"type" binding:"required,oneof=GOAL BUDGET DEBT INCOME_PROFILE"` // GOAL / BUDGET / DEBT / INCOME_PROFILE
	ID   string `json:"id" binding:"required,uuid"`                                    // Entity ID
}

// ListTransactionsQuery represents query parameters for listing transactions
type ListTransactionsQuery struct {
	// Account filter
	AccountID *string `form:"accountId" binding:"omitempty,uuid"`

	// Transaction type filters
	Direction  *string `form:"direction" binding:"omitempty,oneof=DEBIT CREDIT"`
	Instrument *string `form:"instrument" binding:"omitempty,oneof=CASH BANK_ACCOUNT DEBIT_CARD CREDIT_CARD E_WALLET CRYPTO UNKNOWN"`
	Source     *string `form:"source" binding:"omitempty,oneof=BANK_API CSV_IMPORT JSON_IMPORT MANUAL"`

	// Bank filters
	BankCode *string `form:"bankCode"`

	// Date range filters (booking date)
	StartBookingDate *time.Time `form:"startBookingDate" time_format:"2006-01-02"`
	EndBookingDate   *time.Time `form:"endBookingDate" time_format:"2006-01-02"`

	// Date range filters (value date)
	StartValueDate *time.Time `form:"startValueDate" time_format:"2006-01-02"`
	EndValueDate   *time.Time `form:"endValueDate" time_format:"2006-01-02"`

	// Amount range filters
	MinAmount *int64 `form:"minAmount" binding:"omitempty,gte=0"`
	MaxAmount *int64 `form:"maxAmount" binding:"omitempty,gt=0"`

	// Classification filters
	UserCategoryID *string `form:"categoryId" binding:"omitempty,uuid"`

	// Text search (searches in description, userNote, counterparty name)
	Search *string `form:"search"`

	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`

	// Sorting
	SortBy    string `form:"sortBy" binding:"omitempty,oneof=booking_date value_date amount created_at"`
	SortOrder string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}
