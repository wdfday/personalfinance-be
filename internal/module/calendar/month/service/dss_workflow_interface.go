package service

import (
	"context"

	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
)

// MonthDSSWorkflowHandler defines the interface for the DSS workflow
// Initialize → Step 0: Auto-Scoring → Step 1: Goal Prioritization → Step 2: Debt Strategy → Step 3: Budget Allocation → Finalize
type MonthDSSWorkflowHandler interface {
	// ===== Initialize DSS Workflow =====

	// InitializeDSS creates a new DSS session with input snapshot cached in Redis (3h TTL)
	// This MUST be called FIRST before any preview/apply steps
	InitializeDSS(ctx context.Context, req dto.InitializeDSSRequest, userID *uuid.UUID) (*dto.InitializeDSSResponse, error)

	// ===== Step 0: Auto-Scoring Preview (Optional, Preview-Only) =====

	// PreviewAutoScoring runs goal auto-scoring and returns scored goals (preview only, not saved)
	PreviewAutoScoring(ctx context.Context, req dto.PreviewAutoScoringRequest, userID *uuid.UUID) (*dto.PreviewAutoScoringResponse, error)

	// ===== Step 1: Goal Prioritization =====

	// PreviewGoalPrioritization runs AHP goal prioritization and returns preview
	PreviewGoalPrioritization(ctx context.Context, req dto.PreviewGoalPrioritizationRequest, userID *uuid.UUID) (*dto.PreviewGoalPrioritizationResponse, error)

	// ApplyGoalPrioritization saves the user's accepted goal ranking to MonthState
	ApplyGoalPrioritization(ctx context.Context, req dto.ApplyGoalPrioritizationRequest, userID *uuid.UUID) error

	// ===== Step 2: Debt Strategy =====

	// PreviewDebtStrategy runs debt repayment simulations and returns scenarios
	PreviewDebtStrategy(ctx context.Context, req dto.PreviewDebtStrategyRequest, userID *uuid.UUID) (*dto.PreviewDebtStrategyResponse, error)

	// ApplyDebtStrategy saves the user's selected debt strategy to MonthState
	ApplyDebtStrategy(ctx context.Context, req dto.ApplyDebtStrategyRequest, userID *uuid.UUID) error

	// ===== Step 3: Budget Allocation =====

	// PreviewBudgetAllocation runs Goal Programming allocation with 3 scenarios
	PreviewBudgetAllocation(ctx context.Context, req dto.PreviewBudgetAllocationRequest, userID *uuid.UUID) (*dto.PreviewBudgetAllocationResponse, error)

	// ===== Workflow Management =====

	// GetDSSWorkflowStatus returns the current workflow state
	GetDSSWorkflowStatus(ctx context.Context, monthID uuid.UUID, userID *uuid.UUID) (*dto.DSSWorkflowStatusResponse, error)

	// ResetDSSWorkflow clears all DSS workflow results and starts over
	ResetDSSWorkflow(ctx context.Context, req dto.ResetDSSWorkflowRequest, userID *uuid.UUID) error

	// ===== DSS Finalization (Approach 2: Apply All at Once) =====

	// FinalizeDSS applies all DSS results at once and creates a new MonthState version
	// This is the recommended approach: Preview all steps → Review → Finalize once
	FinalizeDSS(ctx context.Context, req dto.FinalizeDSSRequest, monthID uuid.UUID, userID *uuid.UUID) (*dto.FinalizeDSSResponse, error)
}
