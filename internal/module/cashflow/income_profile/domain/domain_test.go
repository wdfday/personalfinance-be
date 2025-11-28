package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncomeProfile_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		profile  *IncomeProfile
		expected bool
	}{
		{
			name:     "active profile",
			profile:  &IncomeProfile{IsActive: true},
			expected: true,
		},
		{
			name:     "inactive profile",
			profile:  &IncomeProfile{IsActive: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncomeProfile_IsVerified(t *testing.T) {
	tests := []struct {
		name     string
		profile  *IncomeProfile
		expected bool
	}{
		{
			name:     "verified profile",
			profile:  &IncomeProfile{IsVerified: true},
			expected: true,
		},
		{
			name:     "unverified profile",
			profile:  &IncomeProfile{IsVerified: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.IsVerified()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncomeType_IsValid(t *testing.T) {
	assert.True(t, IncomeTypeSalary.IsValid())
	assert.True(t, IncomeTypeFreelance.IsValid())
	assert.True(t, IncomeTypeBusiness.IsValid())
	assert.False(t, IncomeType("invalid").IsValid())
}

func TestIncomeFrequency_IsValid(t *testing.T) {
	assert.True(t, FrequencyMonthly.IsValid())
	assert.True(t, FrequencyWeekly.IsValid())
	assert.False(t, IncomeFrequency("invalid").IsValid())
}

func TestIncomeProfile_TableName(t *testing.T) {
	profile := IncomeProfile{}
	assert.Equal(t, "income_profiles", profile.TableName())
}
