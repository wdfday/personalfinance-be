package emergency_fund

import (
	"go.uber.org/fx"
	"personalfinancedss/internal/module/analytics/emergency_fund/handler"
	"personalfinancedss/internal/module/analytics/emergency_fund/service"
)

var Module = fx.Module("emergency_fund",
	fx.Provide(service.NewService, handler.NewHandler),
)
