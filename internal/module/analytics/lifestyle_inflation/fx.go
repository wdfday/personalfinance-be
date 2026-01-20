package lifestyle_inflation

import (
	"go.uber.org/fx"
	"personalfinancedss/internal/module/analytics/lifestyle_inflation/handler"
	"personalfinancedss/internal/module/analytics/lifestyle_inflation/service"
)

var Module = fx.Module("lifestyle_inflation",
	fx.Provide(service.NewService, handler.NewHandler),
)
