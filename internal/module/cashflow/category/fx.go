package category

import (
	"personalfinancedss/internal/module/cashflow/category/handler"
	"personalfinancedss/internal/module/cashflow/category/repository"
	"personalfinancedss/internal/module/cashflow/category/service"

	"go.uber.org/fx"
)

// Module provides category module dependencies
var Module = fx.Module("category",
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
