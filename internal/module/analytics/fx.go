package analytics

import (
	"personalfinancedss/internal/module/analytics/budget_allocation"
	budgetAllocationService "personalfinancedss/internal/module/analytics/budget_allocation/service"
	"personalfinancedss/internal/module/analytics/cashflow_forecast/service"
	"personalfinancedss/internal/module/analytics/debt_strategy"
	debtStrategyService "personalfinancedss/internal/module/analytics/debt_strategy/service"
	"personalfinancedss/internal/module/analytics/debt_tradeoff"
	debtTradeoffService "personalfinancedss/internal/module/analytics/debt_tradeoff/service"
	"personalfinancedss/internal/module/analytics/goal_prioritization"
	goalPrioritizationService "personalfinancedss/internal/module/analytics/goal_prioritization/service"
	"personalfinancedss/internal/module/analytics/models"

	"go.uber.org/fx"
)

// Module provides all analytics services
// It includes sub-modules that provide complete Model->Service->Handler chains
// and also provides named services for consumption by other modules (e.g., month)
var Module = fx.Module("analytics",
	// Include models module (provides all core models)
	models.Module,

	// Include sub-modules (they provide Service, Handler)
	budget_allocation.Module,
	goal_prioritization.Module,
	debt_strategy.Module,
	debt_tradeoff.Module,

	// Provide named services for month module consumption
	fx.Provide(
		fx.Annotate(
			func(svc budgetAllocationService.Service) budgetAllocationService.Service { return svc },
			fx.ResultTags(`name:"budgetAllocationService"`),
		),
		fx.Annotate(
			func(svc goalPrioritizationService.Service) goalPrioritizationService.Service { return svc },
			fx.ResultTags(`name:"goalPrioritizationService"`),
		),
		fx.Annotate(
			func(svc debtStrategyService.Service) debtStrategyService.Service { return svc },
			fx.ResultTags(`name:"debtStrategyService"`),
		),
		fx.Annotate(
			func(svc debtTradeoffService.Service) debtTradeoffService.Service { return svc },
			fx.ResultTags(`name:"debtTradeoffService"`),
		),
		// Cashflow forecast - standalone service (no model dependency)
		fx.Annotate(
			service.NewService,
			fx.ResultTags(`name:"cashflowForecastService"`),
		),
	),
)
