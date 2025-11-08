package budget

import (
	"personalfinancedss/internal/module/cashflow/budget/handler"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"personalfinancedss/internal/module/cashflow/budget/service"

	"go.uber.org/fx"
)

// Module provides budget module dependencies
var Module = fx.Module("budget",
	fx.Provide(
		// Repository - provide as interface
		fx.Annotate(
			repository.New,
			fx.As(new(repository.Repository)),
		),

		// Service - provide as interface
		fx.Annotate(
			service.NewService,
			fx.As(new(service.Service)),
		),

		// Handler
		handler.NewHandler,
	),
)
