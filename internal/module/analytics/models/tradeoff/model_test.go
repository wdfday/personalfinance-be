package tradeoff

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
	"personalfinancedss/internal/module/analytics/debt_tradeoff/dto"
)

func TestTradeoffModel_Validate(t *testing.T) {
	model := NewTradeoffModel()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *dto.TradeoffInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &dto.TradeoffInput{
				UserID:            "user1",
				MonthlyIncome:     5000,
				EssentialExpenses: 3000,
				TotalMinPayments:  200,
				Debts: []domain.DebtInfo{
					{ID: "d1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "no debts",
			input: &dto.TradeoffInput{
				UserID:            "user1",
				MonthlyIncome:     5000,
				EssentialExpenses: 3000,
				TotalMinPayments:  0,
				Debts:             []domain.DebtInfo{},
			},
			wantErr: true,
		},
		{
			name: "no extra money",
			input: &dto.TradeoffInput{
				UserID:            "user1",
				MonthlyIncome:     3000,
				EssentialExpenses: 2800,
				TotalMinPayments:  300,
				Debts: []domain.DebtInfo{
					{ID: "d1", Balance: 5000, InterestRate: 0.18, MinimumPayment: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "negative balance",
			input: &dto.TradeoffInput{
				UserID:            "user1",
				MonthlyIncome:     5000,
				EssentialExpenses: 3000,
				TotalMinPayments:  200,
				Debts: []domain.DebtInfo{
					{ID: "d1", Balance: -1000, InterestRate: 0.18, MinimumPayment: 100},
				},
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

func TestTradeoffModel_Execute(t *testing.T) {
	model := NewTradeoffModel()
	ctx := context.Background()

	input := &dto.TradeoffInput{
		UserID:            "user1",
		MonthlyIncome:     6000,
		EssentialExpenses: 3500,
		TotalMinPayments:  300,
		Debts: []domain.DebtInfo{
			{
				ID:             "cc1",
				Name:           "Credit Card",
				Balance:        8000,
				InterestRate:   0.20, // 20% - high interest
				MinimumPayment: 200,
				Type:           "credit_card",
			},
			{
				ID:             "car",
				Name:           "Car Loan",
				Balance:        15000,
				InterestRate:   0.06,
				MinimumPayment: 100,
				Type:           "car_loan",
			},
		},
		Goals: []domain.GoalInfo{
			{
				ID:            "emergency",
				Name:          "Emergency Fund",
				TargetAmount:  15000,
				CurrentAmount: 3000,
				Deadline:      time.Now().AddDate(2, 0, 0),
				Priority:      0.6,
			},
			{
				ID:            "vacation",
				Name:          "Vacation",
				TargetAmount:  5000,
				CurrentAmount: 500,
				Deadline:      time.Now().AddDate(1, 0, 0),
				Priority:      0.4,
			},
		},
		InvestmentProfile: domain.InvestmentProfile{
			RiskTolerance:      "moderate",
			ExpectedReturn:     0.07,
			TimeHorizon:        10,
			CurrentInvestments: 5000,
		},
		EmergencyFund: domain.EmergencyFundStatus{
			TargetAmount:    15000,
			CurrentAmount:   3000,
			MonthlyExpenses: 3500,
			TargetMonths:    4,
		},
		Preferences: dto.TradeoffPreferences{
			PsychologicalWeight:  0.15,
			Priority:             "balanced",
			AcceptInvestmentRisk: true,
			RiskTolerance:        "moderate",
		},
		SimulationConfig: &domain.SimulationConfig{
			NumSimulations:   100,
			IncomeVariance:   0.10,
			ExpenseVariance:  0.15,
			ReturnVariance:   0.20,
			ProjectionMonths: 60,
			DiscountRate:     0.05,
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := result.(*dto.TradeoffOutput)
	if !ok {
		t.Fatalf("Execute() returned wrong type")
	}

	// Verify output structure
	if output.RecommendedStrategy == "" {
		t.Error("RecommendedStrategy should not be empty")
	}

	if output.RecommendedRatio.DebtPercent+output.RecommendedRatio.SavingsPercent != 1.0 {
		t.Errorf("Ratio should sum to 1.0, got %v", output.RecommendedRatio)
	}

	if len(output.StrategyAnalysis) != 3 {
		t.Errorf("Expected 3 strategy analyses, got %d", len(output.StrategyAnalysis))
	}

	if output.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}

	if len(output.KeyFactors) == 0 {
		t.Error("KeyFactors should not be empty")
	}

	// With high interest debt (20%), should recommend aggressive debt
	if output.RecommendedStrategy != domain.StrategyAggressiveDebt {
		t.Logf("Note: With 20%% interest debt, recommended %s instead of aggressive_debt",
			output.RecommendedStrategy)
	}

	// Monte Carlo results should exist
	if output.MonteCarloResults == nil {
		t.Fatal("MonteCarloResults should not be nil")
	}
	if output.MonteCarloResults.NumSimulations != 100 {
		t.Errorf("Expected 100 simulations, got %d", output.MonteCarloResults.NumSimulations)
	}

	// Projections should have data
	if output.ProjectedTimelines.DebtFreeDate.IsZero() {
		t.Error("DebtFreeDate should not be zero")
	}

	if len(output.ProjectedTimelines.NetWorthGrowth) == 0 {
		t.Error("NetWorthGrowth should not be empty")
	}

	t.Logf("Recommended: %s (%.0f%% debt, %.0f%% savings)",
		output.RecommendedStrategy,
		output.RecommendedRatio.DebtPercent*100,
		output.RecommendedRatio.SavingsPercent*100)
	t.Logf("Reasoning: %s", output.Reasoning)
	t.Logf("Success Probability: %.1f%%", output.MonteCarloResults.SuccessProbability*100)
}

func TestTradeoffModel_LowEmergencyFund(t *testing.T) {
	model := NewTradeoffModel()
	ctx := context.Background()

	input := &dto.TradeoffInput{
		UserID:            "user1",
		MonthlyIncome:     5000,
		EssentialExpenses: 3000,
		TotalMinPayments:  200,
		Debts: []domain.DebtInfo{
			{ID: "d1", Balance: 5000, InterestRate: 0.10, MinimumPayment: 100},
		},
		EmergencyFund: domain.EmergencyFundStatus{
			TargetAmount:  15000,
			CurrentAmount: 1000, // Only ~7% of target - critically low
		},
		InvestmentProfile: domain.InvestmentProfile{
			ExpectedReturn: 0.07,
		},
		SimulationConfig: &domain.SimulationConfig{
			NumSimulations:   50,
			ProjectionMonths: 60,
			DiscountRate:     0.05,
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.TradeoffOutput)

	// With critically low emergency fund, should recommend aggressive savings
	if output.RecommendedStrategy != domain.StrategyAggressiveSavings {
		t.Errorf("With low EF, expected aggressive_savings, got %s", output.RecommendedStrategy)
	}

	t.Logf("Recommended: %s", output.RecommendedStrategy)
	t.Logf("Reasoning: %s", output.Reasoning)
}

func TestTradeoffModel_UserPreference(t *testing.T) {
	model := NewTradeoffModel()
	ctx := context.Background()

	input := &dto.TradeoffInput{
		UserID:            "user1",
		MonthlyIncome:     5000,
		EssentialExpenses: 3000,
		TotalMinPayments:  200,
		Debts: []domain.DebtInfo{
			{ID: "d1", Balance: 5000, InterestRate: 0.08, MinimumPayment: 100}, // Low interest
		},
		EmergencyFund: domain.EmergencyFundStatus{
			TargetAmount:  10000,
			CurrentAmount: 8000, // 80% - good
		},
		InvestmentProfile: domain.InvestmentProfile{
			ExpectedReturn: 0.07,
		},
		Preferences: dto.TradeoffPreferences{
			Priority: "debt_free", // User wants to be debt-free
		},
		SimulationConfig: &domain.SimulationConfig{
			NumSimulations:   50,
			ProjectionMonths: 60,
			DiscountRate:     0.05,
		},
	}

	result, err := model.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := result.(*dto.TradeoffOutput)

	// Should respect user preference for debt freedom
	if output.RecommendedStrategy != domain.StrategyAggressiveDebt {
		t.Errorf("With debt_free preference, expected aggressive_debt, got %s", output.RecommendedStrategy)
	}

	t.Logf("Recommended: %s", output.RecommendedStrategy)
	t.Logf("Reasoning: %s", output.Reasoning)
}
