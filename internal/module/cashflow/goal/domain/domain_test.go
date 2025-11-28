package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGoal_UpdateCalculatedFields(t *testing.T) {
	tests := []struct {
		name               string
		goal               *Goal
		expectedRemaining  float64
		expectedPercentage float64
		expectedStatus     GoalStatus
	}{
		{
			name: "normal progress - 50%",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 5000000,
				Status:        GoalStatusActive,
			},
			expectedRemaining:  5000000,
			expectedPercentage: 50.0,
			expectedStatus:     GoalStatusActive,
		},
		{
			name: "goal completed",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 10000000,
				Status:        GoalStatusActive,
			},
			expectedRemaining:  0,
			expectedPercentage: 100.0,
			expectedStatus:     GoalStatusCompleted,
		},
		{
			name: "goal exceeded",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 11000000,
				Status:        GoalStatusActive,
			},
			expectedRemaining:  0,
			expectedPercentage: 100.0,
			expectedStatus:     GoalStatusCompleted,
		},
		{
			name: "zero progress",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 0,
				Status:        GoalStatusActive,
			},
			expectedRemaining:  10000000,
			expectedPercentage: 0.0,
			expectedStatus:     GoalStatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.goal.UpdateCalculatedFields()

			assert.Equal(t, tt.expectedRemaining, tt.goal.RemainingAmount, "RemainingAmount mismatch")
			assert.Equal(t, tt.expectedPercentage, tt.goal.PercentageComplete, "PercentageComplete mismatch")
			assert.Equal(t, tt.expectedStatus, tt.goal.Status, "Status mismatch")
		})
	}
}

func TestGoal_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		goal     *Goal
		expected bool
	}{
		{
			name: "completed - reached target",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 10000000,
				Status:        GoalStatusActive,
			},
			expected: true,
		},
		{
			name: "completed - exceeded target",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 11000000,
				Status:        GoalStatusActive,
			},
			expected: true,
		},
		{
			name: "completed - status completed",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 8000000,
				Status:        GoalStatusCompleted,
			},
			expected: true,
		},
		{
			name: "not completed",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 5000000,
				Status:        GoalStatusActive,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goal.IsCompleted()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoal_IsOverdue(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name     string
		goal     *Goal
		expected bool
	}{
		{
			name: "overdue - past target date",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 5000000,
				TargetDate:    &yesterday,
				Status:        GoalStatusActive,
			},
			expected: true,
		},
		{
			name: "not overdue - future target date",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 5000000,
				TargetDate:    &tomorrow,
				Status:        GoalStatusActive,
			},
			expected: false,
		},
		{
			name: "not overdue - no target date",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 5000000,
				TargetDate:    nil,
				Status:        GoalStatusActive,
			},
			expected: false,
		},
		{
			name: "not overdue - completed",
			goal: &Goal{
				TargetAmount:  10000000,
				CurrentAmount: 10000000,
				TargetDate:    &yesterday,
				Status:        GoalStatusCompleted,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goal.IsOverdue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoal_DaysRemaining(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 10)

	tests := []struct {
		name     string
		goal     *Goal
		expected int
	}{
		{
			name: "10 days remaining",
			goal: &Goal{
				TargetDate: &future,
			},
			expected: 10,
		},
		{
			name: "no target date",
			goal: &Goal{
				TargetDate: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goal.DaysRemaining()
			// Allow 1 day tolerance for test execution time
			assert.InDelta(t, tt.expected, result, 1)
		})
	}
}

func TestGoal_CalculateSuggestedContribution(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 30) // 30 days from now

	tests := []struct {
		name      string
		goal      *Goal
		frequency ContributionFrequency
		expected  float64
	}{
		{
			name: "monthly contribution",
			goal: &Goal{
				TargetAmount:    10000000,
				CurrentAmount:   4000000,
				RemainingAmount: 6000000,
				TargetDate:      &future,
			},
			frequency: FrequencyMonthly,
			expected:  6000000, // All remaining in one month
		},
		{
			name: "weekly contribution",
			goal: &Goal{
				TargetAmount:    10000000,
				CurrentAmount:   4000000,
				RemainingAmount: 6000000,
				TargetDate:      &future,
			},
			frequency: FrequencyWeekly,
			expected:  1500000, // 6000000 / 4 weeks
		},
		{
			name: "no remaining amount",
			goal: &Goal{
				TargetAmount:    10000000,
				CurrentAmount:   10000000,
				RemainingAmount: 0,
				TargetDate:      &future,
			},
			frequency: FrequencyMonthly,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goal.CalculateSuggestedContribution(tt.frequency)
			// Allow 10% tolerance for calculation differences
			assert.InDelta(t, tt.expected, result, tt.expected*0.1)
		})
	}
}

