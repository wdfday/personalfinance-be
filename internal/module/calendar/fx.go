package calendar

import (
	"personalfinancedss/internal/module/calendar/event"
	"personalfinancedss/internal/module/calendar/month"

	"go.uber.org/fx"
)

// Module provides calendar functionality (events and months).
var Module = fx.Module("calendar",
	event.Module,
	month.Module,
)
