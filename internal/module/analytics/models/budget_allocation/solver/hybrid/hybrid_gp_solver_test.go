package hybrid

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getBalancedScenario returns balanced scenario parameters
func getBalancedScenario() domain.ScenarioParameters {
	return domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioBalanced,
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.25,
			DebtExtraPercent:     0.25,
			GoalsPercent:         0.30,
			FlexiblePercent:      0.20,
		},
	}
}

// getConservativeScenario returns conservative scenario parameters
func getConservativeScenario() domain.ScenarioParameters {
	return domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioConservative,
		GoalContributionFactor: 0.7,
		FlexibleSpendingLevel:  0.3,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.35,
			GoalsPercent:         0.15,
			FlexiblePercent:      0.10,
		},
	}
}

// getAggressiveScenario returns aggressive scenario parameters
func getAggressiveScenario() domain.ScenarioParameters {
	return domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioAggressive,
		GoalContributionFactor: 1.3,
		FlexibleSpendingLevel:  0.7,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.15,
			DebtExtraPercent:     0.20,
			GoalsPercent:         0.40,
			FlexiblePercent:      0.25,
		},
	}
}

// createTestHybridModel creates a test constraint model for hybrid solver
func createTestHybridModel() *domain.ConstraintModel {
	return &domain.ConstraintModel{
		TotalIncome: 100_000_000, // 100M VNĐ

		MandatoryExpenses: map[uuid.UUID]domain.CategoryConstraint{
			uuid.MustParse("11111111-1111-1111-1111-111111111111"): {
				CategoryID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Minimum:    15_000_000, // Rent
				Maximum:    15_000_000,
				IsFlexible: false,
			},
			uuid.MustParse("22222222-2222-2222-2222-222222222222"): {
				CategoryID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Minimum:    5_000_000, // Utilities + Insurance
				Maximum:    5_000_000,
				IsFlexible: false,
			},
		},

		DebtPayments: map[uuid.UUID]domain.DebtConstraint{
			uuid.MustParse("33333333-3333-3333-3333-333333333333"): {
				DebtID:         uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				DebtName:       "Credit Card",
				MinimumPayment: 5_000_000,
				CurrentBalance: 50_000_000,
				InterestRate:   24,
				Priority:       1,
			},
			uuid.MustParse("44444444-4444-4444-4444-444444444444"): {
				DebtID:         uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				DebtName:       "Car Loan",
				MinimumPayment: 8_000_000,
				CurrentBalance: 200_000_000,
				InterestRate:   12,
				Priority:       10,
			},
		},

		GoalTargets: map[uuid.UUID]domain.GoalConstraint{
			uuid.MustParse("55555555-5555-5555-5555-555555555555"): {
				GoalID:                uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				GoalName:              "Emergency Fund",
				GoalType:              "emergency",
				SuggestedContribution: 20_000_000,
				Priority:              "critical",
				PriorityWeight:        1,
				RemainingAmount:       100_000_000,
			},
			uuid.MustParse("66666666-6666-6666-6666-666666666666"): {
				GoalID:                uuid.MustParse("66666666-6666-6666-6666-666666666666"),
				GoalName:              "Japan Trip",
				GoalType:              "travel",
				SuggestedContribution: 15_000_000,
				Priority:              "medium",
				PriorityWeight:        20,
				RemainingAmount:       50_000_000,
			},
			uuid.MustParse("77777777-7777-7777-7777-777777777777"): {
				GoalID:                uuid.MustParse("77777777-7777-7777-7777-777777777777"),
				GoalName:              "New Laptop",
				GoalType:              "purchase",
				SuggestedContribution: 10_000_000,
				Priority:              "low",
				PriorityWeight:        30,
				RemainingAmount:       30_000_000,
			},
		},

		FlexibleExpenses: map[uuid.UUID]domain.CategoryConstraint{
			uuid.MustParse("88888888-8888-8888-8888-888888888888"): {
				CategoryID: uuid.MustParse("88888888-8888-8888-8888-888888888888"),
				Minimum:    5_000_000,
				Maximum:    10_000_000,
				IsFlexible: true,
				Priority:   50,
			},
			uuid.MustParse("99999999-9999-9999-9999-999999999999"): {
				CategoryID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				Minimum:    1_000_000,
				Maximum:    5_000_000,
				IsFlexible: true,
				Priority:   60,
			},
		},
	}
}

func TestHybridGPSolver_Phase1_EnsureMinimums(t *testing.T) {
	model := createTestHybridModel()
	params := getBalancedScenario()

	solver := NewHybridGPSolver(model, params)
	result := &HybridResult{
		MandatoryAllocations: make(map[uuid.UUID]float64),
		DebtMinAllocations:   make(map[uuid.UUID]float64),
		Buckets:              make(map[string]*BudgetBucket),
		FinalAllocations:     make(map[uuid.UUID]AllocationDetail),
		IsFeasible:           true,
	}

	solver.phase1EnsureMinimums(result)

	// Check mandatory allocations
	assert.Equal(t, 15_000_000.0, result.MandatoryAllocations[uuid.MustParse("11111111-1111-1111-1111-111111111111")])
	assert.Equal(t, 5_000_000.0, result.MandatoryAllocations[uuid.MustParse("22222222-2222-2222-2222-222222222222")])

	// Check debt minimum allocations
	assert.Equal(t, 5_000_000.0, result.DebtMinAllocations[uuid.MustParse("33333333-3333-3333-3333-333333333333")])
	assert.Equal(t, 8_000_000.0, result.DebtMinAllocations[uuid.MustParse("44444444-4444-4444-4444-444444444444")])

	// Check totals
	expectedTotal := 15_000_000 + 5_000_000 + 5_000_000 + 8_000_000 // 33M
	assert.Equal(t, float64(expectedTotal), result.Phase1Total)
	assert.Equal(t, 100_000_000.0-float64(expectedTotal), result.Surplus) // 67M surplus
	assert.True(t, result.IsFeasible)
}

