package goal

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/goal/handler"
	"personalfinancedss/internal/module/cashflow/goal/repository"
	"personalfinancedss/internal/module/cashflow/goal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides goal module dependencies
var Module = fx.Module("goal",
	fx.Provide(
		// Repository - provide as interface
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
	fx.Invoke(registerGoalRoutes),
)

func registerGoalRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
