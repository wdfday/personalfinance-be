package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBudget_TableName(t *testing.T) {
	budget := Budget{}
	assert.Equal(t, "budgets", budget.TableName())
}

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
			name: "warning threshold - 80%",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 8000000,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  2000000,
			expectedPercentage: 80.0,
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
			expectedStatus:     BudgetStatusWarning,
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
			expectedStatus:     BudgetStatusWarning, // IsExceeded() is false when exactly at budget
		},
		{
			name: "zero amount budget",
			budget: &Budget{
				Amount:      0,
				SpentAmount: 0,
				Status:      BudgetStatusActive,
			},
			expectedRemaining:  0,
			expectedPercentage: 0.0,
			expectedStatus:     BudgetStatusActive,
		},
		{
			name: "expired budget with under spending",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 5000000,
				Status:      BudgetStatusActive,
				EndDate:     timePtr(time.Now().AddDate(0, 0, -1)),
			},
			expectedRemaining:  5000000,
			expectedPercentage: 50.0,
			expectedStatus:     BudgetStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.budget.UpdateCalculatedFields()

			assert.Equal(t, tt.expectedRemaining, tt.budget.RemainingAmount, "RemainingAmount mismatch")
			assert.InDelta(t, tt.expectedPercentage, tt.budget.PercentageSpent, 0.01, "PercentageSpent mismatch")
			assert.Equal(t, tt.expectedStatus, tt.budget.Status, "Status mismatch")
			assert.NotNil(t, tt.budget.LastCalculatedAt, "LastCalculatedAt should be set")
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
		// Removed - time.Now().After(endDate) is non-deterministic for current time
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_IsExceeded(t *testing.T) {
	tests := []struct {
		name     string
		budget   *Budget
		expected bool
	}{
		{
			name: "not exceeded - under budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 5000000,
			},
			expected: false,
		},
		{
			name: "exceeded budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 11000000,
			},
			expected: true,
		},
		{
			name: "exactly at budget",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 10000000,
			},
			expected: false,
		},
		{
			name: "slightly exceeded",
			budget: &Budget{
				Amount:      10000000,
				SpentAmount: 10000001,
			},
			expected: true,
		},
		{
			name: "zero budget",
			budget: &Budget{
				Amount:      0,
				SpentAmount: 0,
			},
			expected: false,
		},
		{
			name: "negative remaining",
			budget: &Budget{
				Amount:      5000000,
				SpentAmount: 8000000,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.IsExceeded()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_ShouldAlert(t *testing.T) {
	tests := []struct {
		name      string
		budget    *Budget
		threshold AlertThreshold
		expected  bool
	}{
		{
			name: "should alert - alerts enabled and threshold in list",
			budget: &Budget{
				EnableAlerts:    true,
				AlertThresholds: AlertThresholdsJSON{AlertAt50, AlertAt75, AlertAt90},
			},
			threshold: AlertAt75,
			expected:  true,
		},
		{
			name: "should not alert - alerts disabled",
			budget: &Budget{
				EnableAlerts:    false,
				AlertThresholds: AlertThresholdsJSON{AlertAt75, AlertAt90, AlertAt100},
			},
			threshold: AlertAt75,
			expected:  false,
		},
		{
			name: "should not alert - threshold not in list",
			budget: &Budget{
				EnableAlerts:    true,
				AlertThresholds: AlertThresholdsJSON{AlertAt50, AlertAt90},
			},
			threshold: AlertAt75,
			expected:  false,
		},
		{
			name: "should alert - at 50% threshold",
			budget: &Budget{
				EnableAlerts:    true,
				AlertThresholds: AlertThresholdsJSON{AlertAt50, AlertAt75, AlertAt90, AlertAt100},
			},
			threshold: AlertAt50,
			expected:  true,
		},
		{
			name: "should alert - at 100% threshold",
			budget: &Budget{
				EnableAlerts:    true,
				AlertThresholds: AlertThresholdsJSON{AlertAt100},
			},
			threshold: AlertAt100,
			expected:  true,
		},
		{
			name: "should not alert - empty threshold list",
			budget: &Budget{
				EnableAlerts:    true,
				AlertThresholds: AlertThresholdsJSON{},
			},
			threshold: AlertAt75,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.budget.ShouldAlert(tt.threshold)
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
		{"valid - exceeded", BudgetStatusExceeded, true},
		{"valid - expired", BudgetStatusExpired, true},
		{"valid - paused", BudgetStatusPaused, true},
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

func TestAlertThreshold_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		threshold AlertThreshold
		expected  bool
	}{
		{"valid - 50%", AlertAt50, true},
		{"valid - 75%", AlertAt75, true},
		{"valid - 90%", AlertAt90, true},
		{"valid - 100%", AlertAt100, true},
		{"invalid - empty", AlertThreshold(""), false},
		{"invalid - unknown", AlertThreshold("unknown"), false},
		{"invalid - 60%", AlertThreshold("60"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.threshold.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBudget_Structure(t *testing.T) {
	t.Run("create budget with all fields", func(t *testing.T) {
		userID := uuid.New()
		categoryID := uuid.New()
		accountID := uuid.New()
		startDate := time.Now()
		endDate := startDate.AddDate(0, 1, 0)
		description := "Monthly budget for groceries"
		lastCalculated := time.Now()
		carryOver := 50
		autoAdjustPct := 10
		autoAdjustBase := "average_spending"

		budget := Budget{
			ID:                   uuid.New(),
			UserID:               userID,
			Name:                 "Monthly Groceries",
			Description:          &description,
			Amount:               5000000,
			Currency:             "VND",
			Period:               BudgetPeriodMonthly,
			StartDate:            startDate,
			EndDate:              &endDate,
			CategoryID:           &categoryID,
			AccountID:            &accountID,
			SpentAmount:          2000000,
			RemainingAmount:      3000000,
			PercentageSpent:      40.0,
			Status:               BudgetStatusActive,
			LastCalculatedAt:     &lastCalculated,
			EnableAlerts:         true,
			AlertThresholds:      AlertThresholdsJSON{AlertAt50, AlertAt75, AlertAt90},
			AllowRollover:        true,
			CarryOverPercent:     &carryOver,
			AutoAdjust:           true,
			AutoAdjustPercentage: &autoAdjustPct,
			AutoAdjustBasedOn:    &autoAdjustBase,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, budget.ID)
		assert.Equal(t, userID, budget.UserID)
		assert.Equal(t, "Monthly Groceries", budget.Name)
		assert.Equal(t, description, *budget.Description)
		assert.Equal(t, 5000000.0, budget.Amount)
		assert.Equal(t, "VND", budget.Currency)
		assert.Equal(t, BudgetPeriodMonthly, budget.Period)
		assert.Equal(t, categoryID, *budget.CategoryID)
		assert.Equal(t, accountID, *budget.AccountID)
		assert.Equal(t, 2000000.0, budget.SpentAmount)
		assert.Equal(t, 3000000.0, budget.RemainingAmount)
		assert.Equal(t, 40.0, budget.PercentageSpent)
		assert.Equal(t, BudgetStatusActive, budget.Status)
		assert.NotNil(t, budget.LastCalculatedAt)
		assert.True(t, budget.EnableAlerts)
		assert.Len(t, budget.AlertThresholds, 3)
		assert.True(t, budget.AllowRollover)
		assert.Equal(t, 50, *budget.CarryOverPercent)
		assert.True(t, budget.AutoAdjust)
	})

	t.Run("create minimal budget", func(t *testing.T) {
		userID := uuid.New()

		budget := Budget{
			UserID:    userID,
			Name:      "Simple Budget",
			Amount:    1000000,
			Currency:  "VND",
			Period:    BudgetPeriodMonthly,
			StartDate: time.Now(),
			Status:    BudgetStatusActive,
		}

		assert.Equal(t, userID, budget.UserID)
		assert.Equal(t, "Simple Budget", budget.Name)
		assert.Equal(t, 1000000.0, budget.Amount)
		assert.Nil(t, budget.Description)
		assert.Nil(t, budget.EndDate)
		assert.Nil(t, budget.CategoryID)
		assert.Nil(t, budget.AccountID)
	})
}

func TestBudget_NullableFields(t *testing.T) {
	t.Run("all nullable fields are nil", func(t *testing.T) {
		budget := Budget{}

		assert.Nil(t, budget.Description)
		assert.Nil(t, budget.EndDate)
		assert.Nil(t, budget.CategoryID)
		assert.Nil(t, budget.AccountID)
		assert.Nil(t, budget.LastCalculatedAt)
		assert.Nil(t, budget.AlertedAt)
		assert.Nil(t, budget.CarryOverPercent)
		assert.Nil(t, budget.AutoAdjustPercentage)
		assert.Nil(t, budget.AutoAdjustBasedOn)
	})

	t.Run("set nullable fields", func(t *testing.T) {
		description := "Test description"
		endDate := time.Now()
		categoryID := uuid.New()
		accountID := uuid.New()
		lastCalculated := time.Now()
		alertedAt := "{\"75\": \"2024-01-01\"}"
		carryOver := 50
		autoAdjustPct := 10
		autoAdjustBase := "inflation"

		budget := Budget{
			Description:          &description,
			EndDate:              &endDate,
			CategoryID:           &categoryID,
			AccountID:            &accountID,
			LastCalculatedAt:     &lastCalculated,
			AlertedAt:            &alertedAt,
			CarryOverPercent:     &carryOver,
			AutoAdjustPercentage: &autoAdjustPct,
			AutoAdjustBasedOn:    &autoAdjustBase,
		}

		require.NotNil(t, budget.Description)
		assert.Equal(t, description, *budget.Description)

		require.NotNil(t, budget.EndDate)
		assert.Equal(t, endDate.Unix(), budget.EndDate.Unix())

		require.NotNil(t, budget.CategoryID)
		assert.Equal(t, categoryID, *budget.CategoryID)

		require.NotNil(t, budget.AccountID)
		assert.Equal(t, accountID, *budget.AccountID)

		require.NotNil(t, budget.LastCalculatedAt)
		assert.Equal(t, lastCalculated.Unix(), budget.LastCalculatedAt.Unix())

		require.NotNil(t, budget.AlertedAt)
		assert.Equal(t, alertedAt, *budget.AlertedAt)

		require.NotNil(t, budget.CarryOverPercent)
		assert.Equal(t, carryOver, *budget.CarryOverPercent)

		require.NotNil(t, budget.AutoAdjustPercentage)
		assert.Equal(t, autoAdjustPct, *budget.AutoAdjustPercentage)

		require.NotNil(t, budget.AutoAdjustBasedOn)
		assert.Equal(t, autoAdjustBase, *budget.AutoAdjustBasedOn)
	})
}

func TestBudget_BooleanFlags(t *testing.T) {
	t.Run("EnableAlerts flag", func(t *testing.T) {
		budget := Budget{EnableAlerts: true}
		assert.True(t, budget.EnableAlerts)

		budget.EnableAlerts = false
		assert.False(t, budget.EnableAlerts)
	})

	t.Run("AllowRollover flag", func(t *testing.T) {
		budget := Budget{AllowRollover: true}
		assert.True(t, budget.AllowRollover)

		budget.AllowRollover = false
		assert.False(t, budget.AllowRollover)
	})

	t.Run("AutoAdjust flag", func(t *testing.T) {
		budget := Budget{AutoAdjust: true}
		assert.True(t, budget.AutoAdjust)

		budget.AutoAdjust = false
		assert.False(t, budget.AutoAdjust)
	})

	t.Run("NotificationSent flag", func(t *testing.T) {
		budget := Budget{NotificationSent: true}
		assert.True(t, budget.NotificationSent)

		budget.NotificationSent = false
		assert.False(t, budget.NotificationSent)
	})
}

func TestBudget_DifferentPeriods(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name   string
		period BudgetPeriod
	}{
		{"daily budget", BudgetPeriodDaily},
		{"weekly budget", BudgetPeriodWeekly},
		{"monthly budget", BudgetPeriodMonthly},
		{"quarterly budget", BudgetPeriodQuarterly},
		{"yearly budget", BudgetPeriodYearly},
		{"custom budget", BudgetPeriodCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := Budget{
				UserID:    userID,
				Name:      tt.name,
				Period:    tt.period,
				Amount:    1000000,
				StartDate: time.Now(),
			}

			assert.Equal(t, tt.period, budget.Period)
		})
	}
}

func TestBudget_MultipleUpdates(t *testing.T) {
	budget := &Budget{
		Amount:      10000000,
		SpentAmount: 0,
	}

	// First update - 30%
	budget.SpentAmount = 3000000
	budget.UpdateCalculatedFields()
	assert.Equal(t, 7000000.0, budget.RemainingAmount)
	assert.Equal(t, 30.0, budget.PercentageSpent)
	assert.Equal(t, BudgetStatusActive, budget.Status)

	// Second update - 85%
	budget.SpentAmount = 8500000
	budget.UpdateCalculatedFields()
	assert.Equal(t, 1500000.0, budget.RemainingAmount)
	assert.Equal(t, 85.0, budget.PercentageSpent)
	assert.Equal(t, BudgetStatusWarning, budget.Status)

	// Third update - exceeded
	budget.SpentAmount = 11000000
	budget.UpdateCalculatedFields()
	assert.Equal(t, -1000000.0, budget.RemainingAmount)
	assert.InDelta(t, 110.0, budget.PercentageSpent, 0.01)
	assert.Equal(t, BudgetStatusExceeded, budget.Status)
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}
