package dto

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	debtDto "personalfinancedss/internal/module/analytics/debt_strategy/dto"

	"github.com/google/uuid"
)

// ==================== Step 2: Debt Strategy ====================
// All types imported from analytics debt_strategy module - NO DUPLICATION

// PreviewDebtStrategyRequest requests debt strategy preview
type PreviewDebtStrategyRequest struct {
	MonthID           uuid.UUID         `json:"month_id" binding:"required"`
	Debts             []domain.DebtInfo `json:"debts" binding:"required,min=1"`
	TotalDebtBudget   float64           `json:"total_debt_budget" binding:"required,gt=0"`
	PreferredStrategy string            `json:"preferred_strategy,omitempty"` // "avalanche" | "snowball" | "hybrid"
}

// PreviewDebtStrategyResponse type alias to analytics output
type PreviewDebtStrategyResponse = debtDto.DebtStrategyOutput

// ApplyDebtStrategyRequest applies user-selected strategy
type ApplyDebtStrategyRequest struct {
	MonthID          uuid.UUID `json:"month_id" binding:"required"`
	SelectedStrategy string    `json:"selected_strategy" binding:"required"`
}

// Re-export analytics types for convenience
type DebtInfo = domain.DebtInfo
type DebtStrategyInput = debtDto.DebtStrategyInput
type DebtStrategyOutput = debtDto.DebtStrategyOutput
