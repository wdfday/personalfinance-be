package minmax

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMinmaxGPSolver_SimpleBudget(t *testing.T) {
	// Simple budget allocation problem:
	// Total income: $5000
	// Goals:
	//   Rent: target $1500
	//   Food: target $500
	//   Savings: target $1000
	//   Entertainment: target $300

	solver := NewMinmaxGPSolver(5000)

	rentID := uuid.New()
	foodID := uuid.New()
	savingsID := uuid.New()
	entertainmentID := uuid.New()

	rentIdx := solver.AddVariable(MinmaxVariable{
		ID:       rentID,
		Name:     "Rent",
		Type:     "category",
		MinValue: 0,
		MaxValue: 2000,
	})

	foodIdx := solver.AddVariable(MinmaxVariable{
		ID:       foodID,
		Name:     "Food",
		Type:     "category",
		MinValue: 0,
		MaxValue: 800,
	})

	savingsIdx := solver.AddVariable(MinmaxVariable{
		ID:       savingsID,
		Name:     "Savings",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	entertainmentIdx := solver.AddVariable(MinmaxVariable{
		ID:       entertainmentID,
		Name:     "Entertainment",
		Type:     "category",
		MinValue: 0,
		MaxValue: 500,
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "rent",
		Description: "Pay rent",
		TargetValue: 1500,
		VariableIdx: rentIdx,
		Weight:      1500,
		GoalType:    "at_least",
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "food",
		Description: "Food budget",
		TargetValue: 500,
		VariableIdx: foodIdx,
		Weight:      500,
		GoalType:    "at_least",
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "savings",
		Description: "Monthly savings",
		TargetValue: 1000,
		VariableIdx: savingsIdx,
		Weight:      1000,
		GoalType:    "at_least",
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "entertainment",
		Description: "Entertainment",
		TargetValue: 300,
		VariableIdx: entertainmentIdx,
		Weight:      300,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results:")
	t.Logf("  Rent: $%.2f (target: $1500, achievement: %.1f%%)", result.VariableValues[rentID], result.GoalAchievements["rent"])
	t.Logf("  Food: $%.2f (target: $500, achievement: %.1f%%)", result.VariableValues[foodID], result.GoalAchievements["food"])
	t.Logf("  Savings: $%.2f (target: $1000, achievement: %.1f%%)", result.VariableValues[savingsID], result.GoalAchievements["savings"])
	t.Logf("  Entertainment: $%.2f (target: $300, achievement: %.1f%%)", result.VariableValues[entertainmentID], result.GoalAchievements["entertainment"])
	t.Logf("  Min achievement: %.1f%%", result.MinAchievement)
	t.Logf("  Is balanced: %v", result.IsBalanced)
	t.Logf("  Iterations: %d", result.Iterations)

	// All goals should be achieved (total target = 3300, budget = 5000)
	assert.GreaterOrEqual(t, result.VariableValues[rentID], 1500.0)
	assert.GreaterOrEqual(t, result.VariableValues[foodID], 500.0)
	assert.GreaterOrEqual(t, result.VariableValues[savingsID], 1000.0)
	assert.GreaterOrEqual(t, result.VariableValues[entertainmentID], 300.0)
}

func TestMinmaxGPSolver_InsufficientBudget_Balanced(t *testing.T) {
	// Budget too small - Minmax should balance achievement across goals
	// Total income: $1500
	// Goals: A=$1000, B=$1000, C=$1000 (total $3000, only 50% achievable)

	solver := NewMinmaxGPSolver(1500)

	ids := make([]uuid.UUID, 3)
	for i := range ids {
		ids[i] = uuid.New()
	}

	for i := range 3 {
		idx := solver.AddVariable(MinmaxVariable{
			ID:       ids[i],
			Name:     string(rune('A' + i)),
			Type:     "category",
			MinValue: 0,
			MaxValue: 1500,
		})

		solver.AddGoal(MinmaxGoal{
			ID:          string(rune('A' + i)),
			TargetValue: 1000,
			VariableIdx: idx,
			Weight:      1000,
			GoalType:    "at_least",
		})
	}

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Balanced allocation test (insufficient budget):")
	for i := range 3 {
		t.Logf("  %c: $%.2f (achievement: %.1f%%)", 'A'+i, result.VariableValues[ids[i]], result.GoalAchievements[string(rune('A'+i))])
	}
	t.Logf("  Min achievement: %.1f%%", result.MinAchievement)
	t.Logf("  Is balanced: %v", result.IsBalanced)

	// Key difference from Preemptive/Weighted:
	// Minmax should give roughly equal allocation to all goals
	// Each should get ~$500 (50% of target)
	for i := range 3 {
		assert.InDelta(t, 500.0, result.VariableValues[ids[i]], 50.0, "Goal %c should get ~$500", 'A'+i)
	}

	// All achievements should be similar (balanced)
	assert.True(t, result.IsBalanced, "Solution should be balanced")
}

// Note: TestMinmaxGPSolver_CompareWithPreemptive was removed because it
// used NewPreemptiveGPSolver, GPVariable, GPGoal from preemptive package.
// Cross-package comparison tests should be in the parent solver package.

func TestMinmaxGPSolver_WithMinimums(t *testing.T) {
	// Test that minimum values are respected
	solver := NewMinmaxGPSolver(1000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(MinmaxVariable{
		ID:       idA,
		Name:     "A",
		MinValue: 300, // Minimum
		MaxValue: 600,
	})

	idxB := solver.AddVariable(MinmaxVariable{
		ID:       idB,
		Name:     "B",
		MinValue: 200, // Minimum
		MaxValue: 600,
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "A",
		TargetValue: 500,
		VariableIdx: idxA,
		Weight:      500,
		GoalType:    "at_least",
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "B",
		TargetValue: 500,
		VariableIdx: idxB,
		Weight:      500,
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
}

func TestBuildMinmaxGPFromConstraintModel(t *testing.T) {
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

	solver := BuildMinmaxGPFromConstraintModel(model, params)

	assert.NotNil(t, solver)
	assert.Greater(t, len(solver.variables), 0)
	assert.Greater(t, len(solver.goals), 0)

	t.Logf("Built Minmax GP solver with %d variables and %d goals", len(solver.variables), len(solver.goals))

	result, err := solver.Solve()
	assert.NoError(t, err)

	t.Logf("Min achievement: %.1f%%", result.MinAchievement)
	t.Logf("Is balanced: %v", result.IsBalanced)
	t.Logf("Achieved goals: %d", len(result.AchievedGoals))
}

func TestMinmaxGPSolver_EmptyInput(t *testing.T) {
	solver := NewMinmaxGPSolver(1000)

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.VariableValues)
}

func TestMinmaxGPSolver_ZeroBudget(t *testing.T) {
	solver := NewMinmaxGPSolver(0)

	id := uuid.New()
	idx := solver.AddVariable(MinmaxVariable{
		ID:       id,
		Name:     "Test",
		MinValue: 0,
		MaxValue: 100,
	})

	solver.AddGoal(MinmaxGoal{
		ID:          "test",
		TargetValue: 50,
		VariableIdx: idx,
		Weight:      50,
		GoalType:    "at_least",
	})

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.VariableValues[id])
	assert.Equal(t, 0.0, result.GoalAchievements["test"])
}

// Note: Integration tests with GoalProgrammingSolver (SolveTriple, SolveMinmax)
// should be in the parent solver package since NewGoalProgrammingSolver is defined there.
