package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Transaction is a financial transaction record in the system.
// It can represent various types of transactions, including:
// - Bank transaction (bank account)
// - E-wallet transaction
// - Cash transaction (cash in/out)
// - Any other "account" type modeled in your system
type Transaction struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`             // internal ID (UUID / snowflake / ... )
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"userId"`       // FK to user
	AccountID uuid.UUID `gorm:"type:uuid;not null;index;column:account_id" json:"accountId"` // FK to account (bank account / e-wallet / cash account ...)

	// Bank / external system related information (optional)
	BankCode   string            `gorm:"type:varchar(20);column:bank_code" json:"bankCode,omitempty"`      // "TCB", "VCB"..., or blank if not a bank
	Source     TransactionSource `gorm:"type:varchar(20);not null;column:source" json:"source"`            // BANK_API / CSV_IMPORT / MANUAL
	ExternalID string            `gorm:"type:varchar(255);column:external_id" json:"externalId,omitempty"` // External transaction ID from bank/wallet

	Direction Direction `gorm:"type:varchar(20);not null;column:direction" json:"direction"` // DEBIT / CREDIT
	Channel   Channel   `gorm:"type:varchar(50);column:channel" json:"channel,omitempty"`    // MOBILE_APP / INTERNET_BANKING / ATM / POS / UNKNOWN

	// Instrument: payment method / account type
	// - CASH for cash transactions
	// - BANK_ACCOUNT for bank account transactions
	// - E_WALLET, CREDIT_CARD, etc.
	Instrument Instrument `gorm:"type:varchar(50);not null;column:instrument" json:"instrument"`

	// Timestamps
	BookingDate time.Time  `gorm:"type:timestamp;not null;index;column:booking_date" json:"bookingDate"` // Transaction booking/posting date
	ValueDate   time.Time  `gorm:"type:timestamp;not null;column:value_date" json:"valueDate"`           // Effective date (for cash can equal BookingDate)
	CreatedAt   time.Time  `gorm:"autoCreateTime;column:created_at" json:"createdAt"`                    // Created timestamp in your system
	ImportedAt  *time.Time `gorm:"type:timestamp;column:imported_at" json:"importedAt,omitempty"`        // Import timestamp (if different from CreatedAt)

	// Amount
	//
	// Recommendation: use smallest unit (VND = dong), don't use float to avoid rounding errors.
	// Amount is always the "absolute amount" for this transaction; direction (in/out) is based on Direction.
	Amount         int64  `gorm:"type:bigint;not null;column:amount" json:"amount"`                        // e.g. 116286 means 116,286 VND
	Currency       string `gorm:"type:varchar(10);not null;default:'VND';column:currency" json:"currency"` // "VND"
	RunningBalance *int64 `gorm:"type:bigint;column:running_balance" json:"runningBalance,omitempty"`      // Balance after transaction (if available – usually only with bank/wallet)

	// Description information
	//
	// Description: "technical" description from bank / import file
	// UserNote: user-entered note (especially useful for cash)
	Description string `gorm:"type:text;column:description" json:"description,omitempty"`     // Description from bank / file / system logic
	UserNote    string `gorm:"type:text;column:user_note" json:"userNote,omitempty"`          // User note for this transaction
	Reference   string `gorm:"type:varchar(255);column:reference" json:"reference,omitempty"` // Reference code from bank / wallet / external system

	// User-selected category (FK to categories table) - stored as separate column for fast querying
	UserCategoryID *uuid.UUID `gorm:"type:uuid;column:user_category_id;index" json:"userCategoryId,omitempty"`

	// Transaction counterparty (merchant / recipient / sender)
	// For cash, you can still use:
	// - Name: "Street food vendor", "Money to mom", etc.
	Counterparty *Counterparty `gorm:"type:jsonb;column:counterparty" json:"counterparty,omitempty"`

	// Links to other domain entities (budget, goal, debt, ...)
	Links *TransactionLinks `gorm:"type:jsonb;column:links" json:"links,omitempty"`

	// Metadata & raw data from bank / wallet / external systems
	Meta *TransactionMeta `gorm:"type:jsonb;column:meta" json:"meta,omitempty"`
}

// TableName specifies the database table name
func (Transaction) TableName() string {
	return "transactions"
}

// Sub-models

// Counterparty: information about the other party in the transaction.
type Counterparty struct {
	Name          string `json:"name,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	BankName      string `json:"bankName,omitempty"` // e.g.: "Techcombank"

	// partner type: MERCHANT / PERSON / INTERNAL / UNKNOWN
	Type string `json:"type,omitempty"`
}

// Value implements driver.Valuer for JSONB
func (c *Counterparty) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for JSONB
func (c *Counterparty) Scan(value interface{}) error {
	if value == nil {
		*c = Counterparty{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// TransactionLink: links a transaction to another financial entity (goal, budget, debt...)
type TransactionLink struct {
	Type LinkType `json:"type"` // GOAL / BUDGET / DEBT
	ID   string   `json:"id"`   // FK to corresponding entity (goal_id, budget_id, debt_id, ...)
}

// TransactionLinks is a slice of TransactionLink for GORM JSON handling
type TransactionLinks []TransactionLink

// Value implements driver.Valuer for JSONB
func (tl TransactionLinks) Value() (driver.Value, error) {
	if tl == nil {
		return nil, nil
	}
	return json.Marshal(tl)
}

// Scan implements sql.Scanner for JSONB
func (tl *TransactionLinks) Scan(value interface{}) error {
	if value == nil {
		*tl = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, tl)
}

// TransactionMeta: Additional metadata & raw data from banks/wallets/external systems
type TransactionMeta struct {
	CheckImageAvailability string          `json:"checkImageAvailability,omitempty"` // e.g.: "UNAVAILABLE" for bank
	Raw                    json.RawMessage `json:"raw,omitempty"`                    // Original raw JSON from bank / wallet (if available)
}

// Value implements driver.Valuer for JSONB
func (tm *TransactionMeta) Value() (driver.Value, error) {
	if tm == nil {
		return nil, nil
	}
	return json.Marshal(tm)
}

// Scan implements sql.Scanner for JSONB
func (tm *TransactionMeta) Scan(value interface{}) error {
	if value == nil {
		*tm = TransactionMeta{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, tm)
}
