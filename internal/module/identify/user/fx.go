package user

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/identify/user/handler"
	"personalfinancedss/internal/module/identify/user/repository"
	"personalfinancedss/internal/module/identify/user/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides user module dependencies
var Module = fx.Module("user",
	fx.Provide(
		// Repository - provide as interface
		fx.Annotate(
			repository.New, // Using gorm repository
			fx.As(new(repository.Repository)),
		),

		// Service - provide as interface
		fx.Annotate(
			service.NewUserService,
			fx.As(new(service.IUserService)),
		),

		// Handler
		handler.NewHandler,
	),
	fx.Invoke(registerUserRoutes),
)

func registerUserRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
