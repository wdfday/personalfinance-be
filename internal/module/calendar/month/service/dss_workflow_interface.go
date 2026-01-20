package service

import (
	"context"

	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
)

// MonthDSSWorkflowHandler defines the interface for the new sequential 5-step DSS workflow
// Step 0: Auto-Scoring → Step 1: Goal Prioritization → Step 2: Debt Strategy → Step 3: Trade-off → Step 4: Budget
type MonthDSSWorkflowHandler interface {
	// ===== Step 0: Auto-Scoring Preview (Optional) =====

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

	// ===== Step 3: Goal-Debt Trade-off =====

	// PreviewGoalDebtTradeoff runs Monte Carlo trade-off analysis
	PreviewGoalDebtTradeoff(ctx context.Context, req dto.PreviewGoalDebtTradeoffRequest, userID *uuid.UUID) (*dto.PreviewGoalDebtTradeoffResponse, error)

	// ApplyGoalDebtTradeoff saves the user's allocation decision to MonthState
	ApplyGoalDebtTradeoff(ctx context.Context, req dto.ApplyGoalDebtTradeoffRequest, userID *uuid.UUID) error

	// ===== Step 4: Budget Allocation =====

	// PreviewBudgetAllocation runs Goal Programming allocation with results from steps 1-3
	PreviewBudgetAllocation(ctx context.Context, req dto.PreviewBudgetAllocationRequest, userID *uuid.UUID) (*dto.PreviewBudgetAllocationResponse, error)

	// ApplyBudgetAllocation applies the selected allocation scenario to MonthState categories
	ApplyBudgetAllocation(ctx context.Context, req dto.ApplyBudgetAllocationRequest, userID *uuid.UUID) error

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
