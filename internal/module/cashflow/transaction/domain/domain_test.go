package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_IsIncome(t *testing.T) {
	tests := []struct {
		name        string
		transaction *Transaction
		expected    bool
	}{
		{
			name:        "income transaction",
			transaction: &Transaction{Type: TransactionTypeIncome},
			expected:    true,
		},
		{
			name:        "expense transaction",
			transaction: &Transaction{Type: TransactionTypeExpense},
			expected:    false,
		},
		{
			name:        "transfer transaction",
			transaction: &Transaction{Type: TransactionTypeTransfer},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.transaction.IsIncome()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransaction_IsExpense(t *testing.T) {
	tests := []struct {
		name        string
		transaction *Transaction
		expected    bool
	}{
		{
			name:        "expense transaction",
			transaction: &Transaction{Type: TransactionTypeExpense},
			expected:    true,
		},
		{
			name:        "income transaction",
			transaction: &Transaction{Type: TransactionTypeIncome},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.transaction.IsExpense()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransactionType_IsValid(t *testing.T) {
	assert.True(t, TransactionTypeIncome.IsValid())
	assert.True(t, TransactionTypeExpense.IsValid())
	assert.True(t, TransactionTypeTransfer.IsValid())
	assert.False(t, TransactionType("invalid").IsValid())
}

func TestTransactionStatus_IsValid(t *testing.T) {
	assert.True(t, TransactionStatusCompleted.IsValid())
	assert.True(t, TransactionStatusPending.IsValid())
	assert.False(t, TransactionStatus("invalid").IsValid())
}

func TestTransaction_TableName(t *testing.T) {
	transaction := Transaction{}
	assert.Equal(t, "transactions", transaction.TableName())
}
