package dto

import (
	"personalfinancedss/internal/module/calendar/month/domain"

	"github.com/google/uuid"
)

// ========== Constraint Requests ==========

// AddConstraintRequest adds a new constraint to a month
type AddConstraintRequest struct {
	MonthID      uuid.UUID `json:"month_id" binding:"required"`
	CategoryID   uuid.UUID `json:"category_id" binding:"required"`
	CategoryName string    `json:"category_name" binding:"required"`

	// Constraint values
	MinAmount    *float64 `json:"min_amount,omitempty"`
	MaxAmount    *float64 `json:"max_amount,omitempty"`
	TargetAmount *float64 `json:"target_amount,omitempty"`
	Priority     int      `json:"priority" binding:"min=1,max=10"`

	// Flags
	IsRequired bool `json:"is_required"`
	IsFlexible bool `json:"is_flexible"`

	// For one-off constraints
	Notes *string `json:"notes,omitempty"` // VD: "Du lịch Đà Lạt"
}

// UpdateConstraintRequest updates an existing constraint
type UpdateConstraintRequest struct {
	MonthID      uuid.UUID `json:"month_id" binding:"required"`
	ConstraintID uuid.UUID `json:"constraint_id" binding:"required"`

	// Fields to update
	MinAmount    *float64 `json:"min_amount,omitempty"`
	MaxAmount    *float64 `json:"max_amount,omitempty"`
	TargetAmount *float64 `json:"target_amount,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	IsRequired   *bool    `json:"is_required,omitempty"`
	IsFlexible   *bool    `json:"is_flexible,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
}

// RemoveConstraintRequest removes a constraint from a month
type RemoveConstraintRequest struct {
	MonthID      uuid.UUID `json:"month_id" binding:"required"`
	ConstraintID uuid.UUID `json:"constraint_id" binding:"required"`
}

// ========== Constraint Responses ==========

// ConstraintResponse represents a single constraint
type ConstraintResponse struct {
	ID           uuid.UUID               `json:"id"`
	CategoryID   uuid.UUID               `json:"category_id"`
	CategoryName string                  `json:"category_name"`
	Source       domain.ConstraintSource `json:"source"`

	MinAmount    *float64 `json:"min_amount,omitempty"`
	MaxAmount    *float64 `json:"max_amount,omitempty"`
	TargetAmount *float64 `json:"target_amount,omitempty"`
	Priority     int      `json:"priority"`

	IsRequired bool    `json:"is_required"`
	IsFlexible bool    `json:"is_flexible"`
	Notes      *string `json:"notes,omitempty"`
}

// ConstraintListResponse lists all constraints for a month
type ConstraintListResponse struct {
	MonthID          uuid.UUID            `json:"month_id"`
	Month            string               `json:"month"`
	TotalConstraints int                  `json:"total_constraints"`
	RecurringCount   int                  `json:"recurring_count"`
	OneOffCount      int                  `json:"one_off_count"`
	Constraints      []ConstraintResponse `json:"constraints"`
}

// ToConstraintResponse converts domain to DTO
func ToConstraintResponse(c *domain.CategoryConstraint) ConstraintResponse {
	return ConstraintResponse{
		ID:           c.ID,
		CategoryID:   c.CategoryID,
		CategoryName: c.CategoryName,
		Source:       c.Source,
		MinAmount:    c.MinAmount,
		MaxAmount:    c.MaxAmount,
		TargetAmount: c.TargetAmount,
		Priority:     c.Priority,
		IsRequired:   c.IsRequired,
		IsFlexible:   c.IsFlexible,
		Notes:        c.Notes,
	}
}
