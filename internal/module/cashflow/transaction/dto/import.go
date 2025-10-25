package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
)

// BankTransactionJSON represents a transaction from bank JSON export
type BankTransactionJSON struct {
	ID                     string                    `json:"id"`
	ArrangementID          string                    `json:"arrangementId"` // Account ID from bank
	Reference              string                    `json:"reference"`
	Description            string                    `json:"description"`
	TypeGroup              string                    `json:"typeGroup"`
	Type                   string                    `json:"type"`
	Category               string                    `json:"category"`
	BookingDate            string                    `json:"bookingDate"`          // "2025-12-01"
	ValueDate              string                    `json:"valueDate"`            // "2025-12-01"
	CreditDebitIndicator   string                    `json:"creditDebitIndicator"` // "DBIT" or "CRED"
	TransactionAmount      TransactionAmountCurrency `json:"transactionAmountCurrency"`
	CounterPartyName       string                    `json:"counterPartyName"`
	CounterPartyAccountNo  string                    `json:"counterPartyAccountNumber"`
	CounterPartyBankName   string                    `json:"counterPartyBankName"`
	RunningBalance         int64                     `json:"runningBalance"`
	Additions              BankTransactionAdditions  `json:"additions"`
	CheckImageAvailability string                    `json:"checkImageAvailability"`
	CreationTime           string                    `json:"creationTime"` // "2025-12-01T16:41:38+07:00"
	State                  string                    `json:"state"`        // "COMPLETED"
}

// TransactionAmountCurrency represents the amount and currency
type TransactionAmountCurrency struct {
	Amount       string `json:"amount"`       // "116286"
	CurrencyCode string `json:"currencyCode"` // "VND"
}

// BankTransactionAdditions contains additional bank-specific data
type BankTransactionAdditions struct {
	CreditBank        string `json:"creditBank"`
	AccountNoOthCate  string `json:"accountNoOthCate"`
	AccountDebitName  string `json:"accountDebitName"`
	AccountNoOthName  string `json:"accountNoOthName"`
	ExternalID        string `json:"externalId"`
	DebitBank         string `json:"debitBank"`
	AccountNoOth      string `json:"accountNoOth"`
	CreditAcctName    string `json:"creditAcctName"`
	DebitAcctName     string `json:"debitAcctName"`
	CreditAcctNo      string `json:"creditAcctNo"`
	AdditionalInfo    string `json:"additionalInfo"`
	DebitAcctNo       string `json:"debitAcctNo"`
	CreditCate        string `json:"creditCate"`
	DebitCate         string `json:"debitCate"`
	AccountCreditName string `json:"accountCreditName"`
}

// ImportJSONRequest represents the request to import bank transactions
type ImportJSONRequest struct {
	AccountID    string                `json:"accountId" binding:"required,uuid"` // Your system's account ID
	BankCode     string                `json:"bankCode" binding:"required"`       // "TCB", "VCB", etc.
	Transactions []BankTransactionJSON `json:"transactions" binding:"required,min=1"`
}

// ImportJSONResponse represents the response after import
type ImportJSONResponse struct {
	TotalReceived  int                 `json:"totalReceived"`
	SuccessCount   int                 `json:"successCount"`
	SkippedCount   int                 `json:"skippedCount"` // Already exists
	FailedCount    int                 `json:"failedCount"`
	ImportedIDs    []string            `json:"importedIds"`
	SkippedIDs     []string            `json:"skippedIds"`
	Errors         []ImportError       `json:"errors,omitempty"`
	AccountBalance *AccountBalanceSync `json:"accountBalance,omitempty"`
}

// ImportError represents an error during import
type ImportError struct {
	BankTransactionID string `json:"bankTransactionId"`
	Error             string `json:"error"`
}

// AccountBalanceSync represents the synchronized account balance
type AccountBalanceSync struct {
	AccountID       string    `json:"accountId"`
	PreviousBalance int64     `json:"previousBalance"`
	NewBalance      int64     `json:"newBalance"`
	LastSyncedAt    time.Time `json:"lastSyncedAt"`
}

// ToBankTransactionDomain converts BankTransactionJSON to domain.Transaction
func (b *BankTransactionJSON) ToBankTransactionDomain(userID, accountID, bankCode string) (*domain.Transaction, error) {
	// Parse booking date
	bookingDate, err := time.Parse("2006-01-02", b.BookingDate)
	if err != nil {
		return nil, err
	}

	// Parse value date
	valueDate, err := time.Parse("2006-01-02", b.ValueDate)
	if err != nil {
		valueDate = bookingDate // Default to booking date if parse fails
	}

	// Parse amount
	var amount int64
	if _, err := fmt.Sscanf(b.TransactionAmount.Amount, "%d", &amount); err != nil {
		return nil, fmt.Errorf("invalid amount format: %s", b.TransactionAmount.Amount)
	}

	// Determine direction
	var direction domain.Direction
	if b.CreditDebitIndicator == "CRED" {
		direction = domain.DirectionCredit
	} else {
		direction = domain.DirectionDebit
	}

	// Build transaction
	transaction := &domain.Transaction{
		Direction:      direction,
		Instrument:     domain.InstrumentBankAccount,
		Source:         domain.SourceBankAPI,
		BankCode:       bankCode,
		ExternalID:     b.Additions.ExternalID,
		Channel:        domain.ChannelUnknown,
		Amount:         amount,
		Currency:       b.TransactionAmount.CurrencyCode,
		BookingDate:    bookingDate,
		ValueDate:      valueDate,
		Description:    b.Description,
		Reference:      b.Reference,
		RunningBalance: &b.RunningBalance,
	}

	// Set counterparty if available
	if b.CounterPartyName != "" {
		transaction.Counterparty = &domain.Counterparty{
			Name:          b.CounterPartyName,
			AccountNumber: b.CounterPartyAccountNo,
			BankName:      b.CounterPartyBankName,
			Type:          "UNKNOWN",
		}
	}

	// Build metadata with raw bank data
	rawData, _ := json.Marshal(b)
	transaction.Meta = &domain.TransactionMeta{
		CheckImageAvailability: b.CheckImageAvailability,
		Raw:                    rawData,
	}

	// Set timestamps
	now := time.Now()
	transaction.CreatedAt = now
	transaction.ImportedAt = &now

	return transaction, nil
}
