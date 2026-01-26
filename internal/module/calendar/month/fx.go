package month

import (
	"go.uber.org/fx"

	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/calendar/month/handler"
	"personalfinancedss/internal/module/calendar/month/repository"
	"personalfinancedss/internal/module/calendar/month/service"
)

// Module wires the calendar month feature
var Module = fx.Module("calendar-month",
	fx.Provide(
		fx.Annotate(
			repository.NewGormRepository,
			fx.As(new(repository.Repository)),
		),
		// Redis client for DSS caching
		config.NewRedisClient,
		// DSS Cache service
		service.NewDSSCache,
		fx.Annotate(
			service.NewService,
			fx.ParamTags(
				``,                                 // repo
				``,                                 // categoryService
				``,                                 // incomeProfileRepo
				``,                                 // budgetService
				`name:"goalPrioritizationService"`, // goalPrioritization
				`name:"debtStrategyService"`,       // debtStrategy
				`name:"debtTradeoffService"`,       // debtTradeoff
				`name:"budgetAllocationService"`,   // budgetAllocation
				`name:"cashflowForecastService"`,   // cashflowForecast
				``,                                 // dssCache
				``,                                 // logger
			),
			fx.As(new(service.Service)),
		),
		handler.NewHandler,
	),
	// Routes registered centrally in app.go RegisterRoutes
)
