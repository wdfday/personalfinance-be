package debt_strategy

import (
	"personalfinancedss/internal/module/analytics/debt_strategy/service"

	"go.uber.org/fx"
)

var Module = fx.Module("debt_strategy",
	fx.Provide(service.NewService),
)
