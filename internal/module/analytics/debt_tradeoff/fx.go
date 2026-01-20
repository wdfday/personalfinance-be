package debt_tradeoff

import (
	"personalfinancedss/internal/module/analytics/debt_tradeoff/handler"
	"personalfinancedss/internal/module/analytics/debt_tradeoff/service"

	"go.uber.org/fx"
)

var Module = fx.Module("debt_tradeoff",
	fx.Provide(
		service.NewService,
		handler.NewHandler,
	),
)
