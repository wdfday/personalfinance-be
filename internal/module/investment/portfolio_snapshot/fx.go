package portfolio_snapshot

import (
	"personalfinancedss/internal/module/investment/portfolio_snapshot/handler"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/repository"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/service"

	"go.uber.org/fx"
)

// Module provides portfolio snapshot module dependencies
var Module = fx.Module("portfolio_snapshot",
	fx.Provide(
		// Repository - provide as interface
		fx.Annotate(
			repository.NewGormRepository,
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
