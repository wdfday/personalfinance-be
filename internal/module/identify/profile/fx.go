package profile

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/identify/profile/handler"
	"personalfinancedss/internal/module/identify/profile/repository"
	"personalfinancedss/internal/module/identify/profile/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides profile module dependencies.
var Module = fx.Module("profile",
	fx.Provide(
		fx.Annotate(
			repository.New,
			fx.As(new(repository.Repository)),
		),
		service.NewService,
		handler.NewHandler,
	),
	fx.Invoke(registerProfileRoutes),
)

func registerProfileRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
