package account

import (
	"personalfinancedss/internal/broker/client"
	"personalfinancedss/internal/broker/client/okx"
	"personalfinancedss/internal/broker/client/ssi"
	"personalfinancedss/internal/module/cashflow/account/handler"
	"personalfinancedss/internal/module/cashflow/account/repository"
	"personalfinancedss/internal/module/cashflow/account/service"

	"go.uber.org/fx"
)

// provideSSIClient provides SSI broker client
func provideSSIClient() client.BrokerClient {
	return ssi.NewSSIClient()
}

// provideOKXClient provides OKX broker client
func provideOKXClient() client.BrokerClient {
	return okx.NewOKXClient()
}

// Module provides account module dependencies.
var Module = fx.Module("account",
	fx.Provide(
		// Repository - provide as interface
		fx.Annotate(
			repository.New,
			fx.As(new(repository.Repository)),
		),

		// Broker clients
		fx.Annotate(
			provideSSIClient,
			fx.ResultTags(`name:"ssiClient"`),
		),
		fx.Annotate(
			provideOKXClient,
			fx.ResultTags(`name:"okxClient"`),
		),

		// Service - provide as interface with broker clients
		fx.Annotate(
			service.NewService,
			fx.ParamTags(``, `name:"ssiClient"`, `name:"okxClient"`),
			fx.As(new(service.Service)),
		),

		// Handler
		handler.NewHandler,
	),
)
