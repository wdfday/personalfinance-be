package debt_strategy

import (
	"go.uber.org/fx"
	"personalfinancedss/internal/module/analytics/debt_strategy/handler"
	"personalfinancedss/internal/module/analytics/debt_strategy/service"
)

var Module = fx.Module("debt_strategy",
	fx.Provide(service.NewService, handler.NewHandler),
)
