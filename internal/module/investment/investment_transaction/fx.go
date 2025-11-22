package investment_transaction

import (
	"personalfinancedss/internal/module/investment/investment_transaction/handler"
	"personalfinancedss/internal/module/investment/investment_transaction/repository"
	"personalfinancedss/internal/module/investment/investment_transaction/service"

	"go.uber.org/fx"
)

// Module provides investment transaction module dependencies
var Module = fx.Module("investment_transaction",
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
