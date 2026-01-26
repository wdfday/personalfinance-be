package tradeoff

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
	"personalfinancedss/internal/module/analytics/debt_tradeoff/dto"
	"personalfinancedss/internal/module/analytics/models/tradeoff/engine"
)

type TradeoffModel struct {
	name        string
	description string
	calculator  *engine.FinancialCalculator
	simulator   *engine.MonteCarloSimulator
}

func NewTradeoffModel() *TradeoffModel {
	return &TradeoffModel{
		name:        "savings_debt_tradeoff",
		description: "Decision tree analysis with Monte Carlo simulation for optimal debt-savings allocation",
		calculator:  engine.NewFinancialCalculator(),
		simulator:   engine.NewMonteCarloSimulator(),
	}
}

func (m *TradeoffModel) Name() string           { return m.name }
func (m *TradeoffModel) Description() string    { return m.description }
func (m *TradeoffModel) Dependencies() []string { return []string{} }

func (m *TradeoffModel) Validate(ctx context.Context, input interface{}) error {
	ti, ok := input.(*dto.TradeoffInput)
	if !ok {
		return errors.New("input must be *dto.TradeoffInput type")
	}

	// If no debts, return early (no validation needed)
	if len(ti.Debts) == 0 {
		return nil
	}

	if ti.MonthlyIncome <= 0 {
		return errors.New("monthly income must be positive")
	}
	for i, debt := range ti.Debts {
		if debt.Balance < 0 {
			return fmt.Errorf("debt %d: balance cannot be negative", i)
		}
		if debt.InterestRate < 0 || debt.InterestRate > 1 {
			return fmt.Errorf("debt %d: interest rate must be between 0 and 1", i)
		}
	}
	if ti.CalculateExtraMoney() <= 0 {
		return errors.New("no extra money available for allocation")
	}
	return nil
}

func (m *TradeoffModel) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	ti := input.(*dto.TradeoffInput)

	// If no debts, recommend aggressive savings strategy
	if len(ti.Debts) == 0 {
		config := ti.GetSimulationConfig()
		extraMoney := ti.CalculateExtraMoney()

		// All strategies become savings-focused when no debts
		results := []domain.StrategyResult{
			{
				Strategy: domain.StrategyAggressiveSavings,
				Ratio:    domain.AllocationRatio{DebtPercent: 0, SavingsPercent: 1.0},
				Score:    100,
			},
		}

		rec := results[0]
		mc := m.runMonteCarlo(rec.Ratio, ti, config)
		proj := m.generateProjections(rec.Ratio, extraMoney, ti, config)
		recs := m.generateRecommendations(rec, ti, mc)

		return &dto.TradeoffOutput{
			RecommendedStrategy: domain.StrategyAggressiveSavings,
			RecommendedRatio:    rec.Ratio,
			StrategyAnalysis:    results,
			Reasoning:           "No debts to pay off. Focus on aggressive savings.",
			KeyFactors:          []string{"No debt obligations", "Full allocation to savings"},
			ProjectedTimelines:  proj,
			MonteCarloResults:   mc,
			Recommendations:     recs,
		}, nil
	}

	config := ti.GetSimulationConfig()
	extraMoney := ti.CalculateExtraMoney()

	strategies := []domain.Strategy{domain.StrategyAggressiveDebt, domain.StrategyBalanced, domain.StrategyAggressiveSavings}
	results := make([]domain.StrategyResult, 0, 3)
	for _, s := range strategies {
		ratio := m.getStrategyRatio(s)
		results = append(results, m.analyzeStrategy(s, ratio, extraMoney, ti, config))
	}

	m.scoreStrategies(results, ti)
	rec, reasoning, factors := m.selectBestStrategy(results, ti)
	mc := m.runMonteCarlo(rec.Ratio, ti, config)
	proj := m.generateProjections(rec.Ratio, extraMoney, ti, config)
	recs := m.generateRecommendations(rec, ti, mc)

	return &dto.TradeoffOutput{
		RecommendedStrategy: rec.Strategy,
		RecommendedRatio:    rec.Ratio,
		StrategyAnalysis:    results,
		Reasoning:           reasoning,
		KeyFactors:          factors,
		ProjectedTimelines:  proj,
		MonteCarloResults:   mc,
		Recommendations:     recs,
	}, nil
}

func (m *TradeoffModel) getStrategyRatio(s domain.Strategy) domain.AllocationRatio {
	switch s {
	case domain.StrategyAggressiveDebt:
		return domain.AllocationRatio{DebtPercent: 0.75, SavingsPercent: 0.25}
	case domain.StrategyBalanced:
		return domain.AllocationRatio{DebtPercent: 0.50, SavingsPercent: 0.50}
	case domain.StrategyAggressiveSavings:
		return domain.AllocationRatio{DebtPercent: 0.25, SavingsPercent: 0.75}
	}
	return domain.AllocationRatio{DebtPercent: 0.50, SavingsPercent: 0.50}
}

