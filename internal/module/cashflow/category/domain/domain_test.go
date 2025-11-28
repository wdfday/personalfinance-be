package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCategory_IsParent(t *testing.T) {
	tests := []struct {
		name     string
		category *Category
		expected bool
	}{
		{
			name:     "parent category",
			category: &Category{ParentID: nil},
			expected: true,
		},
		{
			name: "child category",
			category: &Category{
				ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.category.IsParent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCategory_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		category *Category
		expected bool
	}{
		{
			name:     "active category",
			category: &Category{IsActive: true},
			expected: true,
		},
		{
			name:     "inactive category",
			category: &Category{IsActive: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.category.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCategoryType_IsValid(t *testing.T) {
	assert.True(t, CategoryTypeIncome.IsValid())
	assert.True(t, CategoryTypeExpense.IsValid())
	assert.False(t, CategoryType("invalid").IsValid())
}

func TestCategory_TableName(t *testing.T) {
	category := Category{}
	assert.Equal(t, "categories", category.TableName())
}
