package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestItemType_Constants(t *testing.T) {
	assert.Equal(t, ItemType("CONSTRAINT"), ItemTypeConstraint)
	assert.Equal(t, ItemType("GOAL"), ItemTypeGoal)
	assert.Equal(t, ItemType("DEBT"), ItemTypeDebt)
}

func TestCategoryState_RecalculateAvailable(t *testing.T) {
	catState := &CategoryState{
		CategoryID: uuid.New(),
		Type:       ItemTypeConstraint,
		Rollover:   1000,
		Assigned:   5000,
		Activity:   -3000,
	}

	catState.RecalculateAvailable()

	// Available = Rollover + Assigned + Activity
	// 1000 + 5000 + (-3000) = 3000
	assert.Equal(t, float64(3000), catState.Available)
}

func TestCategoryState_IsOverspent(t *testing.T) {
	tests := []struct {
		name      string
		rollover  float64
		assigned  float64
		activity  float64
		overspent bool
	}{
		{
			name:      "Not overspent",
			rollover:  1000,
			assigned:  5000,
			activity:  -3000,
			overspent: false,
		},
		{
			name:      "Overspent",
			rollover:  1000,
			assigned:  2000,
			activity:  -4000,
			overspent: true,
		},
		{
			name:      "Exactly zero",
			rollover:  0,
			assigned:  3000,
			activity:  -3000,
			overspent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catState := &CategoryState{
				Rollover: tt.rollover,
				Assigned: tt.assigned,
				Activity: tt.activity,
			}
			catState.RecalculateAvailable()

			assert.Equal(t, tt.overspent, catState.IsOverspent())
		})
	}
}

func TestCategoryState_AddAssignment(t *testing.T) {
	catState := &CategoryState{
		CategoryID: uuid.New(),
		Type:       ItemTypeConstraint,
		Rollover:   1000,
		Assigned:   5000,
		Activity:   0,
	}
	catState.RecalculateAvailable()

	initialAvailable := catState.Available
	catState.AddAssignment(2000)

	assert.Equal(t, float64(7000), catState.Assigned)
	assert.Equal(t, initialAvailable+2000, catState.Available)
}

func TestCategoryState_AddActivity(t *testing.T) {
	catState := &CategoryState{
		CategoryID: uuid.New(),
		Type:       ItemTypeConstraint,
		Rollover:   1000,
		Assigned:   5000,
		Activity:   -1000,
	}
	catState.RecalculateAvailable()

	initialAvailable := catState.Available
	catState.AddActivity(-2000)

	assert.Equal(t, float64(-3000), catState.Activity)
	assert.Equal(t, initialAvailable-2000, catState.Available)
}

func TestCategoryState_DeepCopyConfig(t *testing.T) {
	goalTarget := float64(10000)
	debtMinPayment := float64(500)
	notes := "Test notes"

	original := &CategoryState{
		CategoryID:     uuid.New(),
		Type:           ItemTypeGoal,
		Rollover:       1000,
		Assigned:       5000,
		Activity:       -3000,
		GoalTarget:     &goalTarget,
		DebtMinPayment: &debtMinPayment,
		Notes:          &notes,
	}
	original.RecalculateAvailable()

	copied := original.DeepCopyConfig()

	// Config should be copied
	assert.Equal(t, original.CategoryID, copied.CategoryID)
	assert.Equal(t, *original.GoalTarget, *copied.GoalTarget)
	assert.Equal(t, *original.DebtMinPayment, *copied.DebtMinPayment)
	assert.Equal(t, *original.Notes, *copied.Notes)

	// Amounts should be reset
	assert.Equal(t, float64(0), copied.Rollover)
	assert.Equal(t, float64(0), copied.Assigned)
	assert.Equal(t, float64(0), copied.Activity)
	assert.Equal(t, float64(0), copied.Available)

	// Modify copied should not affect original
	newNotes := "Modified"
	copied.Notes = &newNotes
	assert.NotEqual(t, *original.Notes, *copied.Notes)
}

func TestCategoryState_SetRollover(t *testing.T) {
	catState := &CategoryState{
		CategoryID: uuid.New(),
		Type:       ItemTypeConstraint,
		Rollover:   0,
		Assigned:   0,
		Activity:   0,
	}

	catState.SetRollover(5000)

	assert.Equal(t, float64(5000), catState.Rollover)
	assert.Equal(t, float64(5000), catState.Available)
}

func TestNewCategoryState(t *testing.T) {
	categoryID := uuid.New()
	catState := NewCategoryState(categoryID, 1000, 5000, -3000)

	assert.Equal(t, categoryID, catState.CategoryID)
	assert.Equal(t, float64(1000), catState.Rollover)
	assert.Equal(t, float64(5000), catState.Assigned)
	assert.Equal(t, float64(-3000), catState.Activity)
	assert.Equal(t, float64(3000), catState.Available)
}
