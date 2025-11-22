package investment_asset

import (
	"personalfinancedss/internal/module/investment/investment_asset/handler"
	"personalfinancedss/internal/module/investment/investment_asset/repository"
	"personalfinancedss/internal/module/investment/investment_asset/service"

	"go.uber.org/fx"
)

// Module provides investment asset module dependencies
var Module = fx.Module("investment_asset",
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
