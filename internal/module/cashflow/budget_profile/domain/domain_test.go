package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBudgetConstraint_IsActive(t *testing.T) {
	tests := []struct {
		name       string
		constraint *BudgetConstraint
		expected   bool
	}{
		{
			name:       "active constraint",
			constraint: &BudgetConstraint{IsActive: true},
			expected:   true,
		},
		{
			name:       "inactive constraint",
			constraint: &BudgetConstraint{IsActive: false},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constraint.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudgetConstraint_IsWithinRange(t *testing.T) {
	constraint := &BudgetConstraint{
		MinAmount: 1000000,
		MaxAmount: 5000000,
	}

	assert.True(t, constraint.IsWithinRange(3000000))
	assert.False(t, constraint.IsWithinRange(500000))
	assert.False(t, constraint.IsWithinRange(6000000))
}

func TestConstraintType_IsValid(t *testing.T) {
	assert.True(t, ConstraintTypeHard.IsValid())
	assert.True(t, ConstraintTypeSoft.IsValid())
	assert.False(t, ConstraintType("invalid").IsValid())
}

func TestBudgetConstraint_TableName(t *testing.T) {
	constraint := BudgetConstraint{}
	assert.Equal(t, "budget_constraints", constraint.TableName())
}
