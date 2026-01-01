package calendar

import (
	"personalfinancedss/internal/module/calendar/event"

	"go.uber.org/fx"
)

// Module provides calendar event functionality.
var Module = fx.Module("calendar",
	event.Module,
)
