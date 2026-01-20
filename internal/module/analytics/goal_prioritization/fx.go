package goal_prioritization

import (
	"personalfinancedss/internal/module/analytics/goal_prioritization/handler"
	"personalfinancedss/internal/module/analytics/goal_prioritization/service"

	"go.uber.org/fx"
)

// Module exports AHP module for dependency injection
var Module = fx.Module("goal_prioritization",
	fx.Provide(
		service.NewService,
		handler.NewHandler,
	),
)
