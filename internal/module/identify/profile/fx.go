package profile

import (
	"personalfinancedss/internal/module/identify/profile/handler"
	"personalfinancedss/internal/module/identify/profile/repository"
	"personalfinancedss/internal/module/identify/profile/service"

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
)
