package income_profile

import (
	"personalfinancedss/internal/module/cashflow/income_profile/handler"
	"personalfinancedss/internal/module/cashflow/income_profile/repository"
	"personalfinancedss/internal/module/cashflow/income_profile/service"

	"go.uber.org/fx"
)

// Module provides income profile module dependencies
var Module = fx.Module("income_profile",
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