func TestHybridGPSolver_Phase1_Infeasible(t *testing.T) {
	model := createTestHybridModel()
	model.TotalIncome = 20_000_000 // Only 20M income (less than 33M minimum)

	params := getBalancedScenario()
	solver := NewHybridGPSolver(model, params)
	result := &HybridResult{
		MandatoryAllocations: make(map[uuid.UUID]float64),
		DebtMinAllocations:   make(map[uuid.UUID]float64),
		Buckets:              make(map[string]*BudgetBucket),
		FinalAllocations:     make(map[uuid.UUID]AllocationDetail),
		IsFeasible:           true,
	}

	solver.phase1EnsureMinimums(result)

	assert.False(t, result.IsFeasible)
	assert.Equal(t, 13_000_000.0, result.DeficitAmount) // 33M - 20M = 13M deficit
	assert.Equal(t, 0.0, result.Surplus)
}

func TestHybridGPSolver_Phase2_AllocateSurplus(t *testing.T) {
	model := createTestHybridModel()
	params := getBalancedScenario()

	solver := NewHybridGPSolver(model, params)
	result := &HybridResult{
		MandatoryAllocations: make(map[uuid.UUID]float64),
		DebtMinAllocations:   make(map[uuid.UUID]float64),
		Buckets:              make(map[string]*BudgetBucket),
		FinalAllocations:     make(map[uuid.UUID]AllocationDetail),
		IsFeasible:           true,
		Surplus:              67_000_000, // From Phase 1
	}

	solver.phase2AllocateSurplus(result)

	// Check bucket budgets (Balanced scenario: 25%, 25%, 30%, 20%)
	assert.InDelta(t, 67_000_000*0.25, result.Buckets["emergency"].Budget, 1)
	assert.InDelta(t, 67_000_000*0.25, result.Buckets["debt_extra"].Budget, 1)
	assert.InDelta(t, 67_000_000*0.30, result.Buckets["goals"].Budget, 1)
	assert.InDelta(t, 67_000_000*0.20, result.Buckets["flexible"].Budget, 1)

	// Check items populated
	assert.Len(t, result.Buckets["emergency"].Items, 1)  // Emergency fund goal
	assert.Len(t, result.Buckets["debt_extra"].Items, 2) // 2 debts
	assert.Len(t, result.Buckets["goals"].Items, 2)      // Japan trip + Laptop
	assert.Len(t, result.Buckets["flexible"].Items, 2)   // 2 flexible categories
}

func TestHybridGPSolver_FullSolve(t *testing.T) {
	model := createTestHybridModel()
	params := getBalancedScenario()

	solver := NewHybridGPSolver(model, params)
	result, err := solver.Solve()

	require.NoError(t, err)
	assert.True(t, result.IsFeasible)

	// Check Phase 1 completed
	assert.Equal(t, 33_000_000.0, result.Phase1Total)
	assert.Equal(t, 67_000_000.0, result.Surplus)

	// Check allocations exist
	assert.NotEmpty(t, result.FinalAllocations)

	// Check total allocated is reasonable
	assert.Greater(t, result.TotalAllocated, result.Phase1Total)
	assert.LessOrEqual(t, result.TotalAllocated, 100_000_000.0)

	// Check rewards
	assert.Greater(t, result.TotalReward, 0.0)
	assert.Greater(t, result.MaxPossibleReward, 0.0)

	t.Logf("Total Allocated: %.2f", result.TotalAllocated)
	t.Logf("Total Reward: %.2f / %.2f (%.1f%%)",
		result.TotalReward, result.MaxPossibleReward,
		result.TotalReward/result.MaxPossibleReward*100)
}

// Note: Integration tests with GoalProgrammingSolver should be in the parent solver package
// since NewGoalProgrammingSolver is defined there, not here.

func TestHybridGPSolver_DifferentScenarios(t *testing.T) {
	model := createTestHybridModel()

	scenarios := []struct {
		name   string
		params domain.ScenarioParameters
	}{
		{"Conservative", getConservativeScenario()},
		{"Balanced", getBalancedScenario()},
		{"Aggressive", getAggressiveScenario()},
	}

	t.Logf("\n=== Scenario Comparison ===")

	for _, sc := range scenarios {
		solver := NewHybridGPSolver(model, sc.params)
		result, err := solver.Solve()
		require.NoError(t, err)

		t.Logf("\n%s Scenario:", sc.name)
		t.Logf("  Emergency Budget: %.0f (%.0f%%)",
			result.Buckets["emergency"].Budget,
			sc.params.SurplusAllocation.EmergencyFundPercent*100)
		t.Logf("  Debt Extra Budget: %.0f (%.0f%%)",
			result.Buckets["debt_extra"].Budget,
			sc.params.SurplusAllocation.DebtExtraPercent*100)
		t.Logf("  Goals Budget: %.0f (%.0f%%)",
			result.Buckets["goals"].Budget,
			sc.params.SurplusAllocation.GoalsPercent*100)
		t.Logf("  Flexible Budget: %.0f (%.0f%%)",
			result.Buckets["flexible"].Budget,
			sc.params.SurplusAllocation.FlexiblePercent*100)
		t.Logf("  Total Reward: %.1f / %.1f",
			result.TotalReward, result.MaxPossibleReward)
	}
}
