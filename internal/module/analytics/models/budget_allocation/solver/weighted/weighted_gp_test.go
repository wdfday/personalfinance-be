package weighted

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWeightedGPSolver_SimpleBudget(t *testing.T) {
	// Simple budget allocation problem:
	// Total income: $5000
	// Goals with weights:
	//   Rent: weight 10.0, target $1500
	//   Food: weight 8.0, target $500
	//   Savings: weight 5.0, target $1000
	//   Entertainment: weight 2.0, target $300

	solver := NewWeightedGPSolver(5000)

	rentID := uuid.New()
	foodID := uuid.New()
	savingsID := uuid.New()
	entertainmentID := uuid.New()

	rentIdx := solver.AddVariable(WGPVariable{
		ID:       rentID,
		Name:     "Rent",
		Type:     "category",
		MinValue: 0,
		MaxValue: 2000,
		Weight:   10.0,
	})

	foodIdx := solver.AddVariable(WGPVariable{
		ID:       foodID,
		Name:     "Food",
		Type:     "category",
		MinValue: 0,
		MaxValue: 800,
		Weight:   8.0,
	})

	savingsIdx := solver.AddVariable(WGPVariable{
		ID:       savingsID,
		Name:     "Savings",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
		Weight:   5.0,
	})

	entertainmentIdx := solver.AddVariable(WGPVariable{
		ID:       entertainmentID,
		Name:     "Entertainment",
		Type:     "category",
		MinValue: 0,
		MaxValue: 500,
		Weight:   2.0,
	})

	// Add goals
	solver.AddGoal(WGPGoal{
		ID:          "rent",
		Description: "Pay rent",
		TargetValue: 1500,
		VariableIdx: rentIdx,
		Weight:      10.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "food",
		Description: "Food budget",
		TargetValue: 500,
		VariableIdx: foodIdx,
		Weight:      8.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "savings",
		Description: "Monthly savings",
		TargetValue: 1000,
		VariableIdx: savingsIdx,
		Weight:      5.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "entertainment",
		Description: "Entertainment",
		TargetValue: 300,
		VariableIdx: entertainmentIdx,
		Weight:      2.0,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results:")
	t.Logf("  Rent: $%.2f (target: $1500)", result.VariableValues[rentID])
	t.Logf("  Food: $%.2f (target: $500)", result.VariableValues[foodID])
	t.Logf("  Savings: $%.2f (target: $1000)", result.VariableValues[savingsID])
	t.Logf("  Entertainment: $%.2f (target: $300)", result.VariableValues[entertainmentID])
	t.Logf("  Achieved goals: %v", result.AchievedGoals)
	t.Logf("  Weighted deviation: %.2f", result.WeightedDeviation)
	t.Logf("  Iterations: %d", result.Iterations)

	// All goals should be achieved (total target = 3300, budget = 5000)
	assert.GreaterOrEqual(t, result.VariableValues[rentID], 1500.0)
	assert.GreaterOrEqual(t, result.VariableValues[foodID], 500.0)
	assert.GreaterOrEqual(t, result.VariableValues[savingsID], 1000.0)
	assert.GreaterOrEqual(t, result.VariableValues[entertainmentID], 300.0)
}

func TestWeightedGPSolver_InsufficientBudget(t *testing.T) {
	// Budget too small - WGP should allocate proportionally by weight
	// Total income: $2000
	// Goals:
	//   Rent: weight 10.0, target $1500
	//   Food: weight 5.0, target $500
	//   Savings: weight 2.0, target $1000

	solver := NewWeightedGPSolver(2000)

	rentID := uuid.New()
	foodID := uuid.New()
	savingsID := uuid.New()

	rentIdx := solver.AddVariable(WGPVariable{
		ID:       rentID,
		Name:     "Rent",
		Type:     "category",
		MinValue: 0,
		MaxValue: 1500,
		Weight:   10.0,
	})

	foodIdx := solver.AddVariable(WGPVariable{
		ID:       foodID,
		Name:     "Food",
		Type:     "category",
		MinValue: 0,
		MaxValue: 600,
		Weight:   5.0,
	})

	savingsIdx := solver.AddVariable(WGPVariable{
		ID:       savingsID,
		Name:     "Savings",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
		Weight:   2.0,
	})

	solver.AddGoal(WGPGoal{
		ID:          "rent",
		TargetValue: 1500,
		VariableIdx: rentIdx,
		Weight:      10.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "food",
		TargetValue: 500,
		VariableIdx: foodIdx,
		Weight:      5.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "savings",
		TargetValue: 1000,
		VariableIdx: savingsIdx,
		Weight:      2.0,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results (insufficient budget):")
	t.Logf("  Rent: $%.2f (target: $1500, weight: 10)", result.VariableValues[rentID])
	t.Logf("  Food: $%.2f (target: $500, weight: 5)", result.VariableValues[foodID])
	t.Logf("  Savings: $%.2f (target: $1000, weight: 2)", result.VariableValues[savingsID])
	t.Logf("  Achieved: %v", result.AchievedGoals)
	t.Logf("  Partial: %v", result.PartialGoals)
	t.Logf("  Unachieved: %v", result.UnachievedGoals)
	t.Logf("  Weighted deviation: %.2f", result.WeightedDeviation)

	// Higher weight goals should get more allocation
	// Rent (weight 10) should get more than Food (weight 5) proportionally
	assert.Greater(t, result.VariableValues[rentID], result.VariableValues[foodID])

	// Total should not exceed budget
	total := result.VariableValues[rentID] + result.VariableValues[foodID] + result.VariableValues[savingsID]
	assert.LessOrEqual(t, total, 2000.0+0.01)
}

func TestWeightedGPSolver_ProportionalAllocation(t *testing.T) {
	// Test that allocation is proportional to weights
	// Total income: $1000
	// All goals have same target but different weights

	solver := NewWeightedGPSolver(1000)

	ids := make([]uuid.UUID, 3)
	for i := range ids {
		ids[i] = uuid.New()
	}

	weights := []float64{6.0, 3.0, 1.0} // 60%, 30%, 10%

	for i := range 3 {
		idx := solver.AddVariable(WGPVariable{
			ID:       ids[i],
			Name:     string(rune('A' + i)),
			Type:     "category",
			MinValue: 0,
			MaxValue: 1000,
			Weight:   weights[i],
		})

		solver.AddGoal(WGPGoal{
			ID:          string(rune('A' + i)),
			TargetValue: 500, // Same target for all
			VariableIdx: idx,
			Weight:      weights[i],
			GoalType:    "at_least",
		})
	}

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Proportional allocation test:")
	for i := range 3 {
		t.Logf("  %c (weight %.1f): $%.2f", 'A'+i, weights[i], result.VariableValues[ids[i]])
	}

	// A (weight 6) should get more than B (weight 3) which should get more than C (weight 1)
	assert.Greater(t, result.VariableValues[ids[0]], result.VariableValues[ids[1]])
	assert.Greater(t, result.VariableValues[ids[1]], result.VariableValues[ids[2]])
}

func TestWeightedGPSolver_WithMinimums(t *testing.T) {
	// Test that minimum values are respected
	// Total income: $1000
	// Variable A: min $300, target $500
	// Variable B: min $200, target $500

	solver := NewWeightedGPSolver(1000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(WGPVariable{
		ID:       idA,
		Name:     "A",
		Type:     "category",
		MinValue: 300, // Minimum
		MaxValue: 600,
		Weight:   5.0,
	})

	idxB := solver.AddVariable(WGPVariable{
		ID:       idB,
		Name:     "B",
		Type:     "category",
		MinValue: 200, // Minimum
		MaxValue: 600,
		Weight:   5.0,
	})

	solver.AddGoal(WGPGoal{
		ID:          "A",
		TargetValue: 500,
		VariableIdx: idxA,
		Weight:      5.0,
		GoalType:    "at_least",
	})

	solver.AddGoal(WGPGoal{
		ID:          "B",
		TargetValue: 500,
		VariableIdx: idxB,
		Weight:      5.0,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Minimum values test:")
	t.Logf("  A (min $300): $%.2f", result.VariableValues[idA])
	t.Logf("  B (min $200): $%.2f", result.VariableValues[idB])

	// Both should be at least their minimums
	assert.GreaterOrEqual(t, result.VariableValues[idA], 300.0)
	assert.GreaterOrEqual(t, result.VariableValues[idB], 200.0)

	// Both should reach target (budget is sufficient: 1000 >= 500 + 500)
	assert.GreaterOrEqual(t, result.VariableValues[idA], 500.0)
	assert.GreaterOrEqual(t, result.VariableValues[idB], 500.0)
}

func TestBuildWeightedGPFromConstraintModel(t *testing.T) {
	mandatoryID := uuid.New()
	flexibleID := uuid.New()
	debtID := uuid.New()
	goalID := uuid.New()

	model := &domain.ConstraintModel{
		TotalIncome: 5000,
		MandatoryExpenses: map[uuid.UUID]domain.CategoryConstraint{
			mandatoryID: {
				CategoryID: mandatoryID,
				Minimum:    1500,
				Maximum:    1500,
				IsFlexible: false,
				Priority:   1,
			},
		},
		FlexibleExpenses: map[uuid.UUID]domain.CategoryConstraint{
			flexibleID: {
				CategoryID: flexibleID,
				Minimum:    200,
				Maximum:    500,
				IsFlexible: true,
				Priority:   5,
			},
		},
		DebtPayments: map[uuid.UUID]domain.DebtConstraint{
			debtID: {
				DebtID:         debtID,
				DebtName:       "Credit Card",
				MinimumPayment: 100,
				CurrentBalance: 5000,
				InterestRate:   18.0,
				Priority:       1,
			},
		},
		GoalTargets: map[uuid.UUID]domain.GoalConstraint{
			goalID: {
				GoalID:                goalID,
				GoalName:              "Emergency Fund",
				GoalType:              "emergency",
				SuggestedContribution: 500,
				Priority:              "high",
				PriorityWeight:        5,
				RemainingAmount:       10000,
			},
		},
	}

	params := domain.ScenarioParameters{
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.4,
			DebtExtraPercent:     0.3,
			GoalsPercent:         0.2,
			FlexiblePercent:      0.1,
		},
	}

	solver := BuildWeightedGPFromConstraintModel(model, params)

	assert.NotNil(t, solver)
	assert.Greater(t, len(solver.variables), 0)
	assert.Greater(t, len(solver.goals), 0)

	t.Logf("Built WGP solver with %d variables and %d goals", len(solver.variables), len(solver.goals))

	result, err := solver.Solve()
	assert.NoError(t, err)

	t.Logf("Achieved goals: %d", len(result.AchievedGoals))
	t.Logf("Partial goals: %d", len(result.PartialGoals))
	t.Logf("Unachieved goals: %d", len(result.UnachievedGoals))
	t.Logf("Weighted deviation: %.2f", result.WeightedDeviation)
}

func TestWeightedGPSolver_EmptyInput(t *testing.T) {
	solver := NewWeightedGPSolver(1000)

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.VariableValues)
	assert.Equal(t, 0.0, result.WeightedDeviation)
}

func TestWeightedGPSolver_ZeroBudget(t *testing.T) {
	solver := NewWeightedGPSolver(0)

	id := uuid.New()
	idx := solver.AddVariable(WGPVariable{
		ID:       id,
		Name:     "Test",
		Type:     "category",
		MinValue: 0,
		MaxValue: 100,
		Weight:   1.0,
	})

	solver.AddGoal(WGPGoal{
		ID:          "test",
		TargetValue: 50,
		VariableIdx: idx,
		Weight:      1.0,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.VariableValues[id])
	assert.Contains(t, result.UnachievedGoals, "test")
}

// Note: Integration tests with GoalProgrammingSolver should be in the parent solver package
// since NewGoalProgrammingSolver and SolveDual are defined there.
