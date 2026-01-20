package budget_allocation

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/handler"
	"personalfinancedss/internal/module/analytics/budget_allocation/service"

	"go.uber.org/fx"
)

// Module exports budget allocation module for dependency injection
// Following MBMS pattern: Model -> Service -> Handler
// Note: Model is now provided by models.Module centrally
var Module = fx.Module("budget_allocation",
	fx.Provide(
		// Service (wraps model, adds logging)
		service.NewService,

		// Handler (HTTP layer)
		handler.NewHandler,
	),
)
