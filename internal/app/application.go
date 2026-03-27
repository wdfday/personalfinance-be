package app

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/analytics"
	"personalfinancedss/internal/module/calendar"
	"personalfinancedss/internal/module/cashflow/account"
	"personalfinancedss/internal/module/cashflow/budget"
	"personalfinancedss/internal/module/cashflow/budget_profile"
	"personalfinancedss/internal/module/cashflow/category"
	"personalfinancedss/internal/module/cashflow/debt"
	"personalfinancedss/internal/module/cashflow/goal"
	"personalfinancedss/internal/module/cashflow/income_profile"
	"personalfinancedss/internal/module/cashflow/transaction"
	"personalfinancedss/internal/module/identify/auth"
	"personalfinancedss/internal/module/identify/broker"
	"personalfinancedss/internal/module/identify/profile"
	"personalfinancedss/internal/module/identify/user"
	"personalfinancedss/internal/module/notification"

	"go.uber.org/fx"
)

// Application creates the main FX application with all modules
func Application() *fx.App {
	options := []fx.Option{
		// Core modules
		CoreModule,

		// Infrastructure modules
		notification.Module,

		// Feature modules
		user.Module,
		profile.Module,
		auth.Module,
		broker.Module,
		account.Module,
		category.Module,
		transaction.Module,
		calendar.Module,
		budget.Module,
		income_profile.Module,
		budget_profile.Module,
		goal.Module,
		debt.Module,

		// Analytics module (new - contains all 7 modules for problems)
		analytics.Module,

		// App module (wires everything together)
		AppModule,
	}

	// Suppress FX logs in production for cleaner output
	if config.IsProduction() {
		options = append(options, fx.NopLogger)
	}

	return fx.New(options...)
}
