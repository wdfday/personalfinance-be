package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBudget_UpdateCalculatedFields(t *testing.T) {
	tests := []struct {
		name               string
		budget             *Budget
		expectedRemaining  float64
		expectedPercentage float64
		expectedStatus     BudgetStatus
	}{
		{
			name: "normal spending - under budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 5000000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  5000000,
			expectedPercentage: 50.0,
			expectedStatus:     BudgetStatusActive,
		},
		{
			name: "warning threshold - 75%",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 7500000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  2500000,
			expectedPercentage: 75.0,
			expectedStatus:     BudgetStatusWarning,
		},
		{
			name: "critical threshold - 90%",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 9000000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  1000000,
			expectedPercentage: 90.0,
			expectedStatus:     BudgetStatusCritical,
		},
		{
			name: "exceeded budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 11000000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  -1000000,
			expectedPercentage: 110.0,
			expectedStatus:     BudgetStatusExceeded,
		},
		{
			name: "zero spending",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 0,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  10000000,
			expectedPercentage: 0.0,
			expectedStatus:     BudgetStatusActive,
		},
		{
			name: "exactly at budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 10000000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  0,
			expectedPercentage: 100.0,
			expectedStatus:     BudgetStatusExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.budget.UpdateCalculatedFields()

			assert.Equal(t, tt.expectedRemaining, tt.budget.RemainingAmount, "RemainingAmount mismatch")
			assert.Equal(t, tt.expectedPercentage, tt.budget.PercentageSpent, "PercentageSpent mismatch")
			assert.Equal(t, tt.expectedStatus, tt.budget.Status, "Status mismatch")
		})
	}
}

