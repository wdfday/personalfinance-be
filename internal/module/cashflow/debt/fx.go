package debt

import (
	"personalfinancedss/internal/module/cashflow/debt/handler"
	"personalfinancedss/internal/module/cashflow/debt/repository"
	"personalfinancedss/internal/module/cashflow/debt/service"

	"go.uber.org/fx"
)

// Module provides debt module dependencies
var Module = fx.Module("debt",
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
