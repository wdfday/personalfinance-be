package event

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/calendar/event/handler"
	"personalfinancedss/internal/module/calendar/event/repository"
	"personalfinancedss/internal/module/calendar/event/service"
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
	fx.Invoke(registerCalendarEventRoutes),
)

// registerCalendarEventRoutes registers the calendar event API routes
func registerCalendarEventRoutes(
	router *gin.Engine,
	h *handler.Handler,
	auth *middleware.Middleware,
) {
	h.RegisterRoutes(router, auth)
}
