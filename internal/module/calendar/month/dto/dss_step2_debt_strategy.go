package dto

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	debtDto "personalfinancedss/internal/module/analytics/debt_strategy/dto"

	"github.com/google/uuid"
)

// ==================== Step 2: Debt Strategy ====================
// Debts are READ FROM REDIS CACHE (set during Initialize)

// PreviewDebtStrategyRequest requests debt strategy preview
// Debts are read from cached DSS state (initialized via POST /dss/initialize)
type PreviewDebtStrategyRequest struct {
	MonthID           uuid.UUID `json:"month_id" binding:"required"`
	PreferredStrategy string    `json:"preferred_strategy,omitempty"` // "avalanche" | "snowball" | "hybrid"
	// No debts or budget needed - read from Redis cache
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
