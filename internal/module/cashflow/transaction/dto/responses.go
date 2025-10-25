package dto

import "time"

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	// Core identifiers
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	AccountID string `json:"accountId"`

	// Transaction type
	Direction  string `json:"direction"`  // DEBIT / CREDIT
	Instrument string `json:"instrument"` // CASH / BANK_ACCOUNT / etc.
	Source     string `json:"source"`     // BANK_API / CSV_IMPORT / MANUAL / JSON_IMPORT

	// Bank / external system information
	BankCode   string `json:"bankCode,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
	Channel    string `json:"channel,omitempty"` // MOBILE_APP / INTERNET_BANKING / etc.

	// Amount (in smallest currency unit)
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	RunningBalance *int64 `json:"runningBalance,omitempty"` // Balance after this transaction

	// Timestamps
	BookingDate time.Time  `json:"bookingDate"`          // Transaction booking/posting date
	ValueDate   time.Time  `json:"valueDate"`            // Effective date
	CreatedAt   time.Time  `json:"createdAt"`            // Created in system
	ImportedAt  *time.Time `json:"importedAt,omitempty"` // Import timestamp (if imported)

	// Description fields
	Description string `json:"description,omitempty"` // Technical description
	UserNote    string `json:"userNote,omitempty"`    // User's note
	Reference   string `json:"reference,omitempty"`   // Bank reference code

	// Counterparty information
	Counterparty *CounterpartyResponse `json:"counterparty,omitempty"`

	// Classification
	Classification *ClassificationResponse `json:"classification,omitempty"`

	// Links to other entities
	Links []TransactionLinkResponse `json:"links,omitempty"`

	// Metadata
	Meta *TransactionMetaResponse `json:"meta,omitempty"`
}

// CounterpartyResponse represents counterparty information in API responses
type CounterpartyResponse struct {
	Name          string `json:"name,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	BankName      string `json:"bankName,omitempty"`
	Type          string `json:"type,omitempty"` // MERCHANT / PERSON / INTERNAL / UNKNOWN
}

// ClassificationResponse represents transaction classification in API responses
type ClassificationResponse struct {
	SystemCategory string   `json:"systemCategory,omitempty"` // e.g., "SPENDING:GROCERIES"
	UserCategoryID string   `json:"userCategoryId,omitempty"` // User-selected category
	IsTransfer     bool     `json:"isTransfer,omitempty"`     // Transfer between user's accounts
	IsRefund       bool     `json:"isRefund,omitempty"`       // Refund transaction
	Tags           []string `json:"tags,omitempty"`           // Free-form tags
}

// TransactionLinkResponse represents a link to another financial entity
type TransactionLinkResponse struct {
	Type string `json:"type"` // GOAL / BUDGET / DEBT
	ID   string `json:"id"`   // Entity ID
}

// TransactionMetaResponse represents additional metadata
type TransactionMetaResponse struct {
	CheckImageAvailability string                 `json:"checkImageAvailability,omitempty"`
	Raw                    map[string]interface{} `json:"raw,omitempty"` // Raw data from bank/wallet
}

// TransactionListResponse represents a paginated list of transactions
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   PaginationInfo        `json:"pagination"`
	Summary      *TransactionSummary   `json:"summary,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
	TotalCount int64 `json:"totalCount"`
}

// TransactionSummary provides aggregate information about transactions
type TransactionSummary struct {
	// Total amounts by direction
	TotalDebit  int64 `json:"totalDebit"`  // Total outgoing (expenses, transfers out)
	TotalCredit int64 `json:"totalCredit"` // Total incoming (income, refunds, transfers in)
	NetAmount   int64 `json:"netAmount"`   // Credit - Debit

	// Breakdown by instrument
	ByInstrument map[string]InstrumentSummary `json:"byInstrument,omitempty"`

	// Breakdown by source
	BySource map[string]SourceSummary `json:"bySource,omitempty"`

	// Transaction count
	Count int64 `json:"count"`
}

// InstrumentSummary represents summary for a specific instrument type
type InstrumentSummary struct {
	Debit  int64 `json:"debit"`
	Credit int64 `json:"credit"`
	Count  int64 `json:"count"`
}

// SourceSummary represents summary for a specific source
type SourceSummary struct {
	Debit  int64 `json:"debit"`
	Credit int64 `json:"credit"`
	Count  int64 `json:"count"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
