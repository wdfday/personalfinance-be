package fx

import (
	"personalfinancedss/internal/config"
	// Old analytic modules - commented out, replaced by new analytics module
	// "personalfinancedss/internal/module/analytic/budget_optimizer"
	// "personalfinancedss/internal/module/analytic/cashflow_projection"
	// "personalfinancedss/internal/module/analytic/debt_strategy"
	// "personalfinancedss/internal/module/analytic/dss"
	// "personalfinancedss/internal/module/analytic/goal_prioritization"
	// "personalfinancedss/internal/module/analytic/recommendation"
	// "personalfinancedss/internal/module/analytic/tradeoff_analysis"
	// New analytics module
	"personalfinancedss/internal/module/analytics"
	"personalfinancedss/internal/module/calendar"
	// "personalfinancedss/internal/module/calendar"
	"personalfinancedss/internal/module/cashflow/account"
	"personalfinancedss/internal/module/cashflow/budget"
	"personalfinancedss/internal/module/cashflow/budget_profile"
	"personalfinancedss/internal/module/cashflow/category"
	"personalfinancedss/internal/module/cashflow/debt"
	"personalfinancedss/internal/module/cashflow/goal"
	"personalfinancedss/internal/module/cashflow/income_profile"
	"personalfinancedss/internal/module/cashflow/transaction"
	"personalfinancedss/internal/module/chatbot"
	"personalfinancedss/internal/module/identify/auth"
	"personalfinancedss/internal/module/identify/broker"
	"personalfinancedss/internal/module/identify/profile"
	"personalfinancedss/internal/module/identify/user"
	"personalfinancedss/internal/module/investment"
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
		investment.Module,

		// Analytics module (new - contains all 7 DSS problems)
		analytics.Module,

		// AI Chatbot module
		chatbot.Module,

		// App module (wires everything together)
		AppModule,
	}

	// Suppress FX logs in production for cleaner output
	if config.IsProduction() {
		options = append(options, fx.NopLogger)
	}

	return fx.New(options...)
}
