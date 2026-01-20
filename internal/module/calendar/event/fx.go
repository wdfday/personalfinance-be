package event

import (
	"go.uber.org/fx"

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
	// Routes registered centrally in app.go RegisterRoutes
)
