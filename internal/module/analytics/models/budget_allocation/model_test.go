package budget_allocation

import (
	"context"
	"testing"

	"personalfinancedss/internal/module/analytics/budget_allocation/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBudgetAllocationModel_Name(t *testing.T) {
	model := NewBudgetAllocationModel()
	assert.Equal(t, "budget_allocation_gp", model.Name())
}

func TestBudgetAllocationModel_Description(t *testing.T) {
	model := NewBudgetAllocationModel()
	assert.Contains(t, model.Description(), "Goal Programming")
}

func TestBudgetAllocationModel_Dependencies(t *testing.T) {
	model := NewBudgetAllocationModel()
	deps := model.Dependencies()
	assert.Contains(t, deps, "goal_prioritization")
}

func TestBudgetAllocationModel_Validate(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Invalid input type",
			input:   "invalid",
			wantErr: true,
			errMsg:  "must be *dto.BudgetAllocationModelInput",
		},
		{
			name: "Missing user ID",
			input: &dto.BudgetAllocationModelInput{
				Year:        2024,
				Month:       12,
				TotalIncome: 5000,
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "Invalid year",
			input: &dto.BudgetAllocationModelInput{
				UserID:      uuid.New(),
				Year:        1999,
				Month:       12,
				TotalIncome: 5000,
			},
			wantErr: true,
			errMsg:  "year must be between",
		},
		{
			name: "Invalid month",
			input: &dto.BudgetAllocationModelInput{
				UserID:      uuid.New(),
				Year:        2024,
				Month:       13,
				TotalIncome: 5000,
			},
			wantErr: true,
			errMsg:  "month must be between",
		},
		{
			name: "Zero income",
			input: &dto.BudgetAllocationModelInput{
				UserID:      uuid.New(),
				Year:        2024,
				Month:       12,
				TotalIncome: 0,
			},
			wantErr: true,
			errMsg:  "total income must be positive",
		},
		{
			name: "Valid input",
			input: &dto.BudgetAllocationModelInput{
				UserID:      uuid.New(),
				Year:        2024,
				Month:       12,
				TotalIncome: 5000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.Validate(ctx, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBudgetAllocationModel_Execute_BasicScenario(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	input := &dto.BudgetAllocationModelInput{
		UserID:          uuid.New(),
		Year:            2024,
		Month:           12,
		TotalIncome:     5000,
		UseAllScenarios: false,
		MandatoryExpenses: []dto.MandatoryExpense{
			{CategoryID: uuid.New(), Name: "Rent", Amount: 1500, Priority: 1},
			{CategoryID: uuid.New(), Name: "Utilities", Amount: 200, Priority: 2},
		},
		FlexibleExpenses: []dto.FlexibleExpense{
			{CategoryID: uuid.New(), Name: "Food", MinAmount: 300, MaxAmount: 600, Priority: 3},
		},
		Debts: []dto.DebtInput{
			{DebtID: uuid.New(), Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
		},
		Goals: []dto.GoalInput{
			{GoalID: uuid.New(), Name: "Emergency Fund", Type: "emergency", Priority: "critical", RemainingAmount: 10000, SuggestedContribution: 500},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)

	output, ok := result.(*dto.BudgetAllocationModelOutput)
	require.True(t, ok)

	assert.Equal(t, input.UserID, output.UserID)
	assert.Equal(t, "2024-12", output.Period)
	assert.Equal(t, 5000.0, output.TotalIncome)
	assert.True(t, output.IsFeasible)
	assert.Len(t, output.Scenarios, 1) // Only balanced scenario
	assert.NotNil(t, output.Metadata)
	assert.GreaterOrEqual(t, output.Metadata.ComputationTime, int64(0)) // May be 0 on fast machines
}

func TestBudgetAllocationModel_Execute_AllScenarios(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	input := &dto.BudgetAllocationModelInput{
		UserID:          uuid.New(),
		Year:            2024,
		Month:           12,
		TotalIncome:     5000,
		UseAllScenarios: true,
		MandatoryExpenses: []dto.MandatoryExpense{
			{CategoryID: uuid.New(), Name: "Rent", Amount: 1500, Priority: 1},
		},
		Goals: []dto.GoalInput{
			{GoalID: uuid.New(), Name: "Vacation", Type: "savings", Priority: "medium", RemainingAmount: 3000, SuggestedContribution: 300},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)

	output := result.(*dto.BudgetAllocationModelOutput)
	assert.Len(t, output.Scenarios, 3) // Conservative, Balanced, Aggressive

	// Verify scenario types
	scenarioTypes := make(map[string]bool)
	for _, s := range output.Scenarios {
		scenarioTypes[string(s.ScenarioType)] = true
	}
	assert.True(t, scenarioTypes["conservative"])
	assert.True(t, scenarioTypes["balanced"])
	assert.True(t, scenarioTypes["aggressive"])
}

func TestBudgetAllocationModel_Execute_InfeasibleBudget(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	// Income less than mandatory expenses
	input := &dto.BudgetAllocationModelInput{
		UserID:      uuid.New(),
		Year:        2024,
		Month:       12,
		TotalIncome: 1000, // Not enough
		MandatoryExpenses: []dto.MandatoryExpense{
			{CategoryID: uuid.New(), Name: "Rent", Amount: 1500, Priority: 1},
		},
		Debts: []dto.DebtInput{
			{DebtID: uuid.New(), Name: "Loan", Balance: 10000, InterestRate: 0.10, MinimumPayment: 200},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)

	output := result.(*dto.BudgetAllocationModelOutput)
	assert.False(t, output.IsFeasible)
	assert.NotEmpty(t, output.GlobalWarnings)
	assert.Equal(t, "critical", string(output.GlobalWarnings[0].Severity))
}

func TestBudgetAllocationModel_Execute_WithSensitivity(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	input := &dto.BudgetAllocationModelInput{
		UserID:         uuid.New(),
		Year:           2024,
		Month:          12,
		TotalIncome:    5000,
		RunSensitivity: true,
		MandatoryExpenses: []dto.MandatoryExpense{
			{CategoryID: uuid.New(), Name: "Rent", Amount: 1500, Priority: 1},
			{CategoryID: uuid.New(), Name: "Utilities", Amount: 200, Priority: 2},
		},
		Debts: []dto.DebtInput{
			{DebtID: uuid.New(), Name: "Credit Card", Balance: 5000, InterestRate: 0.18, MinimumPayment: 150},
			{DebtID: uuid.New(), Name: "Car Loan", Balance: 15000, InterestRate: 0.06, MinimumPayment: 300},
		},
		Goals: []dto.GoalInput{
			{GoalID: uuid.New(), Name: "Emergency Fund", Type: "emergency", Priority: "critical", RemainingAmount: 10000, SuggestedContribution: 500},
			{GoalID: uuid.New(), Name: "Vacation", Type: "savings", Priority: "low", RemainingAmount: 3000, SuggestedContribution: 200},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)

	output := result.(*dto.BudgetAllocationModelOutput)
	require.NotNil(t, output.SensitivityResults)

	// Check income impact analysis
	assert.NotEmpty(t, output.SensitivityResults.IncomeImpact)
	for _, impact := range output.SensitivityResults.IncomeImpact {
		assert.NotEmpty(t, impact.Recommendation)
		if impact.IncomeChangePercent < 0 {
			assert.Less(t, impact.NewIncome, input.TotalIncome)
		} else {
			assert.Greater(t, impact.NewIncome, input.TotalIncome)
		}
	}

	// Check interest rate impact analysis
	assert.NotEmpty(t, output.SensitivityResults.InterestRateImpact)
	for _, impact := range output.SensitivityResults.InterestRateImpact {
		assert.NotEmpty(t, impact.AffectedDebts)
		assert.NotEmpty(t, impact.RecommendedAction)
	}

	// Check goal priority impact analysis
	assert.NotEmpty(t, output.SensitivityResults.GoalPriorityImpact)

	// Check summary
	summary := output.SensitivityResults.Summary
	assert.NotEmpty(t, summary.OverallRiskLevel)
	assert.Greater(t, summary.IncomeBreakEvenPoint, 0.0)
}

func TestBudgetAllocationModel_SensitivityAnalysis_IncomeBreakEven(t *testing.T) {
	model := NewBudgetAllocationModel()
	ctx := context.Background()

	// Tight budget scenario
	input := &dto.BudgetAllocationModelInput{
		UserID:         uuid.New(),
		Year:           2024,
		Month:          12,
		TotalIncome:    2500, // Just above mandatory
		RunSensitivity: true,
		MandatoryExpenses: []dto.MandatoryExpense{
			{CategoryID: uuid.New(), Name: "Rent", Amount: 1500, Priority: 1},
			{CategoryID: uuid.New(), Name: "Utilities", Amount: 200, Priority: 2},
		},
		Debts: []dto.DebtInput{
			{DebtID: uuid.New(), Name: "Loan", Balance: 5000, InterestRate: 0.10, MinimumPayment: 200},
		},
	}

	result, err := model.Execute(ctx, input)
	require.NoError(t, err)

	output := result.(*dto.BudgetAllocationModelOutput)
	require.NotNil(t, output.SensitivityResults)

	// Break-even should be around 1900 (1500 + 200 + 200)
	assert.InDelta(t, 1900, output.SensitivityResults.Summary.IncomeBreakEvenPoint, 100)

	// Check income impact exists and has recommendations
	assert.NotEmpty(t, output.SensitivityResults.IncomeImpact)
	for _, impact := range output.SensitivityResults.IncomeImpact {
		assert.NotEmpty(t, impact.Recommendation)
	}
}

func TestBudgetAllocationModel_DebtPriorityCalculation(t *testing.T) {
	model := NewBudgetAllocationModel()

	tests := []struct {
		rate     float64
		expected int
	}{
		{0.25, 1},  // Very high - critical
		{0.18, 10}, // High
		{0.08, 20}, // Medium
		{0.03, 30}, // Low
	}

	for _, tt := range tests {
		priority := model.calculateDebtPriority(tt.rate)
		assert.Equal(t, tt.expected, priority, "Rate %.2f should have priority %d", tt.rate, tt.expected)
	}
}

func TestBudgetAllocationModel_GoalPriorityToWeight(t *testing.T) {
	model := NewBudgetAllocationModel()

	tests := []struct {
		priority string
		expected int
	}{
		{"critical", 1},
		{"high", 10},
		{"medium", 20},
		{"low", 30},
		{"unknown", 50},
	}

	for _, tt := range tests {
		weight := model.goalPriorityToWeight(tt.priority)
		assert.Equal(t, tt.expected, weight, "Priority %s should have weight %d", tt.priority, tt.expected)
	}
}