func (m *TradeoffModel) analyzeStrategy(s domain.Strategy, ratio domain.AllocationRatio, extra float64, ti *dto.TradeoffInput, cfg domain.SimulationConfig) domain.StrategyResult {
	debtPay := extra * ratio.DebtPercent
	savePay := extra * ratio.SavingsPercent
	months, totalInt, intSaved := m.calculator.SimulateDebtPayoff(ti.Debts, debtPay, cfg.ProjectionMonths)
	ret := ti.InvestmentProfile.ExpectedReturn
	if ret == 0 {
		ret = 0.07
	}
	invVal := m.calculator.SimulateInvestmentGrowth(ti.InvestmentProfile.CurrentInvestments, savePay, ret, cfg.ProjectionMonths)
	npv := intSaved + m.calculator.PresentValue(invVal, cfg.DiscountRate/12, cfg.ProjectionMonths)
	goals := make(map[string]int)
	for _, g := range ti.Goals {
		c := savePay
		if g.Priority > 0 {
			c = savePay * g.Priority
		}
		goals[g.ID] = m.calculator.CalculateGoalMonths(g.CurrentAmount, g.TargetAmount, c, ret)
	}
	return domain.StrategyResult{
		Strategy: s, Ratio: ratio, NPV: npv, TotalInterestPaid: totalInt, InterestSaved: intSaved,
		InvestmentValue: invVal, MonthsToDebtFree: months, TimeToGoals: goals,
		RiskScore: m.calcRisk(ratio, ti), Pros: m.getPros(s), Cons: m.getCons(s),
	}
}

func (m *TradeoffModel) calcRisk(ratio domain.AllocationRatio, ti *dto.TradeoffInput) float64 {
	score := ratio.SavingsPercent * 3.0
	ef := ti.GetEmergencyFundProgress()
	if ef < 0.5 {
		score += 3.0
	} else if ef < 1.0 {
		score += 1.5
	}
	if ti.GetHighestInterestRate() > 0.15 && ratio.DebtPercent < 0.5 {
		score += 2.0
	}
	switch ti.Preferences.RiskTolerance {
	case "conservative":
		score *= 1.2
	case "aggressive":
		score *= 0.8
	}
	if score > 10 {
		return 10
	}
	return score
}

func (m *TradeoffModel) scoreStrategies(results []domain.StrategyResult, ti *dto.TradeoffInput) {
	maxNPV, minDF := results[0].NPV, results[0].MonthsToDebtFree
	for _, r := range results {
		if r.NPV > maxNPV {
			maxNPV = r.NPV
		}
		if r.MonthsToDebtFree < minDF {
			minDF = r.MonthsToDebtFree
		}
	}
	pw := ti.Preferences.PsychologicalWeight
	if pw == 0 {
		pw = 0.15
	}
	for i := range results {
		ns := 0.0
		if maxNPV > 0 {
			ns = results[i].NPV / maxNPV
		}
		rs := 1.0 - (results[i].RiskScore / 10.0)
		ds := 0.0
		if results[i].MonthsToDebtFree > 0 {
			ds = float64(minDF) / float64(results[i].MonthsToDebtFree)
		}
		ps := 0.5
		switch results[i].Strategy {
		case domain.StrategyAggressiveDebt:
			ps = 1.0
		case domain.StrategyBalanced:
			ps = 0.7
		}
		results[i].Score = ns*0.35 + rs*0.25 + ds*0.25 + ps*pw
	}
}

func (m *TradeoffModel) selectBestStrategy(results []domain.StrategyResult, ti *dto.TradeoffInput) (domain.StrategyResult, string, []string) {
	kf := make([]string, 0)
	avg := ti.GetWeightedAvgInterestRate()
	high := ti.GetHighestInterestRate()
	ef := ti.GetEmergencyFundProgress()

	if ef < 0.3 {
		return m.find(results, domain.StrategyAggressiveSavings), "Emergency fund critically low - prioritize building safety net", append(kf, fmt.Sprintf("Emergency fund at %.0f%% of target", ef*100))
	}
	switch ti.Preferences.Priority {
	case "debt_free":
		return m.find(results, domain.StrategyAggressiveDebt), "Following your preference to become debt-free faster", append(kf, "User priority: debt freedom")
	case "wealth_building":
		if ef >= 0.5 {
			return m.find(results, domain.StrategyAggressiveSavings), "Following your preference for wealth building", append(kf, "User priority: wealth building")
		}
	}
	if high > 0.18 {
		return m.find(results, domain.StrategyAggressiveDebt), "High interest debt detected - prioritize debt payoff", append(kf, fmt.Sprintf("Highest debt rate: %.1f%%", high*100))
	}
	if ef < 0.5 {
		return m.find(results, domain.StrategyAggressiveSavings), "Emergency fund below 50% - prioritize savings", append(kf, fmt.Sprintf("Emergency fund at %.0f%% of target", ef*100))
	}
	if avg > 0.10 {
		return m.find(results, domain.StrategyAggressiveDebt), "Moderate-high interest debt - debt payoff provides better return", append(kf, fmt.Sprintf("Average debt rate: %.1f%%", avg*100))
	}
	if avg <= 0.10 && ef >= 0.8 {
		return m.find(results, domain.StrategyAggressiveSavings), "Low interest debt and strong emergency fund - maximize wealth building", append(kf, "Low debt rates and adequate emergency fund")
	}
	best := 0
	for i, r := range results {
		if r.Score > results[best].Score {
			best = i
		}
	}
	return results[best], "Balanced approach recommended based on overall analysis", append(kf, fmt.Sprintf("Best composite score: %.2f", results[best].Score))
}

