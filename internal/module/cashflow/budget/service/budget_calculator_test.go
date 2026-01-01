package service

import (
	"testing"
)

// All BudgetCalculator tests are skipped because they require direct DB access
// The calculator uses c.service.db.Table() which cannot be easily mocked.
// These should be tested using integration tests with a real database.

func TestBudgetCalculator_RecalculateBudgetSpending_Success(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RecalculateBudgetSpending_NotFound(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RecalculateAllBudgets_Success(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RecalculateAllBudgets_NoBudgets(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RolloverBudgets_Success(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RolloverBudgets_NoRollover(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}

func TestBudgetCalculator_RolloverBudgets_ExceededBudget(t *testing.T) {
	t.Skip("Skipping due to direct DB access requirement. Use integration tests instead.")
}
