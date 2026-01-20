package dto

import (
	tradeoffDto "personalfinancedss/internal/module/analytics/debt_tradeoff/dto"

	"github.com/google/uuid"
)

// ==================== Step 3: Goal-Debt Trade-off ====================
// Import types from analytics debt_tradeoff module - NO DUPLICATION

// PreviewGoalDebtTradeoffRequest requests trade-off analysis
// Wraps user preferences - data from Step 1 & 2 auto-collected from DSSWorkflow
type PreviewGoalDebtTradeoffRequest struct {
	MonthID     uuid.UUID                       `json:"month_id" binding:"required"`
	Preferences tradeoffDto.TradeoffPreferences `json:"preferences" binding:"required"`
}

// PreviewGoalDebtTradeoffResponse type alias to analytics output
type PreviewGoalDebtTradeoffResponse = tradeoffDto.TradeoffOutput

// ApplyGoalDebtTradeoffRequest applies the user's trade-off decision
type ApplyGoalDebtTradeoffRequest struct {
	MonthID               uuid.UUID `json:"month_id" binding:"required"`
	GoalAllocationPercent float64   `json:"goal_allocation_percent" binding:"required,gte=0,lte=100"`
	DebtAllocationPercent float64   `json:"debt_allocation_percent" binding:"required,gte=0,lte=100"`
}

// Re-export analytics types for convenience
type TradeoffInput = tradeoffDto.TradeoffInput
type TradeoffOutput = tradeoffDto.TradeoffOutput
