package models

import (
	"personalfinancedss/internal/module/analytics/models/ahp"
	"personalfinancedss/internal/module/analytics/models/budget_allocation"
	"personalfinancedss/internal/module/analytics/models/debt_strategy"
	"personalfinancedss/internal/module/analytics/models/emergency_fund"
	"personalfinancedss/internal/module/analytics/models/lifestyle_inflation"
	"personalfinancedss/internal/module/analytics/models/spending_decision"
	"personalfinancedss/internal/module/analytics/models/tradeoff"

	"go.uber.org/fx"
)

// Module provides all analytics models for dependency injection
// These are the core mathematical/algorithmic models used by analytics services
var Module = fx.Module("analytics-models",
	fx.Provide(
		// Budget Allocation Model (Goal Programming solver)
		budget_allocation.NewBudgetAllocationModel,

		// AHP Model (Analytic Hierarchy Process for goal prioritization)
		ahp.NewAHPModel,

		// Debt Strategy Model (Avalanche/Snowball strategies)
		debt_strategy.NewDebtStrategyModel,

		// Tradeoff Model (Goal vs Debt vs Investment tradeoffs)
		tradeoff.NewTradeoffModel,

		// Spending Decision Model (Large purchase analysis)
		spending_decision.NewSpendingDecisionModel,

		// Emergency Fund Model
		emergency_fund.NewEmergency_fundModel,

		// Lifestyle Inflation Model
		lifestyle_inflation.NewLifestyle_inflationModel,
	),
)
