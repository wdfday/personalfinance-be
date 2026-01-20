package service

import (
	"context"

	"personalfinancedss/internal/module/analytics/cashflow_forecast/domain"
)

// Service defines the interface for cashflow forecast operations
type Service interface {
	// CalculateForecast computes income forecast and stability metrics
	CalculateForecast(ctx context.Context, input domain.ForecastInput) (*domain.ForecastResult, error)

	// RunStressTest executes Monte Carlo simulation for financial stress testing
	RunStressTest(ctx context.Context, input domain.StressTestInput) (*domain.StressTestResult, error)

	// CreateJobLossScenario creates a preset job loss stress test
	CreateJobLossScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error)

	// CreateMedicalShockScenario creates a preset medical emergency stress test
	CreateMedicalShockScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error)

	// CreateInflationScenario creates a preset inflation shock stress test
	CreateInflationScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error)

	// CreateFreelancerDrySpellScenario creates a preset freelancer dry spell test
	CreateFreelancerDrySpellScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64, dryMonths int) (*domain.StressTestInput, error)
}

// cashflowForecastService implements the Service interface
type cashflowForecastService struct{}

// NewService creates a new cashflow forecast service
func NewService() Service {
	return &cashflowForecastService{}
}

// CalculateForecast implements Service.CalculateForecast
func (s *cashflowForecastService) CalculateForecast(ctx context.Context, input domain.ForecastInput) (*domain.ForecastResult, error) {
	result := domain.CalculateForecast(input)
	return &result, nil
}

// RunStressTest implements Service.RunStressTest
func (s *cashflowForecastService) RunStressTest(ctx context.Context, input domain.StressTestInput) (*domain.StressTestResult, error) {
	result := domain.RunStressTest(input)
	return &result, nil
}

// CreateJobLossScenario implements Service.CreateJobLossScenario
func (s *cashflowForecastService) CreateJobLossScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error) {
	scenario := domain.CreateJobLossScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund)
	return &scenario, nil
}

// CreateMedicalShockScenario implements Service.CreateMedicalShockScenario
func (s *cashflowForecastService) CreateMedicalShockScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error) {
	scenario := domain.CreateMedicalShockScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund)
	return &scenario, nil
}

// CreateInflationScenario implements Service.CreateInflationScenario
func (s *cashflowForecastService) CreateInflationScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) (*domain.StressTestInput, error) {
	scenario := domain.CreateInflationScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund)
	return &scenario, nil
}

// CreateFreelancerDrySpellScenario implements Service.CreateFreelancerDrySpellScenario
func (s *cashflowForecastService) CreateFreelancerDrySpellScenario(ctx context.Context, currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64, dryMonths int) (*domain.StressTestInput, error) {
	scenario := domain.CreateFreelancerDrySpellScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund, dryMonths)
	return &scenario, nil
}
