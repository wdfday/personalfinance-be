package engine

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
)

// MonteCarloSimulator performs Monte Carlo simulations for tradeoff analysis
type MonteCarloSimulator struct {
	rng        *rand.Rand
	calculator *FinancialCalculator
}

// NewMonteCarloSimulator creates a new simulator
func NewMonteCarloSimulator() *MonteCarloSimulator {
	return &MonteCarloSimulator{
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		calculator: NewFinancialCalculator(),
	}
}

// SimulationInput contains all data needed for simulation
type SimulationInput struct {
	Debts             []domain.DebtInfo
	MonthlyIncome     float64
	EssentialExpenses float64
	TotalMinPayments  float64
	Ratio             domain.AllocationRatio
	ExpectedReturn    float64
	InitialSavings    float64
	Goals             []domain.GoalInfo
	Config            domain.SimulationConfig
}

// SingleSimulationResult result from one simulation run
type SingleSimulationResult struct {
	MonthsToDebtFree int
	FinalNPV         float64
	GoalsAchieved    int
	TotalGoals       int
}

// RunSimulation runs Monte Carlo simulation and returns aggregated results
func (s *MonteCarloSimulator) RunSimulation(input SimulationInput) *domain.MonteCarloResult {
	results := make([]SingleSimulationResult, input.Config.NumSimulations)

	for i := 0; i < input.Config.NumSimulations; i++ {
		results[i] = s.runSingleSimulation(input)
	}

	return s.aggregateResults(results, input.Config)
}

// runSingleSimulation runs one simulation with randomized parameters
func (s *MonteCarloSimulator) runSingleSimulation(input SimulationInput) SingleSimulationResult {
	// Randomize income (±variance)
	incomeMultiplier := 1.0 + s.randomVariance(input.Config.IncomeVariance)
	monthlyIncome := input.MonthlyIncome * incomeMultiplier

	// Randomize expenses (±variance)
	expenseMultiplier := 1.0 + s.randomVariance(input.Config.ExpenseVariance)
	essentialExpenses := input.EssentialExpenses * expenseMultiplier

	// Randomize investment return (±variance)
	returnMultiplier := 1.0 + s.randomVariance(input.Config.ReturnVariance)
	expectedReturn := input.ExpectedReturn * returnMultiplier
	if expectedReturn < 0 {
		expectedReturn = 0
	}

	// Calculate extra money with randomized values
	extraMoney := monthlyIncome - essentialExpenses - input.TotalMinPayments
	if extraMoney < 0 {
		extraMoney = 0
	}

	debtPayment := extraMoney * input.Ratio.DebtPercent
	savingsAmount := extraMoney * input.Ratio.SavingsPercent

	// Simulate debt payoff
	monthsToDebtFree, _, _ := s.calculator.SimulateDebtPayoff(
		input.Debts,
		debtPayment,
		input.Config.ProjectionMonths,
	)

	// Simulate investment growth
	investmentValue := s.calculator.SimulateInvestmentGrowth(
		input.InitialSavings,
		savingsAmount,
		expectedReturn,
		input.Config.ProjectionMonths,
	)

	// Calculate NPV
	monthlyDiscountRate := input.Config.DiscountRate / 12
	npv := s.calculator.PresentValue(investmentValue, monthlyDiscountRate, input.Config.ProjectionMonths)

	// Count goals achieved
	goalsAchieved := 0
	for _, goal := range input.Goals {
		months := s.calculator.CalculateGoalMonths(
			goal.CurrentAmount,
			goal.TargetAmount,
			savingsAmount*goal.Priority, // Weighted by priority
			expectedReturn,
		)
		if months <= input.Config.ProjectionMonths {
			goalsAchieved++
		}
	}

	return SingleSimulationResult{
		MonthsToDebtFree: monthsToDebtFree,
		FinalNPV:         npv,
		GoalsAchieved:    goalsAchieved,
		TotalGoals:       len(input.Goals),
	}
}

// randomVariance returns a random value between -variance and +variance
func (s *MonteCarloSimulator) randomVariance(variance float64) float64 {
	return (s.rng.Float64()*2 - 1) * variance
}

// aggregateResults aggregates simulation results into statistics
func (s *MonteCarloSimulator) aggregateResults(
	results []SingleSimulationResult,
	config domain.SimulationConfig,
) *domain.MonteCarloResult {
	n := len(results)
	if n == 0 {
		return &domain.MonteCarloResult{}
	}

	// Extract arrays for percentile calculations
	debtFreeMonths := make([]int, n)
	npvs := make([]float64, n)
	successCount := 0

	for i, r := range results {
		debtFreeMonths[i] = r.MonthsToDebtFree
		npvs[i] = r.FinalNPV

		// Success = debt-free within projection period AND at least half goals achieved
		if r.MonthsToDebtFree <= config.ProjectionMonths {
			if r.TotalGoals == 0 || r.GoalsAchieved >= r.TotalGoals/2 {
				successCount++
			}
		}
	}

	// Sort for percentile calculations
	sort.Ints(debtFreeMonths)
	sort.Float64s(npvs)

	// Calculate statistics
	npvMean := s.mean(npvs)
	npvStdDev := s.stdDev(npvs, npvMean)

	return &domain.MonteCarloResult{
		NumSimulations:     n,
		SuccessProbability: float64(successCount) / float64(n),
		DebtFreeP50:        debtFreeMonths[n*50/100],
		DebtFreeP75:        debtFreeMonths[n*75/100],
		DebtFreeP90:        debtFreeMonths[n*90/100],
		NPVMean:            npvMean,
		NPVStdDev:          npvStdDev,
		NPVP5:              npvs[n*5/100],
		NPVP95:             npvs[n*95/100],
		ConfidenceInterval95: [2]float64{
			npvMean - 1.96*npvStdDev/math.Sqrt(float64(n)),
			npvMean + 1.96*npvStdDev/math.Sqrt(float64(n)),
		},
	}
}

// mean calculates arithmetic mean
func (s *MonteCarloSimulator) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stdDev calculates standard deviation
func (s *MonteCarloSimulator) stdDev(values []float64, mean float64) float64 {
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
