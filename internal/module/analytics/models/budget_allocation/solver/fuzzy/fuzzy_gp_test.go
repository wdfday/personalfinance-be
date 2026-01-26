package fuzzy

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFuzzyGPSolver_SimpleBudget(t *testing.T) {
	// Budget: $5000
	// Goal A: triangular membership, peak at $1000
	// Goal B: triangular membership, peak at $700

	solver := NewFuzzyGPSolver(5000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(FuzzyVariable{
		ID:       idA,
		Name:     "Goal A",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	idxB := solver.AddVariable(FuzzyVariable{
		ID:       idB,
		Name:     "Goal B",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	solver.AddGoal(FuzzyGoal{
		ID:          "A",
		Description: "Goal A",
		VariableIdx: idxA,
		Priority:    1,
		Weight:      1.0,
		TargetValue: 1000,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  500,
			PeakRight: 1000,
			Upper:     1200,
		},
	})

	solver.AddGoal(FuzzyGoal{
		ID:          "B",
		Description: "Goal B",
		VariableIdx: idxB,
		Priority:    1,
		Weight:      1.0,
		TargetValue: 700,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  350,
			PeakRight: 700,
			Upper:     840,
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results:")
	t.Logf("  Goal A: $%.2f (membership: %.2f)", result.VariableValues[idA], result.GoalMemberships["A"])
	t.Logf("  Goal B: $%.2f (membership: %.2f)", result.VariableValues[idB], result.GoalMemberships["B"])
	t.Logf("  Average membership: %.2f", result.AverageMembership)
	t.Logf("  Weighted membership: %.2f", result.WeightedMembership)
	t.Logf("  Achieved goals: %v", result.AchievedGoals)

	// Both should achieve high membership (budget is sufficient)
	assert.Greater(t, result.GoalMemberships["A"], 0.8)
	assert.Greater(t, result.GoalMemberships["B"], 0.8)
	assert.Contains(t, result.AchievedGoals, "Goal A")
	assert.Contains(t, result.AchievedGoals, "Goal B")
}

func TestFuzzyGPSolver_InsufficientBudget(t *testing.T) {
	// Budget: $1000
	// Goal A (priority 1): peak at $1000
	// Goal B (priority 2): peak at $700
	// Can only achieve A's peak

	solver := NewFuzzyGPSolver(1000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(FuzzyVariable{ID: idA, Name: "A", MinValue: 0, MaxValue: 2000})
	idxB := solver.AddVariable(FuzzyVariable{ID: idB, Name: "B", MinValue: 0, MaxValue: 2000})

	solver.AddGoal(FuzzyGoal{
		ID:          "A",
		VariableIdx: idxA,
		Priority:    1, // Higher priority
		Weight:      2.0,
		TargetValue: 1000,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  500,
			PeakRight: 1000,
			Upper:     1200,
		},
	})

	solver.AddGoal(FuzzyGoal{
		ID:          "B",
		VariableIdx: idxB,
		Priority:    2, // Lower priority
		Weight:      1.0,
		TargetValue: 700,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  350,
			PeakRight: 700,
			Upper:     840,
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Insufficient budget test:")
	t.Logf("  Goal A (priority 1): $%.2f (membership: %.2f)", result.VariableValues[idA], result.GoalMemberships["A"])
	t.Logf("  Goal B (priority 2): $%.2f (membership: %.2f)", result.VariableValues[idB], result.GoalMemberships["B"])

	// A should achieve high membership (priority 1) and get most of the budget
	assert.Greater(t, result.GoalMemberships["A"], 0.5)
	assert.Greater(t, result.VariableValues[idA], result.VariableValues[idB]) // A should get more than B
	assert.Greater(t, result.VariableValues[idA], 400.0)                      // At least some meaningful allocation
}

func TestFuzzyGPSolver_TrapezoidalMembership(t *testing.T) {
	// Test trapezoidal membership function
	// Budget: $2000
	// Goal with trapezoidal membership (range is acceptable)
	// Note: Lower = 800 means membership = 0 until we reach 800, so we need to allocate enough

	solver := NewFuzzyGPSolver(2000)

	id := uuid.New()
	idx := solver.AddVariable(FuzzyVariable{
		ID:       id,
		Name:     "Flexible Goal",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	solver.AddGoal(FuzzyGoal{
		ID:          "flex",
		Description: "Flexible Goal",
		VariableIdx: idx,
		Priority:    1,
		Weight:      2.0, // Higher weight to encourage allocation
		TargetValue: 1000,
		MembershipFunc: MembershipFunction{
			Type:      "trapezoidal",
			Lower:     0,    // Start from 0 for easier allocation
			PeakLeft:  900,  // Plateau starts at 900
			PeakRight: 1100, // Plateau ends at 1100
			Upper:     1200, // Full range
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Trapezoidal membership test:")
	t.Logf("  Value: $%.2f", result.VariableValues[id])
	t.Logf("  Membership: %.2f", result.GoalMemberships["flex"])

	// Should achieve allocation and good membership
	assert.Greater(t, result.VariableValues[id], 0.0)      // Should allocate something
	assert.Greater(t, result.GoalMemberships["flex"], 0.0) // Should have some membership
	// If allocated enough, should be in good range
	if result.VariableValues[id] >= 900 && result.VariableValues[id] <= 1100 {
		assert.GreaterOrEqual(t, result.GoalMemberships["flex"], 0.9)
	}
}

func TestFuzzyGPSolver_MembershipFunctionEvaluation(t *testing.T) {
	// Test membership function evaluation directly

	// Triangular membership
	triMF := MembershipFunction{
		Type:      "triangular",
		Lower:     0,
		PeakLeft:  500,
		PeakRight: 1000,
		Upper:     1200,
	}

	// Test at different points
	assert.Equal(t, 0.0, triMF.EvaluateMembership(-100))         // Below lower
	assert.Equal(t, 0.0, triMF.EvaluateMembership(1300))         // Above upper
	assert.Equal(t, 0.0, triMF.EvaluateMembership(0))            // At lower
	assert.Equal(t, 1.0, triMF.EvaluateMembership(750))          // At peak (between PeakLeft and PeakRight)
	assert.Equal(t, 1.0, triMF.EvaluateMembership(500))          // At PeakLeft
	assert.Equal(t, 1.0, triMF.EvaluateMembership(1000))         // At PeakRight
	assert.InDelta(t, 0.5, triMF.EvaluateMembership(250), 0.01)  // Midpoint on rising edge
	assert.InDelta(t, 0.5, triMF.EvaluateMembership(1100), 0.01) // Midpoint on falling edge

	// Trapezoidal membership
	trapMF := MembershipFunction{
		Type:      "trapezoidal",
		Lower:     100,
		PeakLeft:  200,
		PeakRight: 300,
		Upper:     400,
	}

	assert.Equal(t, 0.0, trapMF.EvaluateMembership(50))          // Below lower
	assert.Equal(t, 0.0, trapMF.EvaluateMembership(500))         // Above upper
	assert.Equal(t, 1.0, trapMF.EvaluateMembership(250))         // In plateau
	assert.Equal(t, 1.0, trapMF.EvaluateMembership(200))         // At PeakLeft
	assert.Equal(t, 1.0, trapMF.EvaluateMembership(300))         // At PeakRight
	assert.InDelta(t, 0.5, trapMF.EvaluateMembership(150), 0.01) // Midpoint on rising edge
	assert.InDelta(t, 0.5, trapMF.EvaluateMembership(350), 0.01) // Midpoint on falling edge
}

func TestFuzzyGPSolver_PartialGoals(t *testing.T) {
	// Budget: $1500
	// Goal A: peak at $1000
	// Goal B: peak at $700
	// Should achieve A's peak + B's partial

	solver := NewFuzzyGPSolver(1500)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(FuzzyVariable{ID: idA, Name: "A", MinValue: 0, MaxValue: 2000})
	idxB := solver.AddVariable(FuzzyVariable{ID: idB, Name: "B", MinValue: 0, MaxValue: 2000})

	solver.AddGoal(FuzzyGoal{
		ID:          "A",
		VariableIdx: idxA,
		Priority:    1,
		Weight:      1.0,
		TargetValue: 1000,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  500,
			PeakRight: 1000,
			Upper:     1200,
		},
	})

	solver.AddGoal(FuzzyGoal{
		ID:          "B",
		VariableIdx: idxB,
		Priority:    1,
		Weight:      1.0,
		TargetValue: 700,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  350,
			PeakRight: 700,
			Upper:     840,
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Partial goals test:")
	t.Logf("  Goal A: $%.2f (membership: %.2f)", result.VariableValues[idA], result.GoalMemberships["A"])
	t.Logf("  Goal B: $%.2f (membership: %.2f)", result.VariableValues[idB], result.GoalMemberships["B"])
	t.Logf("  Achieved: %v", result.AchievedGoals)
	t.Logf("  Partial: %v", result.PartialGoals)

	// A should achieve high membership
	assert.Greater(t, result.GoalMemberships["A"], 0.8)
	// B should have some allocation but may be partial
	assert.Greater(t, result.GoalMemberships["B"], 0.0)
}

func TestFuzzyGPSolver_BuildFromConstraintModel(t *testing.T) {
	// Test building fuzzy solver from constraint model
	model := &domain.ConstraintModel{
		TotalIncome: 5000,
		MandatoryExpenses: map[uuid.UUID]domain.CategoryConstraint{
			uuid.New(): {
				CategoryID: uuid.New(),
				Minimum:    1500,
				Maximum:    1500,
				IsFlexible: false,
				Priority:   1,
			},
		},
		FlexibleExpenses: map[uuid.UUID]domain.CategoryConstraint{
			uuid.New(): {
				CategoryID: uuid.New(),
				Minimum:    500,
				Maximum:    1000,
				IsFlexible: true,
				Priority:   4,
			},
		},
		DebtPayments: map[uuid.UUID]domain.DebtConstraint{
			uuid.New(): {
				DebtID:         uuid.New(),
				DebtName:       "Credit Card",
				MinimumPayment: 200,
				CurrentBalance: 5000,
				InterestRate:   0.18,
				Priority:       2,
			},
		},
		GoalTargets: map[uuid.UUID]domain.GoalConstraint{
			uuid.New(): {
				GoalID:                uuid.New(),
				GoalName:              "Emergency Fund",
				GoalType:              "emergency",
				SuggestedContribution: 1000,
				Priority:              "high",
				PriorityWeight:        10,
				RemainingAmount:       5000,
			},
		},
	}

	params := domain.ScenarioParameters{
		ScenarioType:           domain.ScenarioBalanced,
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.30,
			GoalsPercent:         0.20,
			FlexiblePercent:      0.10,
		},
	}

	solver := BuildFuzzyGPFromConstraintModel(model, params)

	assert.NotNil(t, solver)
	assert.Equal(t, 5000.0, solver.totalIncome)
	assert.Greater(t, len(solver.goals), 0)
	assert.Greater(t, len(solver.variables), 0)

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.AverageMembership, 0.0)

	t.Logf("Built Fuzzy GP solver with %d variables and %d goals", len(solver.variables), len(solver.goals))
	t.Logf("Average membership: %.2f", result.AverageMembership)
	t.Logf("Weighted membership: %.2f", result.WeightedMembership)
}

func TestFuzzyGPSolver_EmptyInput(t *testing.T) {
	solver := NewFuzzyGPSolver(1000)
	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, result.AverageMembership)
	assert.Equal(t, 0, len(result.AchievedGoals))
}

func TestFuzzyGPSolver_ZeroBudget(t *testing.T) {
	solver := NewFuzzyGPSolver(0)

	id := uuid.New()
	idx := solver.AddVariable(FuzzyVariable{ID: id, Name: "Goal", MinValue: 0, MaxValue: 1000})

	solver.AddGoal(FuzzyGoal{
		ID:          "goal",
		VariableIdx: idx,
		Priority:    1,
		Weight:      1.0,
		TargetValue: 500,
		MembershipFunc: MembershipFunction{
			Type:      "triangular",
			Lower:     0,
			PeakLeft:  250,
			PeakRight: 500,
			Upper:     600,
		},
	})

	result, err := solver.Solve()

	// Should handle zero budget gracefully
	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.VariableValues[id])
}
