package domain

import (
	"math"

	"github.com/google/uuid"
)

// ===== CASHFLOW FORECAST DSS MODULE =====
// Based on docs/freelancer_financial_model.md and docs/rolling_cashflow_forecast.md
// Provides income calculation, stability analysis, and liquidity stress testing

// IncomeMode represents the user's income type
type IncomeMode string

const (
	IncomeModeFixed      IncomeMode = "fixed"      // Salaried employee (σ ≈ 0)
	IncomeModeFreelancer IncomeMode = "freelancer" // Variable income (high σ)
	IncomeModeMixed      IncomeMode = "mixed"      // Both fixed and variable
)

// ===== INPUT TYPES =====

// IncomeSource represents a single income stream
type IncomeSource struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	IsRecurring bool      `json:"is_recurring"`
	Frequency   string    `json:"frequency,omitempty"` // monthly, bi-weekly, weekly
	SourceType  string    `json:"source_type"`         // "salary", "freelance", "commission", "adhoc"
	IsConfirmed bool      `json:"is_confirmed"`        // Has transaction matched?
}

// FreelancerConfig contains settings for freelancer income smoothing
type FreelancerConfig struct {
	// Reservoir Model - "Salary to Self"
	ReservoirBalance float64 `json:"reservoir_balance"`  // Current balance in virtual reservoir
	SafeSalaryMonths int     `json:"safe_salary_months"` // How many months reservoir should cover (default: 6)
	SmoothingFactor  float64 `json:"smoothing_factor"`   // Multiplier for 12-month average (default: 0.8)

	// Variable Income Buffer
	VIBMultiplier float64 `json:"vib_multiplier"` // K factor for VIB calculation (default: 3)

	// Tax Vault (Vietnam: 10% PIT withholding)
	TaxWithholdingRate float64 `json:"tax_withholding_rate"` // Auto-reserve for taxes (default: 0.1)

	// Mandatory expenses not covered by clients (BHXH, BHYT)
	MandatoryExpenses float64 `json:"mandatory_expenses"`
}

// ForecastInput contains all data needed for cashflow forecast
type ForecastInput struct {
	Mode IncomeMode `json:"mode"`

	// Current month data
	RolloverTBB    float64        `json:"rollover_tbb"`    // TBB carried from previous month
	IncomeSources  []IncomeSource `json:"income_sources"`  // All income for this month
	MonthlyExpense float64        `json:"monthly_expense"` // Expected expenses

	// Historical data (last 12 months for variance calculation)
	HistoricalIncome []float64 `json:"historical_income,omitempty"`

	// Freelancer configuration
	FreelancerConfig *FreelancerConfig `json:"freelancer_config,omitempty"`
}

// ===== OUTPUT TYPES =====

// ForecastResult contains the calculated income and stability values
type ForecastResult struct {
	// Core values for budgeting
	TotalAvailable   float64 `json:"total_available"`    // What can be budgeted this month
	ProjectedIncome  float64 `json:"projected_income"`   // Raw expected income
	RolloverFromPrev float64 `json:"rollover_from_prev"` // TBB from previous month

	// Income breakdown
	FixedIncome    float64 `json:"fixed_income"`    // Stable recurring (salary)
	VariableIncome float64 `json:"variable_income"` // Freelance/commission
	AdHocIncome    float64 `json:"adhoc_income"`    // One-time expected

	// Freelancer mode outputs
	SafeSalary       float64 `json:"safe_salary,omitempty"`       // Calculated "salary to self"
	ReservoirDeposit float64 `json:"reservoir_deposit,omitempty"` // Excess going to reservoir
	TaxReserve       float64 `json:"tax_reserve,omitempty"`       // Reserved for taxes

	// Stability analysis
	Stability StabilityMetrics `json:"stability"`
}

// StabilityMetrics measures income predictability and risk
type StabilityMetrics struct {
	// Coefficient of Variation (CV = σ/μ) - lower = more stable
	CoefficientOfVariation float64 `json:"coefficient_of_variation"`

	// Stability score (0-100, higher = more stable)
	StabilityScore float64 `json:"stability_score"`

	// Risk assessment
	RiskLevel string `json:"risk_level"` // "low", "medium", "high", "critical"

	// Recommended Variable Income Buffer
	RecommendedVIB float64 `json:"recommended_vib"`

	// Liquidity Coverage Ratio (Basel III inspired)
	// LCR = Liquid Assets / Net Cash Outflow (3 months)
	LiquidityCoverage float64 `json:"liquidity_coverage"`

	// Months covered by current average income
	MonthsCovered float64 `json:"months_covered"`
}

// ===== CALCULATOR =====

// CalculateForecast computes income forecast based on mode and inputs
func CalculateForecast(input ForecastInput) ForecastResult {
	result := ForecastResult{
		RolloverFromPrev: input.RolloverTBB,
	}

	// Categorize and sum income sources
	for _, src := range input.IncomeSources {
		result.ProjectedIncome += src.Amount
		switch src.SourceType {
		case "salary":
			result.FixedIncome += src.Amount
		case "freelance", "commission":
			result.VariableIncome += src.Amount
		case "adhoc":
			result.AdHocIncome += src.Amount
		default:
			if src.IsRecurring {
				result.FixedIncome += src.Amount
			} else {
				result.VariableIncome += src.Amount
			}
		}
	}

	// Calculate stability metrics
	result.Stability = CalculateStability(input.HistoricalIncome, input.MonthlyExpense)

	// Apply mode-specific logic
	switch input.Mode {
	case IncomeModeFreelancer:
		result = applyFreelancerMode(input, result)
	case IncomeModeMixed:
		result = applyMixedMode(input, result)
	default: // Fixed mode - simple calculation
		result.TotalAvailable = result.ProjectedIncome + input.RolloverTBB
	}

	return result
}

