package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== BudgetConstraint Tests ====================

func TestNewBudgetConstraint(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully create budget constraint", func(t *testing.T) {
		bc, err := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		require.NoError(t, err)
		assert.NotNil(t, bc)
		assert.Equal(t, userID, bc.UserID)
		assert.Equal(t, categoryID, bc.CategoryID)
		assert.Equal(t, 1000.00, bc.MinimumAmount)
		assert.Equal(t, "monthly", bc.Period)
		assert.False(t, bc.IsFlexible)
		assert.Equal(t, ConstraintStatusActive, bc.Status)
		assert.True(t, bc.IsRecurring)
		assert.Equal(t, 99, bc.Priority)
	})

	t.Run("error when user ID is nil", func(t *testing.T) {
		bc, err := NewBudgetConstraint(uuid.Nil, categoryID, 1000.00, startDate)

		assert.Error(t, err)
		assert.Nil(t, bc)
		assert.Equal(t, ErrInvalidUserID, err)
	})

	t.Run("error when category ID is nil", func(t *testing.T) {
		bc, err := NewBudgetConstraint(userID, uuid.Nil, 1000.00, startDate)

		assert.Error(t, err)
		assert.Nil(t, bc)
		assert.Equal(t, ErrInvalidCategoryID, err)
	})

	t.Run("error when minimum amount is negative", func(t *testing.T) {
		bc, err := NewBudgetConstraint(userID, categoryID, -100.00, startDate)

		assert.Error(t, err)
		assert.Nil(t, bc)
		assert.Equal(t, ErrNegativeAmount, err)
	})
}

func TestNewFlexibleBudgetConstraint(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully create flexible constraint", func(t *testing.T) {
		bc, err := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)

		require.NoError(t, err)
		assert.NotNil(t, bc)
		assert.True(t, bc.IsFlexible)
		assert.Equal(t, 1000.00, bc.MinimumAmount)
		assert.Equal(t, 2000.00, bc.MaximumAmount)
	})

	t.Run("error when max is below min", func(t *testing.T) {
		bc, err := NewFlexibleBudgetConstraint(userID, categoryID, 2000.00, 1000.00, startDate)

		assert.Error(t, err)
		assert.Nil(t, bc)
		assert.Equal(t, ErrMaximumBelowMinimum, err)
	})
}

func TestBudgetConstraint_GetRange(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed constraint returns same min and max", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		min, max := bc.GetRange()

		assert.Equal(t, 1000.00, min)
		assert.Equal(t, 1000.00, max)
	})

	t.Run("flexible constraint returns proper range", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)

		min, max := bc.GetRange()

		assert.Equal(t, 1000.00, min)
		assert.Equal(t, 2000.00, max)
	})

	t.Run("flexible without max returns 0 for unlimited", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.IsFlexible = true
		bc.MaximumAmount = 0

		min, max := bc.GetRange()

		assert.Equal(t, 1000.00, min)
		assert.Equal(t, 0.0, max) // 0 = no limit
	})
}

func TestBudgetConstraint_IsFixed(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed constraint", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.True(t, bc.IsFixed())
	})

	t.Run("flexible constraint", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.False(t, bc.IsFixed())
	})
}

func TestBudgetConstraint_HasUpperLimit(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed has no upper limit", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.False(t, bc.HasUpperLimit())
	})

	t.Run("flexible with max has upper limit", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.True(t, bc.HasUpperLimit())
	})

	t.Run("flexible without max has no upper limit", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.IsFlexible = true
		bc.MaximumAmount = 0
		assert.False(t, bc.HasUpperLimit())
	})
}

func TestBudgetConstraint_GetFlexibilityRange(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed returns 0", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.Equal(t, 0.0, bc.GetFlexibilityRange())
	})

	t.Run("flexible with max returns difference", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2500.00, startDate)
		assert.Equal(t, 1500.0, bc.GetFlexibilityRange())
	})

	t.Run("flexible without max returns 0", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.IsFlexible = true
		bc.MaximumAmount = 0
		assert.Equal(t, 0.0, bc.GetFlexibilityRange())
	})
}

func TestBudgetConstraint_CanAllocate(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed - exact amount allowed", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.True(t, bc.CanAllocate(1000.00))
	})

	t.Run("fixed - below minimum not allowed", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.False(t, bc.CanAllocate(500.00))
	})

	t.Run("fixed - above minimum not allowed", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.False(t, bc.CanAllocate(1500.00))
	})

	t.Run("flexible - within range allowed", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.True(t, bc.CanAllocate(1500.00))
	})

	t.Run("flexible - below minimum not allowed", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.False(t, bc.CanAllocate(500.00))
	})

	t.Run("flexible - above maximum not allowed", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.False(t, bc.CanAllocate(2500.00))
	})

	t.Run("flexible - unlimited allows any above min", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.IsFlexible = true
		bc.MaximumAmount = 0
		assert.True(t, bc.CanAllocate(5000.00))
	})
}

