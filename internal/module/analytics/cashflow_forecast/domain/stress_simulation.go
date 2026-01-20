package domain

import (
	"math"
	"math/rand"
	"time"
)

// ===== STRESS TEST / LIQUIDITY SIMULATION =====
// Based on docs/rolling_cashflow_forecast.md
// Monte Carlo simulation for cashflow stress testing

// StressScenario represents a type of financial stress event
type StressScenario string

const (
	ScenarioJobLoss        StressScenario = "job_loss"        // Complete income loss
	ScenarioMedicalShock   StressScenario = "medical_shock"   // Large unexpected expense
	ScenarioInflationSpike StressScenario = "inflation_spike" // Permanent expense increase
	ScenarioDrySpell       StressScenario = "dry_spell"       // No freelance income for N months
	ScenarioCustom         StressScenario = "custom"
)

// StressTestInput contains parameters for running stress simulation
type StressTestInput struct {
	// Current financial state
	CurrentCash     float64 `json:"current_cash"`     // Available liquid assets
	MonthlyIncome   float64 `json:"monthly_income"`   // Expected monthly income
	MonthlyExpenses float64 `json:"monthly_expenses"` // Base monthly expenses
	EmergencyFund   float64 `json:"emergency_fund"`   // Separate emergency savings

	// Stress parameters
	Scenario        StressScenario `json:"scenario"`
	MonthsToProject int            `json:"months_to_project"` // Default: 12
	Simulations     int            `json:"simulations"`       // Default: 1000

	// Scenario-specific parameters
	JobLossMonth   int     `json:"job_loss_month,omitempty"`   // Month when job loss occurs (0 = immediate)
	MedicalCost    float64 `json:"medical_cost,omitempty"`     // Medical shock cost (default: 6x income)
	InflationRate  float64 `json:"inflation_rate,omitempty"`   // Inflation spike (default: 0.20 = 20%)
	DrySpellMonths int     `json:"dry_spell_months,omitempty"` // Freelancer dry spell duration

	// Income volatility (for Monte Carlo)
	IncomeVolatility float64 `json:"income_volatility,omitempty"` // CV of income

	// Optional: Planned large expenses
	PlannedEvents []PlannedEvent `json:"planned_events,omitempty"`
}

// PlannedEvent represents a scheduled large expense
type PlannedEvent struct {
	Month  int     `json:"month"`  // Which month (1-indexed)
	Amount float64 `json:"amount"` // Cost
	Name   string  `json:"name"`   // Description
}

// StressTestResult contains the stress test analysis
type StressTestResult struct {
	// Core risk metrics
	InsolvencyProbability float64 `json:"insolvency_probability"` // P(balance < 0)
	RiskLevel             string  `json:"risk_level"`             // "safe", "warning", "danger"

	// Timing analysis
	FirstInsolvencyMonth  int     `json:"first_insolvency_month,omitempty"` // When first insolvency occurs
	AverageMonthsSurvived float64 `json:"avg_months_survived"`              // Before running out
	MedianLowestBalance   float64 `json:"median_lowest_balance"`            // Typical lowest point

	// Danger zones
	DangerZones []DangerZone `json:"danger_zones,omitempty"`

	// Liquidity Coverage Ratio (Basel III)
	LCR float64 `json:"lcr"` // Liquid Assets / 3-month net outflow

	// Recommendations
	Recommendations []string `json:"recommendations"`

	// Cash path percentiles (for visualization)
	Percentile5  []float64 `json:"percentile_5,omitempty"`  // Worst case
	Percentile50 []float64 `json:"percentile_50,omitempty"` // Median
	Percentile95 []float64 `json:"percentile_95,omitempty"` // Best case
}

// DangerZone identifies high-risk periods
type DangerZone struct {
	Month       int     `json:"month"`
	Probability float64 `json:"probability"` // Risk of insolvency this month
	Reason      string  `json:"reason"`      // Why this month is risky
}

