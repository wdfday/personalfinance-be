package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebt_UpdateCalculatedFields(t *testing.T) {
	debt := &Debt{
		PrincipalAmount: 10000000,
		CurrentBalance:  6000000,
		TotalPaid:       4000000,
	}

	debt.UpdateCalculatedFields()

	assert.Equal(t, 4000000.0, debt.RemainingAmount)
	assert.Equal(t, 40.0, debt.PercentagePaid)
}

func TestDebt_AddPayment(t *testing.T) {
	debt := &Debt{
		PrincipalAmount: 10000000,
		CurrentBalance:  6000000,
		TotalPaid:       4000000,
		InterestRate:    0.1,
	}

	debt.AddPayment(1000000)

	assert.Equal(t, 5000000.0, debt.CurrentBalance)
	assert.Equal(t, 5000000.0, debt.TotalPaid)
}

func TestDebt_IsPaidOff(t *testing.T) {
	tests := []struct {
		name     string
		debt     *Debt
		expected bool
	}{
		{
			name: "paid off - zero balance",
			debt: &Debt{
				CurrentBalance: 0,
				Status:         DebtStatusActive,
			},
			expected: true,
		},
		{
			name: "paid off - status",
			debt: &Debt{
				CurrentBalance: 100,
				Status:         DebtStatusPaidOff,
			},
			expected: true,
		},
		{
			name: "not paid off",
			debt: &Debt{
				CurrentBalance: 5000000,
				Status:         DebtStatusActive,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.debt.IsPaidOff()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDebtType_IsValid(t *testing.T) {
	assert.True(t, DebtTypeCreditCard.IsValid())
	assert.True(t, DebtTypeLoan.IsValid())
	assert.False(t, DebtType("invalid").IsValid())
}

func TestDebtStatus_IsValid(t *testing.T) {
	assert.True(t, DebtStatusActive.IsValid())
	assert.True(t, DebtStatusPaidOff.IsValid())
	assert.False(t, DebtStatus("invalid").IsValid())
}