func TestBudgetConstraint_UpdateMinimum(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully update minimum", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.UpdateMinimum(1500.00)

		assert.NoError(t, err)
		assert.Equal(t, 1500.00, bc.MinimumAmount)
	})

	t.Run("error on negative amount", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.UpdateMinimum(-100.00)

		assert.Error(t, err)
		assert.Equal(t, ErrNegativeAmount, err)
	})

	t.Run("error when min exceeds max on flexible", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)

		err := bc.UpdateMinimum(2500.00)

		assert.Error(t, err)
		assert.Equal(t, ErrMinimumExceedsMaximum, err)
	})
}

func TestBudgetConstraint_SetFlexible(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully make flexible", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.SetFlexible(2000.00)

		assert.NoError(t, err)
		assert.True(t, bc.IsFlexible)
		assert.Equal(t, 2000.00, bc.MaximumAmount)
	})

	t.Run("error when max below min", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.SetFlexible(500.00)

		assert.Error(t, err)
		assert.Equal(t, ErrMaximumBelowMinimum, err)
	})
}

func TestBudgetConstraint_SetFixed(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully make fixed", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)

		bc.SetFixed()

		assert.False(t, bc.IsFlexible)
		assert.Equal(t, 0.0, bc.MaximumAmount)
	})
}

func TestBudgetConstraint_SetPriority(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("successfully set priority", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.SetPriority(5)

		assert.NoError(t, err)
		assert.Equal(t, 5, bc.Priority)
	})

	t.Run("error on invalid priority", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		err := bc.SetPriority(0)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPriority, err)
	})
}

func TestBudgetConstraint_Archive(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()
	archivedBy := uuid.New()

	t.Run("successfully archive", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		bc.Archive(archivedBy)

		assert.Equal(t, ConstraintStatusArchived, bc.Status)
		assert.NotNil(t, bc.ArchivedAt)
		assert.Equal(t, &archivedBy, bc.ArchivedBy)
	})
}

func TestBudgetConstraint_IsActive(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()

	t.Run("active constraint", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(-time.Hour))
		assert.True(t, bc.IsActive())
	})

	t.Run("future start date not active", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(time.Hour))
		assert.False(t, bc.IsActive())
	})

	t.Run("past end date not active", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(-2*time.Hour))
		pastDate := time.Now().Add(-time.Hour)
		bc.EndDate = &pastDate
		assert.False(t, bc.IsActive())
	})

	t.Run("archived not active", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(-time.Hour))
		bc.Status = ConstraintStatusArchived
		assert.False(t, bc.IsActive())
	})
}

func TestBudgetConstraint_IsArchived(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("not archived", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.False(t, bc.IsArchived())
	})

	t.Run("archived by status", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.Status = ConstraintStatusArchived
		assert.True(t, bc.IsArchived())
	})

	t.Run("archived by timestamp", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		now := time.Now()
		bc.ArchivedAt = &now
		assert.True(t, bc.IsArchived())
	})
}

func TestBudgetConstraint_CreateNewVersion(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("creates new version with correct data", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)

		newVersion := bc.CreateNewVersion()

		assert.NotEqual(t, bc.ID, newVersion.ID)
		assert.Equal(t, bc.UserID, newVersion.UserID)
		assert.Equal(t, bc.CategoryID, newVersion.CategoryID)
		assert.Equal(t, bc.MinimumAmount, newVersion.MinimumAmount)
		assert.Equal(t, &bc.ID, newVersion.PreviousVersionID)
		assert.Equal(t, ConstraintStatusActive, newVersion.Status)
	})
}

func TestBudgetConstraint_String(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	startDate := time.Now()

	t.Run("fixed constraint string", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		assert.Contains(t, bc.String(), "Fixed")
	})

	t.Run("flexible with limit string", func(t *testing.T) {
		bc, _ := NewFlexibleBudgetConstraint(userID, categoryID, 1000.00, 2000.00, startDate)
		assert.Contains(t, bc.String(), "Flexible")
	})

	t.Run("flexible unlimited string", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, startDate)
		bc.IsFlexible = true
		bc.MaximumAmount = 0
		assert.Contains(t, bc.String(), "no limit")
	})
}

func TestBudgetConstraint_TableName(t *testing.T) {
	bc := BudgetConstraint{}
	assert.Equal(t, "budget_constraints", bc.TableName())
}

func TestBudgetConstraint_CheckAndArchiveIfEnded(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()

	t.Run("archives ended constraint", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(-2*time.Hour))
		pastDate := time.Now().Add(-time.Hour)
		bc.EndDate = &pastDate

		result := bc.CheckAndArchiveIfEnded()

		assert.True(t, result)
		assert.Equal(t, ConstraintStatusEnded, bc.Status)
		assert.NotNil(t, bc.ArchivedAt)
	})

	t.Run("does not archive active constraint", func(t *testing.T) {
		bc, _ := NewBudgetConstraint(userID, categoryID, 1000.00, time.Now().Add(-time.Hour))

		result := bc.CheckAndArchiveIfEnded()

		assert.False(t, result)
		assert.Equal(t, ConstraintStatusActive, bc.Status)
	})
}

