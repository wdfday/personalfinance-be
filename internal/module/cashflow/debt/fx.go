package debt

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/debt/handler"
	"personalfinancedss/internal/module/cashflow/debt/repository"
	"personalfinancedss/internal/module/cashflow/debt/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides debt module dependencies
var Module = fx.Module("debt",
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
	fx.Invoke(registerDebtRoutes),
)

func registerDebtRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
