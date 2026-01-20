package domain

import (
	"github.com/google/uuid"
)

// ConstraintSource indicates where the constraint came from
type ConstraintSource string

const (
	// ConstraintSourceRecurring = copied from budget template (hàng tháng)
	ConstraintSourceRecurring ConstraintSource = "recurring"
	// ConstraintSourceOneOff = added for this month only (du lịch, đột xuất)
	ConstraintSourceOneOff ConstraintSource = "one_off"
)

// IsValid checks if the constraint source is valid
func (s ConstraintSource) IsValid() bool {
	switch s {
	case ConstraintSourceRecurring, ConstraintSourceOneOff:
		return true
	}
	return false
}

// CategoryConstraint represents a budget constraint for a category in a month
// These are INPUT for DSS to calculate optimal budget allocation
type CategoryConstraint struct {
	ID           uuid.UUID        `json:"id"` // Unique constraint ID
	CategoryID   uuid.UUID        `json:"category_id"`
	CategoryName string           `json:"category_name"`
	Source       ConstraintSource `json:"source"` // "recurring" or "one_off"

	// Constraint values (used by DSS for optimization)
	MinAmount    *float64 `json:"min_amount,omitempty"`    // Minimum required
	MaxAmount    *float64 `json:"max_amount,omitempty"`    // Maximum allowed
	TargetAmount *float64 `json:"target_amount,omitempty"` // Ideal target
	Priority     int      `json:"priority"`                // 1-10, higher = more important

	// Flags
	IsRequired bool `json:"is_required"` // Must be allocated (true = hard constraint)
	IsFlexible bool `json:"is_flexible"` // Can be reduced if needed

	// Context
	Notes *string `json:"notes,omitempty"` // VD: "Du lịch Đà Lạt", "Sửa xe"
}

// NewRecurringConstraint creates a constraint copied from budget template
func NewRecurringConstraint(categoryID uuid.UUID, categoryName string, target float64, priority int) *CategoryConstraint {
	return &CategoryConstraint{
		ID:           uuid.New(),
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Source:       ConstraintSourceRecurring,
		TargetAmount: &target,
		Priority:     priority,
		IsRequired:   true,
		IsFlexible:   false,
	}
}

// NewOneOffConstraint creates a constraint for this month only
func NewOneOffConstraint(categoryID uuid.UUID, categoryName string, amount float64, notes string) *CategoryConstraint {
	return &CategoryConstraint{
		ID:           uuid.New(),
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Source:       ConstraintSourceOneOff,
		TargetAmount: &amount,
		Priority:     5, // Default mid priority
		IsRequired:   false,
		IsFlexible:   true,
		Notes:        &notes,
	}
}

// Validate validates the constraint
func (c *CategoryConstraint) Validate() error {
	if c.CategoryID == uuid.Nil {
		return ErrInvalidCategoryID
	}
	if c.Priority < 1 || c.Priority > 10 {
		return ErrInvalidPriority
	}
	if !c.Source.IsValid() {
		return ErrInvalidConstraintSource
	}
	return nil
}

// Errors
var (
	ErrInvalidCategoryID       = NewDomainError("invalid category ID")
	ErrInvalidPriority         = NewDomainError("priority must be between 1 and 10")
	ErrInvalidConstraintSource = NewDomainError("invalid constraint source")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Message string
}

func NewDomainError(msg string) *DomainError {
	return &DomainError{Message: msg}
}

func (e *DomainError) Error() string {
	return e.Message
}