func TestBudget_IsExpired(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name     string
		budget   *Budget
		expected bool
	}{
		{
			name: "expired budget",
			budget: &Budget{
				EndDate: &yesterday,
				Status:  BudgetStatusActive,
			},
			expected: true,
		},
		{
			name: "not expired - future end date",
			budget: &Budget{
				EndDate: &tomorrow,
				Status:  BudgetStatusActive,
			},
			expected: false,
		},
		{
			name: "no end date - recurring budget",
			budget: &Budget{
				EndDate: nil,
				Status:  BudgetStatusActive,
			},
			expected: false,
		},
		{
			name: "already expired status",
			budget: &Budget{
				EndDate: &yesterday,
				Status:  BudgetStatusExpired,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		budget   *Budget
		expected bool
	}{
		{
			name: "active budget",
			budget: &Budget{
				Status: BudgetStatusActive,
			},
			expected: true,
		},
		{
			name: "warning budget - still active",
			budget: &Budget{
				Status: BudgetStatusWarning,
			},
			expected: true,
		},
		{
			name: "critical budget - still active",
			budget: &Budget{
				Status: BudgetStatusCritical,
			},
			expected: true,
		},
		{
			name: "exceeded budget - not active",
			budget: &Budget{
				Status: BudgetStatusExceeded,
			},
			expected: false,
		},
		{
			name: "expired budget - not active",
			budget: &Budget{
				Status: BudgetStatusExpired,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_ShouldAlert(t *testing.T) {
	tests := []struct {
		name     string
		budget   *Budget
		expected bool
	}{
		{
			name: "should alert - alerts enabled and at threshold",
			budget: &Budget{
				EnableAlerts:    true,
				Amount:          10000000,
				SpentAmount:     7500000,
				PercentageSpent: 75.0,
				AlertThresholds: []AlertThreshold{AlertAt75, AlertAt90, AlertAt100},
			},
			expected: true,
		},
		{
			name: "should not alert - alerts disabled",
			budget: &Budget{
				EnableAlerts:    false,
				Amount:          10000000,
				SpentAmount:     7500000,
				PercentageSpent: 75.0,
				AlertThresholds: []AlertThreshold{AlertAt75, AlertAt90, AlertAt100},
			},
			expected: false,
		},
		{
			name: "should not alert - below threshold",
			budget: &Budget{
				EnableAlerts:    true,
				Amount:          10000000,
				SpentAmount:     5000000,
				PercentageSpent: 50.0,
				AlertThresholds: []AlertThreshold{AlertAt75, AlertAt90, AlertAt100},
			},
			expected: false,
		},
		{
			name: "should alert - at 90% threshold",
			budget: &Budget{
				EnableAlerts:    true,
				Amount:          10000000,
				SpentAmount:     9000000,
				PercentageSpent: 90.0,
				AlertThresholds: []AlertThreshold{AlertAt75, AlertAt90, AlertAt100},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.ShouldAlert()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_CalculateRolloverAmount(t *testing.T) {
	tests := []struct {
		name     string
		budget   *Budget
		expected float64
	}{
		{
			name: "rollover with carry over percent",
			budget: &Budget{
				Amount:           10000000,
				SpentAmount:      7000000,
				RemainingAmount:  3000000,
				AllowRollover:    true,
				CarryOverPercent: intPtr(50),
			},
			expected: 1500000, // 50% of 3000000
		},
		{
			name: "rollover full remaining amount",
			budget: &Budget{
				Amount:           10000000,
				SpentAmount:      7000000,
				RemainingAmount:  3000000,
				AllowRollover:    true,
				CarryOverPercent: nil,
			},
			expected: 3000000, // 100% of remaining
		},
		{
			name: "no rollover - disabled",
			budget: &Budget{
				Amount:           10000000,
				SpentAmount:      7000000,
				RemainingAmount:  3000000,
				AllowRollover:    false,
				CarryOverPercent: intPtr(50),
			},
			expected: 0,
		},
		{
			name: "no rollover - exceeded budget",
			budget: &Budget{
				Amount:           10000000,
				SpentAmount:      11000000,
				RemainingAmount:  -1000000,
				AllowRollover:    true,
				CarryOverPercent: intPtr(50),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.CalculateRolloverAmount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudgetPeriod_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		period   BudgetPeriod
		expected bool
	}{
		{"valid - daily", BudgetPeriodDaily, true},
		{"valid - weekly", BudgetPeriodWeekly, true},
		{"valid - monthly", BudgetPeriodMonthly, true},
		{"valid - quarterly", BudgetPeriodQuarterly, true},
		{"valid - yearly", BudgetPeriodYearly, true},
		{"valid - custom", BudgetPeriodCustom, true},
		{"invalid - empty", BudgetPeriod(""), false},
		{"invalid - unknown", BudgetPeriod("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.period.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudgetStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   BudgetStatus
		expected bool
	}{
		{"valid - active", BudgetStatusActive, true},
		{"valid - warning", BudgetStatusWarning, true},
		{"valid - critical", BudgetStatusCritical, true},
		{"valid - exceeded", BudgetStatusExceeded, true},
		{"valid - expired", BudgetStatusExpired, true},
		{"invalid - empty", BudgetStatus(""), false},
		{"invalid - unknown", BudgetStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertThreshold_Value(t *testing.T) {
	tests := []struct {
		name      string
		threshold AlertThreshold
		expected  float64
	}{
		{"50%", AlertAt50, 50.0},
		{"75%", AlertAt75, 75.0},
		{"90%", AlertAt90, 90.0},
		{"100%", AlertAt100, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.threshold.Value()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_TableName(t *testing.T) {
	budget := Budget{}
	assert.Equal(t, "budgets", budget.TableName())
}

func TestBudget_NewBudget(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	budget := &Budget{
		UserID:     userID,
		Name:       "Monthly Groceries",
		Amount:     5000000,
		Currency:   "VND",
		Period:     BudgetPeriodMonthly,
		StartDate:  startDate,
		CategoryID: &categoryID,
		Status:     BudgetStatusActive,
	}

	assert.NotNil(t, budget)
	assert.Equal(t, userID, budget.UserID)
	assert.Equal(t, "Monthly Groceries", budget.Name)
	assert.Equal(t, 5000000.0, budget.Amount)
	assert.Equal(t, BudgetPeriodMonthly, budget.Period)
	assert.Equal(t, BudgetStatusActive, budget.Status)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
