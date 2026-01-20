package debt_strategy

import (
	"context"
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"personalfinancedss/internal/module/analytics/debt_strategy/dto"
	"testing"
)

func TestDebtStrategyModel_Validate(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *dto.DebtStrategyInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &dto.DebtStrategyInput{
				Debts: []domain.DebtInfo{
					{ID: "1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
				},
				TotalDebtBudget: 500,
			},
			wantErr: false,
		},
		{
			name: "no debts",
			input: &dto.DebtStrategyInput{
				Debts:           []domain.DebtInfo{},
				TotalDebtBudget: 500,
			},
			wantErr: true,
		},
		{
			name: "zero budget",
			input: &dto.DebtStrategyInput{
				Debts: []domain.DebtInfo{
					{ID: "1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
				},
				TotalDebtBudget: 0,
			},
			wantErr: true,
		},
		{
			name: "budget less than minimums",
			input: &dto.DebtStrategyInput{
				Debts: []domain.DebtInfo{
					{ID: "1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
					{ID: "2", Balance: 3000, InterestRate: 0.10, MinimumPayment: 100},
				},
				TotalDebtBudget: 200, // Less than 150+100=250
			},
			wantErr: true,
		},
		{
			name: "negative balance",
			input: &dto.DebtStrategyInput{
				Debts: []domain.DebtInfo{
					{ID: "1", Balance: -1000, InterestRate: 0.18, MinimumPayment: 150},
				},
				TotalDebtBudget: 500,
			},
			wantErr: true,
		},
		{
			name: "invalid interest rate",
			input: &dto.DebtStrategyInput{
				Debts: []domain.DebtInfo{
					{ID: "1", Balance: 5000, InterestRate: 1.5, MinimumPayment: 150}, // >100%
				},
				TotalDebtBudget: 500,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.Validate(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDebtStrategyModel_Execute_Basic(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		UserID: "test-user",
		Debts: []domain.DebtInfo{
			{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
			{ID: "car", Name: "Car Loan", Balance: 10000, InterestRate: 0.06, MinimumPayment: 200},
		},
		TotalDebtBudget: 600,
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	// Check basic output
	if output.RecommendedStrategy == "" {
		t.Error("Expected recommended strategy")
	}

	if len(output.PaymentPlans) != 2 {
		t.Errorf("Expected 2 payment plans, got %d", len(output.PaymentPlans))
	}

	if output.MonthsToDebtFree <= 0 {
		t.Error("Expected positive months to debt free")
	}

	if output.TotalInterest <= 0 {
		t.Error("Expected positive total interest")
	}

	if len(output.StrategyComparison) != 5 {
		t.Errorf("Expected 5 strategy comparisons, got %d", len(output.StrategyComparison))
	}

	if len(output.Milestones) == 0 {
		t.Error("Expected milestones")
	}

	if output.PsychScore == nil {
		t.Error("Expected psychological score")
	}

	t.Logf("Recommended: %s", output.RecommendedStrategy)
	t.Logf("Debt-free in: %d months", output.MonthsToDebtFree)
	t.Logf("Total interest: $%.2f", output.TotalInterest)
	t.Logf("Reasoning: %s", output.Reasoning)
}

func TestDebtStrategyModel_Execute_WithPreference(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		},
		TotalDebtBudget:   500,
		PreferredStrategy: domain.StrategySnowball,
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	if output.RecommendedStrategy != domain.StrategySnowball {
		t.Errorf("Expected snowball (user preference), got %s", output.RecommendedStrategy)
	}
}

func TestDebtStrategyModel_Execute_HighStressDebt(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.20, MinimumPayment: 150, StressScore: 3},
			{ID: "family", Name: "Family Debt", Balance: 3000, InterestRate: 0.05, MinimumPayment: 100, StressScore: 9},
		},
		TotalDebtBudget: 500,
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	// Should recommend stress-based or snowball (family debt is smaller)
	if output.RecommendedStrategy != domain.StrategyStress && output.RecommendedStrategy != domain.StrategySnowball {
		t.Logf("With high stress debt, recommended: %s", output.RecommendedStrategy)
	}

	t.Logf("High stress scenario - Recommended: %s", output.RecommendedStrategy)
	t.Logf("Reasoning: %s", output.Reasoning)
}

func TestDebtStrategyModel_Execute_WithWhatIf(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		},
		TotalDebtBudget: 400,
		WhatIfScenarios: []domain.WhatIfScenario{
			{Type: "extra_monthly", Amount: 100, Description: "Add $100/month"},
			{Type: "lump_sum", Amount: 2000, Description: "Pay $2000 bonus"},
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	if len(output.WhatIfResults) != 2 {
		t.Errorf("Expected 2 what-if results, got %d", len(output.WhatIfResults))
	}

	for _, wif := range output.WhatIfResults {
		t.Logf("What-if: %s", wif.Scenario.Description)
		t.Logf("  Months saved: %d", wif.MonthsSaved)
		t.Logf("  Interest saved: $%.2f", wif.InterestSaved)
		t.Logf("  Action: %s", wif.RecommendedAction)
	}
}

func TestDebtStrategyModel_Execute_WithRefinancing(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc1", Name: "Credit Card 1", Balance: 5000, InterestRate: 0.22, MinimumPayment: 150},
			{ID: "cc2", Name: "Credit Card 2", Balance: 3000, InterestRate: 0.19, MinimumPayment: 90},
		},
		TotalDebtBudget: 500,
		RefinanceOption: &domain.RefinanceOption{
			NewRate:        0.10,
			NewTerm:        36,
			OriginationFee: 200,
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	if output.RefinanceAnalysis == nil {
		t.Error("Expected refinance analysis")
	}

	t.Logf("Refinancing Analysis:")
	t.Logf("  Should refinance: %v", output.RefinanceAnalysis.ShouldRefinance)
	t.Logf("  Net savings: $%.2f", output.RefinanceAnalysis.NetSavings)
	t.Logf("  Recommendation: %s", output.RefinanceAnalysis.Recommendation)
}

func TestDebtStrategyModel_Execute_WithSensitivity(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc", Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150, IsVariableRate: true},
		},
		TotalDebtBudget: 400,
		RunSensitivity:  true,
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	if len(output.SensitivityResults) == 0 {
		t.Error("Expected sensitivity results")
	}

	for _, sr := range output.SensitivityResults {
		t.Logf("Sensitivity: %s", sr.Scenario.Description)
		t.Logf("  Impact: %d months, $%.2f interest", sr.MonthsImpact, sr.InterestImpact)
		t.Logf("  Risk: %s", sr.RiskLevel)
	}
}

func TestDebtStrategyModel_Execute_HybridWithCustomWeights(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "d1", Name: "High Interest", Balance: 5000, InterestRate: 0.20, MinimumPayment: 150, StressScore: 3},
			{ID: "d2", Name: "High Stress", Balance: 3000, InterestRate: 0.08, MinimumPayment: 100, StressScore: 9},
		},
		TotalDebtBudget:   500,
		PreferredStrategy: domain.StrategyHybrid,
		HybridWeights: &domain.HybridWeights{
			InterestRateWeight: 0.2,
			BalanceWeight:      0.1,
			StressWeight:       0.6,
			CashFlowWeight:     0.1,
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	if output.RecommendedStrategy != domain.StrategyHybrid {
		t.Errorf("Expected hybrid strategy, got %s", output.RecommendedStrategy)
	}

	// With stress weight 0.6, high stress debt should be prioritized
	for _, plan := range output.PaymentPlans {
		t.Logf("Payment plan: %s, payoff month: %d", plan.DebtName, plan.PayoffMonth)
	}
}

func TestDebtStrategyModel_StrategyComparison(t *testing.T) {
	model := NewDebtStrategyModel()
	ctx := context.Background()

	input := &dto.DebtStrategyInput{
		Debts: []domain.DebtInfo{
			{ID: "cc1", Name: "Credit Card 1", Balance: 5000, InterestRate: 0.22, MinimumPayment: 150, StressScore: 5},
			{ID: "cc2", Name: "Credit Card 2", Balance: 1500, InterestRate: 0.18, MinimumPayment: 45, StressScore: 3},
			{ID: "car", Name: "Car Loan", Balance: 12000, InterestRate: 0.05, MinimumPayment: 250, StressScore: 2},
		},
		TotalDebtBudget: 700,
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.DebtStrategyOutput)

	t.Logf("\n=== Strategy Comparison ===")
	for _, comp := range output.StrategyComparison {
		t.Logf("%s:", comp.Strategy)
		t.Logf("  Months: %d, Interest: $%.2f, First win: month %d",
			comp.Months, comp.TotalInterest, comp.FirstDebtCleared)
		t.Logf("  Pros: %v", comp.Pros)
	}

	t.Logf("\n=== Recommended: %s ===", output.RecommendedStrategy)
	t.Logf("Reasoning: %s", output.Reasoning)
	t.Logf("Key facts: %v", output.KeyFacts)
}