// RunStressTest executes Monte Carlo simulation for the given scenario
func RunStressTest(input StressTestInput) StressTestResult {
	// Set defaults
	if input.MonthsToProject == 0 {
		input.MonthsToProject = 12
	}
	if input.Simulations == 0 {
		input.Simulations = 1000
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Collect simulation results
	allPaths := make([][]float64, input.Simulations)
	insolvencyCount := 0
	monthsSurvived := make([]float64, input.Simulations)
	lowestBalances := make([]float64, input.Simulations)

	for sim := 0; sim < input.Simulations; sim++ {
		path, survived, lowest := runSingleSimulation(input, rng)
		allPaths[sim] = path
		monthsSurvived[sim] = float64(survived)
		lowestBalances[sim] = lowest

		if survived < input.MonthsToProject {
			insolvencyCount++
		}
	}

	// Analyze results
	result := StressTestResult{
		InsolvencyProbability: float64(insolvencyCount) / float64(input.Simulations),
		AverageMonthsSurvived: calculateAverage(monthsSurvived),
		MedianLowestBalance:   percentile(lowestBalances, 0.5),
	}

	// Risk level
	switch {
	case result.InsolvencyProbability < 0.05:
		result.RiskLevel = "safe"
	case result.InsolvencyProbability < 0.25:
		result.RiskLevel = "warning"
	default:
		result.RiskLevel = "danger"
	}

	// Find first insolvency month
	for month := 0; month < input.MonthsToProject; month++ {
		monthInsolvencies := 0
		for sim := 0; sim < input.Simulations; sim++ {
			if allPaths[sim][month] < 0 {
				monthInsolvencies++
			}
		}
		if monthInsolvencies > 0 {
			result.FirstInsolvencyMonth = month + 1
			break
		}
	}

	// Calculate danger zones
	result.DangerZones = findDangerZones(allPaths, input)

	// Calculate LCR
	netOutflow3M := input.MonthlyExpenses * 3
	if input.MonthlyIncome > input.MonthlyExpenses {
		netOutflow3M = (input.MonthlyExpenses - input.MonthlyIncome*0.5) * 3 // Partial income loss
	}
	if netOutflow3M > 0 {
		result.LCR = (input.CurrentCash + input.EmergencyFund) / netOutflow3M
	} else {
		result.LCR = 999 // Very safe
	}

	// Generate percentile paths for visualization
	result.Percentile5 = getPathPercentile(allPaths, 0.05)
	result.Percentile50 = getPathPercentile(allPaths, 0.50)
	result.Percentile95 = getPathPercentile(allPaths, 0.95)

	// Generate recommendations
	result.Recommendations = generateRecommendations(input, result)

	return result
}

// runSingleSimulation runs one Monte Carlo path
func runSingleSimulation(input StressTestInput, rng *rand.Rand) (path []float64, monthsSurvived int, lowestBalance float64) {
	balance := input.CurrentCash + input.EmergencyFund
	lowestBalance = balance
	path = make([]float64, input.MonthsToProject)
	monthsSurvived = input.MonthsToProject

	for month := 0; month < input.MonthsToProject; month++ {
		// Base income with volatility
		income := input.MonthlyIncome
		if input.IncomeVolatility > 0 {
			// Log-normal shock
			shock := math.Exp(rng.NormFloat64() * input.IncomeVolatility)
			income *= shock
		}

		// Apply scenario shocks
		income = applyScenarioShock(input, month, income, rng)

		// Base expenses with small inflation noise
		expenses := input.MonthlyExpenses * (1 + rng.Float64()*0.05) // ±5% variance

		// Apply inflation shock if applicable
		if input.Scenario == ScenarioInflationSpike && input.InflationRate > 0 {
			expenses *= (1 + input.InflationRate)
		}

		// Medical shock (one-time)
		if input.Scenario == ScenarioMedicalShock && month == 0 {
			medicalCost := input.MedicalCost
			if medicalCost == 0 {
				medicalCost = input.MonthlyIncome * 6 // Default: 6 months income
			}
			expenses += medicalCost
		}

		// Planned events
		for _, event := range input.PlannedEvents {
			if event.Month == month+1 {
				expenses += event.Amount
			}
		}

		// Update balance
		balance += income - expenses
		path[month] = balance

		if balance < lowestBalance {
			lowestBalance = balance
		}

		// Check insolvency
		if balance < 0 && monthsSurvived == input.MonthsToProject {
			monthsSurvived = month + 1
		}
	}

	return path, monthsSurvived, lowestBalance
}

// applyScenarioShock applies scenario-specific income modifications
func applyScenarioShock(input StressTestInput, month int, income float64, rng *rand.Rand) float64 {
	switch input.Scenario {
	case ScenarioJobLoss:
		// Complete income loss from specified month
		if month >= input.JobLossMonth {
			return 0
		}
	case ScenarioDrySpell:
		// No freelance income for N months
		if month < input.DrySpellMonths {
			return 0
		}
	}
	return income
}

// findDangerZones identifies months with high insolvency risk
func findDangerZones(allPaths [][]float64, input StressTestInput) []DangerZone {
	var zones []DangerZone
	simCount := len(allPaths)

	for month := 0; month < input.MonthsToProject; month++ {
		insolvencies := 0
		for sim := 0; sim < simCount; sim++ {
			if allPaths[sim][month] < 0 {
				insolvencies++
			}
		}

		prob := float64(insolvencies) / float64(simCount)
		if prob > 0.10 { // >10% risk = danger zone
			reason := "High cash burn"

			// Check for planned events
			for _, event := range input.PlannedEvents {
				if event.Month == month+1 {
					reason = "Planned event: " + event.Name
					break
				}
			}

			zones = append(zones, DangerZone{
				Month:       month + 1,
				Probability: prob,
				Reason:      reason,
			})
		}
	}

	return zones
}

// generateRecommendations creates actionable advice based on results
func generateRecommendations(input StressTestInput, result StressTestResult) []string {
	var recs []string

	if result.RiskLevel == "safe" && result.InsolvencyProbability < 0.01 {
		recs = append(recs, "Tình hình tài chính ổn định. Có thể cân nhắc đầu tư phần tiền dư.")
	}

	if result.RiskLevel == "warning" {
		recs = append(recs, "Cảnh báo: Có rủi ro thiếu hụt thanh khoản. Nên xây dựng quỹ dự phòng thêm.")
	}

	if result.RiskLevel == "danger" {
		recs = append(recs, "NGUY HIỂM: Xác suất mất khả năng thanh toán cao. Cần hành động ngay!")
		recs = append(recs, "Khuyến nghị: Cắt giảm chi tiêu không thiết yếu hoặc tìm nguồn thu nhập bổ sung.")
	}

	if result.LCR < 1.0 {
		recs = append(recs, "Tỷ lệ thanh khoản (LCR) dưới 1.0. Cần tăng tiền mặt hoặc tài sản dễ chuyển đổi.")
	}

	if result.AverageMonthsSurvived < float64(input.MonthsToProject)*0.5 {
		months := int(result.AverageMonthsSurvived)
		recs = append(recs, "Trung bình chỉ duy trì được "+string(rune('0'+months/10))+string(rune('0'+months%10))+" tháng trước khi hết tiền.")
	}

	for _, zone := range result.DangerZones {
		if zone.Probability > 0.25 {
			recs = append(recs, "Tháng "+string(rune('0'+zone.Month/10))+string(rune('0'+zone.Month%10))+" là vùng nguy hiểm: "+zone.Reason)
		}
	}

	return recs
}

// getPathPercentile extracts a percentile path from all simulations
func getPathPercentile(allPaths [][]float64, p float64) []float64 {
	if len(allPaths) == 0 {
		return nil
	}

	months := len(allPaths[0])
	result := make([]float64, months)

	for month := 0; month < months; month++ {
		monthValues := make([]float64, len(allPaths))
		for sim := 0; sim < len(allPaths); sim++ {
			monthValues[sim] = allPaths[sim][month]
		}
		result[month] = percentile(monthValues, p)
	}

	return result
}

// percentile calculates the p-th percentile of a slice
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	bubbleSort(sorted)

	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}

