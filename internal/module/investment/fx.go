package investment

import (
	"personalfinancedss/internal/module/investment/investment_asset"
	"personalfinancedss/internal/module/investment/investment_transaction"
	"personalfinancedss/internal/module/investment/portfolio_snapshot"

	"go.uber.org/fx"
)

// Module provides all investment-related module dependencies
var Module = fx.Module("investment",
	// Include all sub-modules
	investment_asset.Module,
	investment_transaction.Module,
	portfolio_snapshot.Module,
)
