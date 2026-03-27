package budget

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/budget/handler"
	"personalfinancedss/internal/module/cashflow/budget/repository"
	"personalfinancedss/internal/module/cashflow/budget/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides budget module dependencies
var Module = fx.Module("budget",
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
	fx.Invoke(registerBudgetRoutes),
)

func registerBudgetRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
