package transaction

import (
	budgetService "personalfinancedss/internal/module/cashflow/budget/service"
	debtService "personalfinancedss/internal/module/cashflow/debt/service"
	incomeProfileService "personalfinancedss/internal/module/cashflow/income_profile/service"
	"personalfinancedss/internal/module/cashflow/transaction/handler"
	"personalfinancedss/internal/module/cashflow/transaction/repository"
	"personalfinancedss/internal/module/cashflow/transaction/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides transaction module dependencies
var Module = fx.Module("transaction",
	fx.Provide(
		// Repository - provide as interface
		fx.Annotate(
			repository.NewGormRepository,
			fx.As(new(repository.Repository)),
		),

		// LinkProcessor - handles transaction link processing
		NewLinkProcessor,

		// Service - provide as interface
		fx.Annotate(
			service.NewService,
			fx.As(new(service.Service)),
		),

		// Handler
		handler.NewHandler,
	),
)

// NewLinkProcessor creates a new link processor with all required dependencies
func NewLinkProcessor(
	budgetSvc budgetService.Service,
	debtSvc debtService.Service,
	incomeProfileSvc incomeProfileService.Service,
	logger *zap.Logger,
) *service.LinkProcessor {
	return service.NewLinkProcessor(budgetSvc, debtSvc, incomeProfileSvc, logger)
}
