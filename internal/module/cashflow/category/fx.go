package category

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/category/handler"
	"personalfinancedss/internal/module/cashflow/category/repository"
	"personalfinancedss/internal/module/cashflow/category/service"

	"github.com/gin-gonic/gin"
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
	fx.Invoke(registerCategoryRoutes),
)

func registerCategoryRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
