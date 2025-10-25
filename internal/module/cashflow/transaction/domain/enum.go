package domain

/*
Transaction represents a financial transaction
Direction: DEBIT / CREDIT
  - DEBIT: money goes OUT from this account (expense, transfer out)
  - CREDIT: money comes IN to this account (income, refund, transfer in)
*/
type Direction string

const (
	DirectionDebit  Direction = "DEBIT"
	DirectionCredit Direction = "CREDIT"
)

// Channel: where the transaction was initiated (if known)
type Channel string

const (
	ChannelMobile  Channel = "MOBILE_APP"
	ChannelWeb     Channel = "INTERNET_BANKING"
	ChannelATM     Channel = "ATM"
	ChannelPOS     Channel = "POS"
	ChannelUnknown Channel = "UNKNOWN"
)

// TransactionSource: transaction origin source
type TransactionSource string

const (
	SourceBankAPI    TransactionSource = "BANK_API"    // pulled from bank API
	SourceCsvImport  TransactionSource = "CSV_IMPORT"  // imported from CSV/Excel
	SourceJsonImport TransactionSource = "JSON_IMPORT" // imported from JSON
	SourceManual     TransactionSource = "MANUAL"      // user manually entered (cash, adjustment...)
)

// Instrument: financial instrument type, used to indicate the type of account involved
// CASH, BANK_ACCOUNT, CARD, E_WALLET, CRYPTO, ...
type Instrument string

const (
	InstrumentUnknown     Instrument = "UNKNOWN"
	InstrumentCash        Instrument = "CASH"         // Cash
	InstrumentBankAccount Instrument = "BANK_ACCOUNT" // Bank account
	InstrumentDebitCard   Instrument = "DEBIT_CARD"
	InstrumentCreditCard  Instrument = "CREDIT_CARD"
	InstrumentEWallet     Instrument = "E_WALLET" // E-wallets like Momo, ZaloPay, etc.
	InstrumentCrypto      Instrument = "CRYPTO"
)

// =========================
// Unified links (GOAL / BUDGET / DEBT)
// =========================
type LinkType string

const (
	LinkGoal   LinkType = "GOAL"
	LinkBudget LinkType = "BUDGET"
	LinkDebt   LinkType = "DEBT"
)