// ==================== BudgetConstraints Collection Tests ====================

func TestBudgetConstraints_TotalMandatoryExpenses(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()

	bc1, _ := NewBudgetConstraint(userID, uuid.New(), 1000.00, startDate)
	bc2, _ := NewBudgetConstraint(userID, uuid.New(), 2000.00, startDate)
	bc3, _ := NewBudgetConstraint(userID, uuid.New(), 500.00, startDate)

	bcs := BudgetConstraints{bc1, bc2, bc3}

	assert.Equal(t, 3500.0, bcs.TotalMandatoryExpenses())
}

func TestBudgetConstraints_GetByCategory(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()
	cat1 := uuid.New()
	cat2 := uuid.New()

	bc1, _ := NewBudgetConstraint(userID, cat1, 1000.00, startDate)
	bc2, _ := NewBudgetConstraint(userID, cat2, 2000.00, startDate)

	bcs := BudgetConstraints{bc1, bc2}

	t.Run("found", func(t *testing.T) {
		result := bcs.GetByCategory(cat1)
		assert.NotNil(t, result)
		assert.Equal(t, cat1, result.CategoryID)
	})

	t.Run("not found", func(t *testing.T) {
		result := bcs.GetByCategory(uuid.New())
		assert.Nil(t, result)
	})
}

func TestBudgetConstraints_GetFlexible(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()

	bc1, _ := NewBudgetConstraint(userID, uuid.New(), 1000.00, startDate)
	bc2, _ := NewFlexibleBudgetConstraint(userID, uuid.New(), 1000.00, 2000.00, startDate)
	bc3, _ := NewFlexibleBudgetConstraint(userID, uuid.New(), 500.00, 1000.00, startDate)

	bcs := BudgetConstraints{bc1, bc2, bc3}

	flexible := bcs.GetFlexible()
	assert.Len(t, flexible, 2)
}

func TestBudgetConstraints_GetFixed(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()

	bc1, _ := NewBudgetConstraint(userID, uuid.New(), 1000.00, startDate)
	bc2, _ := NewFlexibleBudgetConstraint(userID, uuid.New(), 1000.00, 2000.00, startDate)
	bc3, _ := NewBudgetConstraint(userID, uuid.New(), 500.00, startDate)

	bcs := BudgetConstraints{bc1, bc2, bc3}

	fixed := bcs.GetFixed()
	assert.Len(t, fixed, 2)
}

func TestBudgetConstraints_SortByPriority(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()

	bc1, _ := NewBudgetConstraint(userID, uuid.New(), 1000.00, startDate)
	bc1.Priority = 5
	bc2, _ := NewBudgetConstraint(userID, uuid.New(), 2000.00, startDate)
	bc2.Priority = 1
	bc3, _ := NewBudgetConstraint(userID, uuid.New(), 500.00, startDate)
	bc3.Priority = 3

	bcs := BudgetConstraints{bc1, bc2, bc3}

	sorted := bcs.SortByPriority()

	assert.Equal(t, 1, sorted[0].Priority)
	assert.Equal(t, 3, sorted[1].Priority)
	assert.Equal(t, 5, sorted[2].Priority)
}

func TestBudgetConstraints_HasDuplicateCategories(t *testing.T) {
	userID := uuid.New()
	startDate := time.Now()
	cat1 := uuid.New()

	t.Run("no duplicates", func(t *testing.T) {
		bc1, _ := NewBudgetConstraint(userID, uuid.New(), 1000.00, startDate)
		bc2, _ := NewBudgetConstraint(userID, uuid.New(), 2000.00, startDate)

		bcs := BudgetConstraints{bc1, bc2}
		assert.False(t, bcs.HasDuplicateCategories())
	})

	t.Run("has duplicates", func(t *testing.T) {
		bc1, _ := NewBudgetConstraint(userID, cat1, 1000.00, startDate)
		bc2, _ := NewBudgetConstraint(userID, cat1, 2000.00, startDate)

		bcs := BudgetConstraints{bc1, bc2}
		assert.True(t, bcs.HasDuplicateCategories())
	})
}

func TestBudgetConstraints_Validate(t *testing.T) {
	startDate := time.Now()

	t.Run("all valid", func(t *testing.T) {
		bc1, _ := NewBudgetConstraint(uuid.New(), uuid.New(), 1000.00, startDate)
		bc2, _ := NewBudgetConstraint(uuid.New(), uuid.New(), 2000.00, startDate)

		bcs := BudgetConstraints{bc1, bc2}
		assert.NoError(t, bcs.Validate())
	})

	t.Run("contains invalid", func(t *testing.T) {
		bc1, _ := NewBudgetConstraint(uuid.New(), uuid.New(), 1000.00, startDate)
		bc2 := &BudgetConstraint{
			UserID:     uuid.Nil, // Invalid
			CategoryID: uuid.New(),
		}

		bcs := BudgetConstraints{bc1, bc2}
		assert.Error(t, bcs.Validate())
	})
}