func TestGoal_AddContribution(t *testing.T) {
	goal := &Goal{
		TargetAmount:  10000000,
		CurrentAmount: 5000000,
		Status:        GoalStatusActive,
	}

	goal.AddContribution(2000000)

	assert.Equal(t, 7000000.0, goal.CurrentAmount)
	assert.Equal(t, 3000000.0, goal.RemainingAmount)
	assert.Equal(t, 70.0, goal.PercentageComplete)
}

func TestGoalType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		goalType GoalType
		expected bool
	}{
		{"valid - savings", GoalTypeSavings, true},
		{"valid - debt", GoalTypeDebt, true},
		{"valid - investment", GoalTypeInvestment, true},
		{"valid - purchase", GoalTypePurchase, true},
		{"valid - emergency", GoalTypeEmergency, true},
		{"valid - retirement", GoalTypeRetirement, true},
		{"valid - education", GoalTypeEducation, true},
		{"valid - other", GoalTypeOther, true},
		{"invalid - empty", GoalType(""), false},
		{"invalid - unknown", GoalType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goalType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoalStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   GoalStatus
		expected bool
	}{
		{"valid - active", GoalStatusActive, true},
		{"valid - completed", GoalStatusCompleted, true},
		{"valid - paused", GoalStatusPaused, true},
		{"valid - cancelled", GoalStatusCancelled, true},
		{"valid - overdue", GoalStatusOverdue, true},
		{"invalid - empty", GoalStatus(""), false},
		{"invalid - unknown", GoalStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoalPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority GoalPriority
		expected bool
	}{
		{"valid - low", GoalPriorityLow, true},
		{"valid - medium", GoalPriorityMedium, true},
		{"valid - high", GoalPriorityHigh, true},
		{"valid - critical", GoalPriorityCritical, true},
		{"invalid - empty", GoalPriority(""), false},
		{"invalid - unknown", GoalPriority("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.priority.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContributionFrequency_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		frequency ContributionFrequency
		expected  bool
	}{
		{"valid - one_time", FrequencyOneTime, true},
		{"valid - daily", FrequencyDaily, true},
		{"valid - weekly", FrequencyWeekly, true},
		{"valid - biweekly", FrequencyBiweekly, true},
		{"valid - monthly", FrequencyMonthly, true},
		{"valid - quarterly", FrequencyQuarterly, true},
		{"valid - yearly", FrequencyYearly, true},
		{"invalid - empty", ContributionFrequency(""), false},
		{"invalid - unknown", ContributionFrequency("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.frequency.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContributionFrequency_DaysPerPeriod(t *testing.T) {
	tests := []struct {
		name      string
		frequency ContributionFrequency
		expected  int
	}{
		{"daily", FrequencyDaily, 1},
		{"weekly", FrequencyWeekly, 7},
		{"biweekly", FrequencyBiweekly, 14},
		{"monthly", FrequencyMonthly, 30},
		{"quarterly", FrequencyQuarterly, 90},
		{"yearly", FrequencyYearly, 365},
		{"one_time", FrequencyOneTime, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.frequency.DaysPerPeriod()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoal_TableName(t *testing.T) {
	goal := Goal{}
	assert.Equal(t, "goals", goal.TableName())
}

func TestGoal_NewGoal(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()

	goal := &Goal{
		UserID:       userID,
		Name:         "Emergency Fund",
		Type:         GoalTypeEmergency,
		Priority:     GoalPriorityHigh,
		TargetAmount: 50000000,
		StartDate:    startDate,
		Status:       GoalStatusActive,
	}

	assert.NotNil(t, goal)
	assert.Equal(t, userID, goal.UserID)
	assert.Equal(t, "Emergency Fund", goal.Name)
	assert.Equal(t, GoalTypeEmergency, goal.Type)
	assert.Equal(t, GoalPriorityHigh, goal.Priority)
	assert.Equal(t, 50000000.0, goal.TargetAmount)
	assert.Equal(t, GoalStatusActive, goal.Status)
}
