package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==================== Workflow Management ====================

// DSSWorkflowStatusResponse shows the current DSS workflow state
type DSSWorkflowStatusResponse struct {
	MonthID        uuid.UUID  `json:"month_id"`
	CurrentStep    int        `json:"current_step"`    // 0-3 (0=not started, 1-3=steps)
	CompletedSteps []int      `json:"completed_steps"` // Which steps have been applied
	CanProceed     bool       `json:"can_proceed"`     // Can proceed to next step?
	IsComplete     bool       `json:"is_complete"`     // All steps completed?
	LastUpdated    *time.Time `json:"last_updated,omitempty"`

	// Step status details
	Step1Applied bool `json:"step1_applied"` // Goal Prioritization
	Step2Applied bool `json:"step2_applied"` // Debt Strategy
	Step3Applied bool `json:"step3_applied"` // Goal-Debt Trade-off (removed, always false)
	Step4Applied bool `json:"step4_applied"` // Budget Allocation (Step 3 trong workflow mới)
}

// ResetDSSWorkflowRequest resets the DSS workflow
type ResetDSSWorkflowRequest struct {
	MonthID uuid.UUID `json:"month_id" binding:"required"`
}