func bubbleSort(arr []float64) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

// === PRESET SCENARIOS ===

// CreateJobLossScenario creates a job loss stress test
func CreateJobLossScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) StressTestInput {
	return StressTestInput{
		CurrentCash:     currentCash,
		MonthlyIncome:   monthlyIncome,
		MonthlyExpenses: monthlyExpenses,
		EmergencyFund:   emergencyFund,
		Scenario:        ScenarioJobLoss,
		JobLossMonth:    0, // Immediate job loss
		MonthsToProject: 12,
		Simulations:     1000,
	}
}

// CreateMedicalShockScenario creates a medical emergency stress test
func CreateMedicalShockScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) StressTestInput {
	return StressTestInput{
		CurrentCash:     currentCash,
		MonthlyIncome:   monthlyIncome,
		MonthlyExpenses: monthlyExpenses,
		EmergencyFund:   emergencyFund,
		Scenario:        ScenarioMedicalShock,
		MedicalCost:     monthlyIncome * 6, // 6 months income
		MonthsToProject: 12,
		Simulations:     1000,
	}
}

// CreateInflationScenario creates an inflation shock stress test
func CreateInflationScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64) StressTestInput {
	return StressTestInput{
		CurrentCash:     currentCash,
		MonthlyIncome:   monthlyIncome,
		MonthlyExpenses: monthlyExpenses,
		EmergencyFund:   emergencyFund,
		Scenario:        ScenarioInflationSpike,
		InflationRate:   0.20, // 20% permanent increase
		MonthsToProject: 12,
		Simulations:     1000,
	}
}

// CreateFreelancerDrySpellScenario creates a freelancer dry spell test
func CreateFreelancerDrySpellScenario(currentCash, monthlyIncome, monthlyExpenses, emergencyFund float64, dryMonths int) StressTestInput {
	return StressTestInput{
		CurrentCash:     currentCash,
		MonthlyIncome:   monthlyIncome,
		MonthlyExpenses: monthlyExpenses,
		EmergencyFund:   emergencyFund,
		Scenario:        ScenarioDrySpell,
		DrySpellMonths:  dryMonths,
		MonthsToProject: 12,
		Simulations:     1000,
	}
}
