package budget_profile

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/budget_profile/handler"
	"personalfinancedss/internal/module/cashflow/budget_profile/repository"
	"personalfinancedss/internal/module/cashflow/budget_profile/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides budget profile module dependencies
var Module = fx.Module("budget_profile",
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
	fx.Invoke(registerBudgetProfileRoutes),
)

func registerBudgetProfileRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
