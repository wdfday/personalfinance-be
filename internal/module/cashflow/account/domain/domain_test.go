package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_UpdateBalance(t *testing.T) {
	account := &Account{
		Balance: 10000000,
	}

	account.UpdateBalance(5000000)

	assert.Equal(t, 15000000.0, account.Balance)
}

func TestAccount_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		account  *Account
		expected bool
	}{
		{
			name:     "active account",
			account:  &Account{Status: AccountStatusActive},
			expected: true,
		},
		{
			name:     "inactive account",
			account:  &Account{Status: AccountStatusInactive},
			expected: false,
		},
		{
			name:     "closed account",
			account:  &Account{Status: AccountStatusClosed},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccountType_IsValid(t *testing.T) {
	assert.True(t, AccountTypeBank.IsValid())
	assert.True(t, AccountTypeCash.IsValid())
	assert.True(t, AccountTypeCreditCard.IsValid())
	assert.False(t, AccountType("invalid").IsValid())
}

func TestAccountStatus_IsValid(t *testing.T) {
	assert.True(t, AccountStatusActive.IsValid())
	assert.True(t, AccountStatusInactive.IsValid())
	assert.False(t, AccountStatus("invalid").IsValid())
}

func TestAccount_TableName(t *testing.T) {
	account := Account{}
	assert.Equal(t, "accounts", account.TableName())
}