func (m *TradeoffModel) find(results []domain.StrategyResult, s domain.Strategy) domain.StrategyResult {
	for _, r := range results {
		if r.Strategy == s {
			return r
		}
	}
	return results[0]
}

func (m *TradeoffModel) runMonteCarlo(ratio domain.AllocationRatio, ti *dto.TradeoffInput, cfg domain.SimulationConfig) *domain.MonteCarloResult {
	return m.simulator.RunSimulation(engine.SimulationInput{
		Debts: ti.Debts, MonthlyIncome: ti.MonthlyIncome, EssentialExpenses: ti.EssentialExpenses,
		TotalMinPayments: ti.TotalMinPayments, Ratio: ratio, ExpectedReturn: ti.InvestmentProfile.ExpectedReturn,
		InitialSavings: ti.InvestmentProfile.CurrentInvestments, Goals: ti.Goals, Config: cfg,
	})
}

func (m *TradeoffModel) generateProjections(ratio domain.AllocationRatio, extra float64, ti *dto.TradeoffInput, cfg domain.SimulationConfig) dto.ProjectionResult {
	dp := extra * ratio.DebtPercent
	sp := extra * ratio.SavingsPercent
	months, _, _ := m.calculator.SimulateDebtPayoff(ti.Debts, dp, cfg.ProjectionMonths)
	efGap := ti.GetEmergencyFundGap()
	efMonths := 0
	if sp > 0 && efGap > 0 {
		efMonths = int(efGap / sp)
	}
	ret := ti.InvestmentProfile.ExpectedReturn
	if ret == 0 {
		ret = 0.07
	}
	gd := make(map[string]time.Time)
	for _, g := range ti.Goals {
		c := sp
		if g.Priority > 0 {
			c = sp * g.Priority
		}
		gd[g.ID] = time.Now().AddDate(0, m.calculator.CalculateGoalMonths(g.CurrentAmount, g.TargetAmount, c, ret), 0)
	}
	nw := m.calculator.GenerateNetWorthTimeline(ti.Debts, ti.InvestmentProfile.CurrentInvestments, dp, sp, ret, cfg.ProjectionMonths, 6)
	return dto.ProjectionResult{DebtFreeDate: time.Now().AddDate(0, months, 0), EmergencyFundDate: time.Now().AddDate(0, efMonths, 0), GoalDates: gd, NetWorthGrowth: nw}
}

func (m *TradeoffModel) generateRecommendations(sel domain.StrategyResult, ti *dto.TradeoffInput, mc *domain.MonteCarloResult) []string {
	recs := make([]string, 0)
	switch sel.Strategy {
	case domain.StrategyAggressiveDebt:
		recs = append(recs, "Focus extra payments on highest interest debt first (Avalanche method)")
		if ti.GetHighestInterestRate() > 0.15 {
			recs = append(recs, "Consider balance transfer or debt consolidation for high-rate debts")
		}
	case domain.StrategyAggressiveSavings:
		recs = append(recs, "Maximize retirement account contributions for tax benefits")
		recs = append(recs, "Consider low-cost index funds for long-term growth")
	case domain.StrategyBalanced:
		recs = append(recs, "Review allocation quarterly and adjust based on progress")
	}
	if ti.GetEmergencyFundProgress() < 0.5 {
		recs = append(recs, fmt.Sprintf("Build emergency fund to at least 50%% (currently %.0f%%)", ti.GetEmergencyFundProgress()*100))
	}
	if mc != nil && mc.SuccessProbability < 0.7 {
		recs = append(recs, "Consider increasing income or reducing expenses to improve success probability")
	}
	return recs
}

func (m *TradeoffModel) getPros(s domain.Strategy) []string {
	switch s {
	case domain.StrategyAggressiveDebt:
		return []string{"Minimize total interest paid", "Become debt-free faster", "Psychological relief", "Guaranteed return"}
	case domain.StrategyBalanced:
		return []string{"Balanced approach", "Flexibility", "Moderate risk", "Progress on multiple fronts"}
	case domain.StrategyAggressiveSavings:
		return []string{"Maximize compound growth", "Build wealth faster", "Time in market advantage", "Better prepared for opportunities"}
	}
	return []string{}
}

func (m *TradeoffModel) getCons(s domain.Strategy) []string {
	switch s {
	case domain.StrategyAggressiveDebt:
		return []string{"Slower wealth accumulation", "Miss compound growth", "Less flexibility"}
	case domain.StrategyBalanced:
		return []string{"Not optimized for single goal", "Slower progress"}
	case domain.StrategyAggressiveSavings:
		return []string{"Pay more interest", "Debt lasts longer", "Returns not guaranteed", "Higher risk"}
	}
	return []string{}
}
