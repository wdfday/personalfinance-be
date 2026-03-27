package event

import (
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/calendar/event/handler"
	"personalfinancedss/internal/module/calendar/event/repository"
	"personalfinancedss/internal/module/calendar/event/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module wires the calendar event feature.
var Module = fx.Module("calendar-event",
	fx.Provide(
		fx.Annotate(
			repository.NewGormRepository,
			fx.As(new(repository.Repository)),
		),
		fx.Annotate(
			service.NewService,
			fx.As(new(service.Service)),
		),
		handler.NewHandler,
	),
	fx.Invoke(registerEventRoutes),
)

func registerEventRoutes(router *gin.Engine, h *handler.Handler, authMiddleware *middleware.Middleware) {
	h.RegisterRoutes(router, authMiddleware)
}
