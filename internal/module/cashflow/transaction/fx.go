package transaction

import (
	"personalfinancedss/internal/module/cashflow/transaction/handler"
	"personalfinancedss/internal/module/cashflow/transaction/repository"
	"personalfinancedss/internal/module/cashflow/transaction/service"

	"go.uber.org/fx"
)

// Module provides transaction module dependencies
var Module = fx.Module("transaction",
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
