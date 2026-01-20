package preemptive

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPreemptiveGPSolver_SimpleBudget(t *testing.T) {
	// Simple budget allocation problem:
	// Total income: $5000
	// Goals:
	//   Priority 1: Rent >= $1500 (mandatory)
	//   Priority 2: Food >= $500 (mandatory)
	//   Priority 3: Savings >= $1000 (goal)
	//   Priority 4: Entertainment >= $300 (flexible)

	solver := NewPreemptiveGPSolver(5000)

	// Add variables
	rentID := uuid.New()
	foodID := uuid.New()
	savingsID := uuid.New()
	entertainmentID := uuid.New()

	rentIdx := solver.AddVariable(GPVariable{
		ID:       rentID,
		Name:     "Rent",
		Type:     "category",
		MinValue: 0, // Start from 0, goal will push to target
		MaxValue: 2000,
	})

	foodIdx := solver.AddVariable(GPVariable{
		ID:       foodID,
		Name:     "Food",
		Type:     "category",
		MinValue: 0,
		MaxValue: 800,
	})

	savingsIdx := solver.AddVariable(GPVariable{
		ID:       savingsID,
		Name:     "Savings",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	entertainmentIdx := solver.AddVariable(GPVariable{
		ID:       entertainmentID,
		Name:     "Entertainment",
		Type:     "category",
		MinValue: 0,
		MaxValue: 500,
	})

	// Add goals in priority order
	solver.AddGoal(GPGoal{
		ID:          "rent",
		Description: "Pay rent",
		Priority:    1,
		TargetValue: 1500,
		VariableIdx: rentIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	solver.AddGoal(GPGoal{
		ID:          "food",
		Description: "Food budget",
		Priority:    2,
		TargetValue: 500,
		VariableIdx: foodIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	solver.AddGoal(GPGoal{
		ID:          "savings",
		Description: "Monthly savings",
		Priority:    3,
		TargetValue: 1000,
		VariableIdx: savingsIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	solver.AddGoal(GPGoal{
		ID:          "entertainment",
		Description: "Entertainment",
		Priority:    4,
		TargetValue: 300,
		VariableIdx: entertainmentIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	// Solve
	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.True(t, result.IsFeasible)

	t.Logf("Results:")
	t.Logf("  Rent: $%.2f (target: $1500)", result.VariableValues[rentID])
	t.Logf("  Food: $%.2f (target: $500)", result.VariableValues[foodID])
	t.Logf("  Savings: $%.2f (target: $1000)", result.VariableValues[savingsID])
	t.Logf("  Entertainment: $%.2f (target: $300)", result.VariableValues[entertainmentID])
	t.Logf("  Achieved goals: %v", result.AchievedGoals)
	t.Logf("  Unachieved goals: %v", result.UnachievedGoals)

	// All goals should be achieved (total target = 3300, budget = 5000)
	assert.GreaterOrEqual(t, result.VariableValues[rentID], 1500.0)
	assert.GreaterOrEqual(t, result.VariableValues[foodID], 500.0)
	assert.GreaterOrEqual(t, result.VariableValues[savingsID], 1000.0)
	assert.GreaterOrEqual(t, result.VariableValues[entertainmentID], 300.0)
}

func TestPreemptiveGPSolver_InsufficientBudget(t *testing.T) {
	// Budget too small to satisfy all goals
	// Total income: $2500
	// Goals:
	//   Priority 1: Rent >= $1500 (mandatory)
	//   Priority 2: Food >= $500 (mandatory)
	//   Priority 3: Savings >= $1000 (goal) - won't be fully satisfied

	solver := NewPreemptiveGPSolver(2500)

	rentID := uuid.New()
	foodID := uuid.New()
	savingsID := uuid.New()

	rentIdx := solver.AddVariable(GPVariable{
		ID:       rentID,
		Name:     "Rent",
		Type:     "category",
		MinValue: 0,
		MaxValue: 1500,
	})

	foodIdx := solver.AddVariable(GPVariable{
		ID:       foodID,
		Name:     "Food",
		Type:     "category",
		MinValue: 0,
		MaxValue: 600,
	})

	savingsIdx := solver.AddVariable(GPVariable{
		ID:       savingsID,
		Name:     "Savings",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	solver.AddGoal(GPGoal{
		ID:          "rent",
		Priority:    1,
		TargetValue: 1500,
		VariableIdx: rentIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	solver.AddGoal(GPGoal{
		ID:          "food",
		Priority:    2,
		TargetValue: 500,
		VariableIdx: foodIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	solver.AddGoal(GPGoal{
		ID:          "savings",
		Priority:    3,
		TargetValue: 1000,
		VariableIdx: savingsIdx,
		GoalType:    "at_least",
		Weight:      1.0,
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results:")
	t.Logf("  Rent: $%.2f (target: $1500)", result.VariableValues[rentID])
	t.Logf("  Food: $%.2f (target: $500)", result.VariableValues[foodID])
	t.Logf("  Savings: $%.2f (target: $1000)", result.VariableValues[savingsID])
	t.Logf("  Achieved: %v", result.AchievedGoals)
	t.Logf("  Unachieved: %v", result.UnachievedGoals)

	// High priority goals should be satisfied
	assert.InDelta(t, 1500.0, result.VariableValues[rentID], 1.0)
	assert.InDelta(t, 500.0, result.VariableValues[foodID], 1.0)

	// Savings should get whatever is left ($2500 - $1500 - $500 = $500)
	assert.InDelta(t, 500.0, result.VariableValues[savingsID], 1.0)

	// Savings goal should be unachieved (got $500, target was $1000)
	assert.Contains(t, result.UnachievedGoals, "savings")
}

func TestPreemptiveGPSolver_PriorityOrder(t *testing.T) {
	// Test that higher priority goals are satisfied before lower priority
	// Total income: $2500
	// Priority 1: A >= $1000
	// Priority 2: B >= $1000
	// Priority 3: C >= $1000
	// Only 2.5 can be fully satisfied

	solver := NewPreemptiveGPSolver(2500)

	ids := make([]uuid.UUID, 3)
	for i := range ids {
		ids[i] = uuid.New()
	}

	for i := 0; i < 3; i++ {
		idx := solver.AddVariable(GPVariable{
			ID:       ids[i],
			Name:     string(rune('A' + i)),
			Type:     "category",
			MinValue: 0,
			MaxValue: 1500,
		})

		solver.AddGoal(GPGoal{
			ID:          string(rune('A' + i)),
			Priority:    i + 1,
			TargetValue: 1000,
			VariableIdx: idx,
			GoalType:    "at_least",
			Weight:      1.0,
		})
	}

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Priority order test:")
	for i := 0; i < 3; i++ {
		t.Logf("  %c (Priority %d): $%.2f", 'A'+i, i+1, result.VariableValues[ids[i]])
	}
	t.Logf("  Achieved: %v", result.AchievedGoals)
	t.Logf("  Unachieved: %v", result.UnachievedGoals)

	// First 2 priorities should be fully satisfied (A, B = $1000 each)
	assert.InDelta(t, 1000.0, result.VariableValues[ids[0]], 1.0) // A
	assert.InDelta(t, 1000.0, result.VariableValues[ids[1]], 1.0) // B

	// C gets whatever is left ($500)
	assert.InDelta(t, 500.0, result.VariableValues[ids[2]], 1.0) // C

	// A and B should be achieved, C should not
	assert.Contains(t, result.AchievedGoals, "A")
	assert.Contains(t, result.AchievedGoals, "B")
	assert.Contains(t, result.UnachievedGoals, "C")
}

func TestPreemptiveGPSolver_WeightWithinPriority(t *testing.T) {
	// Test that within same priority, higher weight goals get priority
	// Total income: $1500
	// Priority 1: A >= $1000 (weight 1.0)
	// Priority 1: B >= $1000 (weight 2.0) - should be satisfied first

	solver := NewPreemptiveGPSolver(1500)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(GPVariable{
		ID:       idA,
		Name:     "A",
		Type:     "category",
		MinValue: 0,
		MaxValue: 1500,
	})

	idxB := solver.AddVariable(GPVariable{
		ID:       idB,
		Name:     "B",
		Type:     "category",
		MinValue: 0,
		MaxValue: 1500,
	})

	// Same priority, different weights
	solver.AddGoal(GPGoal{
		ID:          "A",
		Priority:    1,
		TargetValue: 1000,
		VariableIdx: idxA,
		GoalType:    "at_least",
		Weight:      1.0, // Lower weight
	})

	solver.AddGoal(GPGoal{
		ID:          "B",
		Priority:    1,
		TargetValue: 1000,
		VariableIdx: idxB,
		GoalType:    "at_least",
		Weight:      2.0, // Higher weight - should be satisfied first
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Weight test:")
	t.Logf("  A (weight 1.0): $%.2f", result.VariableValues[idA])
	t.Logf("  B (weight 2.0): $%.2f", result.VariableValues[idB])

	// B should be fully satisfied (higher weight)
	assert.InDelta(t, 1000.0, result.VariableValues[idB], 1.0)

	// A gets the remainder ($500)
	assert.InDelta(t, 500.0, result.VariableValues[idA], 1.0)
}

func TestBuildPreemptiveGPFromConstraintModel(t *testing.T) {
	// Test building GP solver from constraint model
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

	solver := BuildPreemptiveGPFromConstraintModel(model, params)

	assert.NotNil(t, solver)
	assert.Greater(t, len(solver.variables), 0)
	assert.Greater(t, len(solver.goals), 0)

	t.Logf("Built GP solver with %d variables and %d goals", len(solver.variables), len(solver.goals))

	// Solve and check results
	result, err := solver.Solve()
	assert.NoError(t, err)

	t.Logf("Solution feasible: %v", result.IsFeasible)
	t.Logf("Achieved goals: %d / %d", len(result.AchievedGoals), len(result.AchievedGoals)+len(result.UnachievedGoals))
	t.Logf("Total deviation: %.2f", result.TotalDeviation)

	// Should be feasible with $5000 income
	assert.True(t, result.IsFeasible)
}
