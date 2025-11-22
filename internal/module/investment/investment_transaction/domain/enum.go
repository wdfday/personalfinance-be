package domain

// TransactionType represents the type of investment transaction
type TransactionType string

const (
	TransactionTypeBuy      TransactionType = "buy"      // Purchase of assets
	TransactionTypeSell     TransactionType = "sell"     // Sale of assets
	TransactionTypeDividend TransactionType = "dividend" // Dividend payment
	TransactionTypeSplit    TransactionType = "split"    // Stock split
	TransactionTypeMerger   TransactionType = "merger"   // Merger/acquisition
	TransactionTypeTransfer TransactionType = "transfer" // Transfer between accounts
	TransactionTypeFee      TransactionType = "fee"      // Fee deduction
	TransactionTypeOther    TransactionType = "other"    // Other transaction types
)

// IsValid checks if the transaction type is valid
func (tt TransactionType) IsValid() bool {
	switch tt {
	case TransactionTypeBuy, TransactionTypeSell, TransactionTypeDividend,
		TransactionTypeSplit, TransactionTypeMerger, TransactionTypeTransfer,
		TransactionTypeFee, TransactionTypeOther:
		return true
	}
	return false
}

// String returns the string representation
func (tt TransactionType) String() string {
	return string(tt)
}

// TransactionStatus represents the status of an investment transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"   // Transaction pending
	TransactionStatusCompleted TransactionStatus = "completed" // Transaction completed
	TransactionStatusCancelled TransactionStatus = "cancelled" // Transaction cancelled
	TransactionStatusFailed    TransactionStatus = "failed"    // Transaction failed
)

// IsValid checks if the transaction status is valid
func (ts TransactionStatus) IsValid() bool {
	switch ts {
	case TransactionStatusPending, TransactionStatusCompleted,
		TransactionStatusCancelled, TransactionStatusFailed:
		return true
	}
	return false
}

// String returns the string representation
func (ts TransactionStatus) String() string {
	return string(ts)
}
