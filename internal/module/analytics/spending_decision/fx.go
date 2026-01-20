package spending_decision

import (
	"go.uber.org/fx"
	"personalfinancedss/internal/module/analytics/spending_decision/handler"
	"personalfinancedss/internal/module/analytics/spending_decision/service"
)

var Module = fx.Module("spending_decision",
	fx.Provide(service.NewService, handler.NewHandler),
)
