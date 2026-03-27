package month

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/calendar/month/handler"
	"personalfinancedss/internal/module/calendar/month/repository"
	"personalfinancedss/internal/module/calendar/month/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
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
	fx.Invoke(registerMonthRoutes),
)

func registerMonthRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
