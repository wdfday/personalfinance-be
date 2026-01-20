package service

import (
	"testing"

	"personalfinancedss/internal/module/cashflow/transaction/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransactionLink_Types(t *testing.T) {
	tests := []struct {
		name     string
		linkType domain.LinkType
		expected domain.LinkType
	}{
		{"Goal link", domain.LinkGoal, "GOAL"},
		{"Budget link", domain.LinkBudget, "BUDGET"},
		{"Debt link", domain.LinkDebt, "DEBT"},
		{"Income profile link", domain.LinkIncomeProfile, "INCOME_PROFILE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.linkType)
		})
	}
}

func TestTransactionLink_Creation(t *testing.T) {
	goalID := uuid.New()

	link := domain.TransactionLink{
		Type: domain.LinkGoal,
		ID:   goalID.String(),
	}

	assert.Equal(t, domain.LinkGoal, link.Type)
	assert.Equal(t, goalID.String(), link.ID)

	// Verify ID is a valid UUID
	parsed, err := uuid.Parse(link.ID)
	assert.NoError(t, err)
	assert.Equal(t, goalID, parsed)
}

func TestTransactionLink_MultipleLinks(t *testing.T) {
	goalID := uuid.New()
	budgetID := uuid.New()

	links := []domain.TransactionLink{
		{Type: domain.LinkGoal, ID: goalID.String()},
		{Type: domain.LinkBudget, ID: budgetID.String()},
	}

	assert.Len(t, links, 2)
	assert.Equal(t, domain.LinkGoal, links[0].Type)
	assert.Equal(t, domain.LinkBudget, links[1].Type)
}

func TestDirection_Types(t *testing.T) {
	assert.Equal(t, domain.Direction("DEBIT"), domain.DirectionDebit)
	assert.Equal(t, domain.Direction("CREDIT"), domain.DirectionCredit)
}