// applyFreelancerMode implements the Reservoir Model for income smoothing
func applyFreelancerMode(input ForecastInput, result ForecastResult) ForecastResult {
	config := input.FreelancerConfig
	if config == nil {
		config = DefaultFreelancerConfig()
	}

	// Calculate 12-month average income
	avgIncome := calculateAverage(input.HistoricalIncome)
	if avgIncome == 0 {
		avgIncome = result.ProjectedIncome // Fallback if no history
	}

	// === INCOME SMOOTHING: "Salary to Self" ===
	// S_safe = min(AverageIncome_12m × 0.8, ReservoirBalance / 6)
	salaryFromAverage := avgIncome * config.SmoothingFactor
	salaryFromReservoir := float64(0)
	if config.SafeSalaryMonths > 0 {
		salaryFromReservoir = config.ReservoirBalance / float64(config.SafeSalaryMonths)
	}

	if salaryFromReservoir > 0 {
		result.SafeSalary = math.Min(salaryFromAverage, salaryFromReservoir)
	} else {
		result.SafeSalary = salaryFromAverage
	}

	// Ensure safe salary covers mandatory expenses (BHXH/BHYT)
	if result.SafeSalary < config.MandatoryExpenses {
		result.SafeSalary = config.MandatoryExpenses
	}

	// === TAX RESERVE (Tax Vault) ===
	// VN: Auto-reserve 10% for PIT withholding
	result.TaxReserve = result.ProjectedIncome * config.TaxWithholdingRate

	// === RESERVOIR DEPOSIT ===
	// Excess income goes to reservoir for future months
	incomeAfterTax := result.ProjectedIncome - result.TaxReserve
	if incomeAfterTax > result.SafeSalary {
		result.ReservoirDeposit = incomeAfterTax - result.SafeSalary
	}

	// === TOTAL AVAILABLE ===
	// Freelancer budgets from SafeSalary + Rollover, NOT raw income
	result.TotalAvailable = result.SafeSalary + input.RolloverTBB

	return result
}

// applyMixedMode handles users with both stable salary and variable income
func applyMixedMode(input ForecastInput, result ForecastResult) ForecastResult {
	config := input.FreelancerConfig
	if config == nil {
		config = DefaultFreelancerConfig()
	}

	// Fixed income is available immediately
	// Variable income goes through smoothing

	// Extract variable income history (total - fixed)
	variableHistory := make([]float64, len(input.HistoricalIncome))
	for i, total := range input.HistoricalIncome {
		variable := total - result.FixedIncome
		if variable < 0 {
			variable = 0
		}
		variableHistory[i] = variable
	}

	// Apply smoothing to variable portion only
	avgVariable := calculateAverage(variableHistory)
	if avgVariable == 0 {
		avgVariable = result.VariableIncome
	}

	variableSafeSalary := avgVariable * config.SmoothingFactor

	// Tax reserve on variable income only
	result.TaxReserve = result.VariableIncome * config.TaxWithholdingRate

	// Reservoir deposit from variable excess
	variableAfterTax := result.VariableIncome - result.TaxReserve
	if variableAfterTax > variableSafeSalary {
		result.ReservoirDeposit = variableAfterTax - variableSafeSalary
	}

	result.SafeSalary = variableSafeSalary

	// Total = Fixed + Smoothed Variable + Rollover
	result.TotalAvailable = result.FixedIncome + variableSafeSalary + input.RolloverTBB

	return result
}

// CalculateStability computes income stability metrics
func CalculateStability(historicalIncome []float64, monthlyExpense float64) StabilityMetrics {
	metrics := StabilityMetrics{
		RiskLevel:      "unknown",
		StabilityScore: 50, // Neutral when unknown
	}

	if len(historicalIncome) < 3 {
		return metrics // Not enough data
	}

	// Calculate mean and standard deviation
	mean := calculateAverage(historicalIncome)
	stdDev := calculateStdDev(historicalIncome, mean)

	// Coefficient of Variation (CV = σ/μ)
	if mean > 0 {
		metrics.CoefficientOfVariation = stdDev / mean
	}

	// Stability Score: 100 - (CV × 100), capped 0-100
	metrics.StabilityScore = math.Max(0, math.Min(100, 100-(metrics.CoefficientOfVariation*100)))

	// Risk Level based on CV thresholds
	switch {
	case metrics.CoefficientOfVariation < 0.1:
		metrics.RiskLevel = "low" // Salaried, very stable
	case metrics.CoefficientOfVariation < 0.3:
		metrics.RiskLevel = "medium" // Some variation
	case metrics.CoefficientOfVariation < 0.5:
		metrics.RiskLevel = "high" // Significant variation (typical freelancer)
	default:
		metrics.RiskLevel = "critical" // Extreme variation
	}

	// Recommended Variable Income Buffer
	// VIB = μ_expenses × CV × K where K = 3
	K := 3.0
	metrics.RecommendedVIB = monthlyExpense * metrics.CoefficientOfVariation * K

	// Liquidity Coverage Ratio
	// LCR = Average Income / (3 months of expenses)
	if monthlyExpense > 0 {
		metrics.LiquidityCoverage = mean / (monthlyExpense * 3)
		metrics.MonthsCovered = mean / monthlyExpense
	}

	return metrics
}

// ===== HELPERS =====

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

// DefaultFreelancerConfig returns sensible defaults for Vietnam context
func DefaultFreelancerConfig() *FreelancerConfig {
	return &FreelancerConfig{
		SafeSalaryMonths:   6,
		SmoothingFactor:    0.8,
		VIBMultiplier:      3,
		TaxWithholdingRate: 0.1, // 10% PIT withholding in VN
		MandatoryExpenses:  0,   // User should set BHXH/BHYT
	}
}
