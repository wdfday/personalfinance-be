package account

import (
	"personalfinancedss/internal/module/cashflow/account/handler"
	"personalfinancedss/internal/module/cashflow/account/repository"
	"personalfinancedss/internal/module/cashflow/account/service"

	"go.uber.org/fx"
)

// Module provides account module dependencies.
var Module = fx.Module("account",
	fx.Provide(
		// Repository - provide as interface (không dùng name; accountRepo vs transactionRepo khác type nên Fx phân biệt được)
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
