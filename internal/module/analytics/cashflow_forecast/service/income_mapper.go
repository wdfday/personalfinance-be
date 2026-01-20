package service

import (
	cashflowForecastDomain "personalfinancedss/internal/module/analytics/cashflow_forecast/domain"
	incomeProfileDomain "personalfinancedss/internal/module/cashflow/income_profile/domain"
)

// IncomeForecastMapper maps income profiles to cashflow forecast inputs
type IncomeForecastMapper struct{}

// NewIncomeForecastMapper creates a new mapper
func NewIncomeForecastMapper() *IncomeForecastMapper {
	return &IncomeForecastMapper{}
}

// MapToForecastInput converts income profiles to cashflow forecast input
// Returns ForecastInput ready for CalculateForecast()
func (m *IncomeForecastMapper) MapToForecastInput(
	activeProfiles []*incomeProfileDomain.IncomeProfile,
	historicalMonthlyIncome []float64,
	rolloverTBB float64,
	monthlyExpense float64,
) cashflowForecastDomain.ForecastInput {

	// Detect income mode from profile characteristics
	mode := m.DetectIncomeMode(activeProfiles)

	// Convert profiles to income sources
	sources := m.ConvertToIncomeSources(activeProfiles)

	// Build forecast input
	input := cashflowForecastDomain.ForecastInput{
		Mode:             mode,
		RolloverTBB:      rolloverTBB,
		IncomeSources:    sources,
		MonthlyExpense:   monthlyExpense,
		HistoricalIncome: historicalMonthlyIncome,
	}

	// Add freelancer config if applicable
	if mode == cashflowForecastDomain.IncomeModeFreelancer || mode == cashflowForecastDomain.IncomeModeMixed {
		input.FreelancerConfig = m.BuildFreelancerConfig(activeProfiles, monthlyExpense)
	}

	return input
}

// DetectIncomeMode determines income mode from profile characteristics
func (m *IncomeForecastMapper) DetectIncomeMode(profiles []*incomeProfileDomain.IncomeProfile) cashflowForecastDomain.IncomeMode {
	if len(profiles) == 0 {
		return cashflowForecastDomain.IncomeModeFixed // Default
	}

	hasFixedIncome := false
	hasVariableIncome := false

	for _, p := range profiles {
		if !p.IsActive() {
			continue
		}

		// Fixed: Salary with base salary component
		if p.BaseSalary > 0 && p.IsRecurring && p.Frequency == "monthly" {
			hasFixedIncome = true
		}

		// Variable: Commission, freelance, or one-time
		if p.Commission > 0 || p.Source == "Freelance" || !p.IsRecurring {
			hasVariableIncome = true
		}

		// Check source keywords
		switch {
		case contains(p.Source, []string{"Freelance", "Contractor", "Gig"}):
			hasVariableIncome = true
		case contains(p.Source, []string{"Salary", "Wage"}):
			hasFixedIncome = true
		}
	}

	// Determine mode
	if hasFixedIncome && hasVariableIncome {
		return cashflowForecastDomain.IncomeModeMixed
	} else if hasVariableIncome {
		return cashflowForecastDomain.IncomeModeFreelancer
	}
	return cashflowForecastDomain.IncomeModeFixed
}

// ConvertToIncomeSources converts income profiles to forecast income sources
func (m *IncomeForecastMapper) ConvertToIncomeSources(profiles []*incomeProfileDomain.IncomeProfile) []cashflowForecastDomain.IncomeSource {
	var sources []cashflowForecastDomain.IncomeSource

	for _, p := range profiles {
		if !p.IsActive() {
			continue
		}

		// Determine source type
		sourceType := m.DetermineSourceType(p)

		source := cashflowForecastDomain.IncomeSource{
			ID:          p.ID,
			Name:        p.Source,
			Amount:      p.Amount,
			IsRecurring: p.IsRecurring,
			Frequency:   p.Frequency,
			SourceType:  sourceType,
			IsConfirmed: p.IsVerified,
		}

		sources = append(sources, source)
	}

	return sources
}

// DetermineSourceType maps income profile to forecast source type
func (m *IncomeForecastMapper) DetermineSourceType(profile *incomeProfileDomain.IncomeProfile) string {
	// Check explicit components
	if profile.BaseSalary > 0 && profile.Commission == 0 {
		return "salary"
	}
	if profile.Commission > 0 {
		return "commission"
	}

	// Check source keywords
	switch {
	case contains(profile.Source, []string{"Freelance", "Contractor", "Gig", "Project"}):
		return "freelance"
	case contains(profile.Source, []string{"Salary", "Wage"}):
		return "salary"
	case !profile.IsRecurring:
		return "adhoc"
	default:
		// Default: recurring = salary, non-recurring = freelance
		if profile.IsRecurring {
			return "salary"
		}
		return "freelance"
	}
}

// BuildFreelancerConfig creates freelancer config from profiles
func (m *IncomeForecastMapper) BuildFreelancerConfig(
	profiles []*incomeProfileDomain.IncomeProfile,
	monthlyExpense float64,
) *cashflowForecastDomain.FreelancerConfig {

	// Calculate mandatory expenses (BHXH/BHYT for Vietnam)
	// Typically ~10.5% of base salary for self-employed
	var mandatoryExpenses float64
	for _, p := range profiles {
		if p.BaseSalary > 0 {
			mandatoryExpenses += p.BaseSalary * 0.105 // 10.5% social insurance
		}
	}

	return &cashflowForecastDomain.FreelancerConfig{
		ReservoirBalance:   0, // TODO: Track in user settings
		SafeSalaryMonths:   6,
		SmoothingFactor:    0.8,
		VIBMultiplier:      3.0,
		TaxWithholdingRate: 0.1, // 10% PIT in Vietnam
		MandatoryExpenses:  mandatoryExpenses,
	}
}

// Helper: contains checks if any keyword is in the source string
func contains(source string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(source) >= len(keyword) {
			// Simple substring check (case-insensitive would be better)
			for i := 0; i <= len(source)-len(keyword); i++ {
				if source[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}
