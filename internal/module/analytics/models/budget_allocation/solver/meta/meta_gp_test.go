package meta

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMetaGPSolver_SimpleBudget(t *testing.T) {
	// Budget: $5000
	// Goal A: min=$500, satisfactory=$800, ideal=$1000
	// Goal B: min=$300, satisfactory=$500, ideal=$700

	solver := NewMetaGPSolver(5000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(MetaVariable{
		ID:       idA,
		Name:     "Goal A",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	idxB := solver.AddVariable(MetaVariable{
		ID:       idB,
		Name:     "Goal B",
		Type:     "goal",
		MinValue: 0,
		MaxValue: 2000,
	})

	solver.AddGoal(MetaGoal{
		ID:          "A",
		Description: "Goal A",
		VariableIdx: idxA,
		Priority:    1,
		Weight:      1.0,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 500, Reward: 30},
			{Level: "satisfactory", Value: 800, Reward: 60},
			{Level: "ideal", Value: 1000, Reward: 100},
		},
	})

	solver.AddGoal(MetaGoal{
		ID:          "B",
		Description: "Goal B",
		VariableIdx: idxB,
		Priority:    1,
		Weight:      1.0,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 300, Reward: 30},
			{Level: "satisfactory", Value: 500, Reward: 60},
			{Level: "ideal", Value: 700, Reward: 100},
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Results:")
	t.Logf("  Goal A: $%.2f (level: %s)", result.VariableValues[idA], result.GoalLevels["A"])
	t.Logf("  Goal B: $%.2f (level: %s)", result.VariableValues[idB], result.GoalLevels["B"])
	t.Logf("  Total reward: %.1f / %.1f (%.1f%%)", result.TotalReward, result.MaxPossibleReward, result.RewardRatio*100)
	t.Logf("  Ideal goals: %v", result.IdealGoals)

	// Both should reach ideal level (budget is sufficient)
	assert.Equal(t, "ideal", result.GoalLevels["A"])
	assert.Equal(t, "ideal", result.GoalLevels["B"])
	assert.Contains(t, result.IdealGoals, "A")
	assert.Contains(t, result.IdealGoals, "B")
}

func TestMetaGPSolver_InsufficientBudget(t *testing.T) {
	// Budget: $1000
	// Goal A (priority 1): min=$500, satisfactory=$800, ideal=$1000
	// Goal B (priority 2): min=$300, satisfactory=$500, ideal=$700
	// Can only achieve A's ideal OR A's satisfactory + B's satisfactory

	solver := NewMetaGPSolver(1000)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(MetaVariable{ID: idA, Name: "A", MinValue: 0, MaxValue: 2000})
	idxB := solver.AddVariable(MetaVariable{ID: idB, Name: "B", MinValue: 0, MaxValue: 2000})

	solver.AddGoal(MetaGoal{
		ID:          "A",
		VariableIdx: idxA,
		Priority:    1, // Higher priority
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 500, Reward: 30},
			{Level: "satisfactory", Value: 800, Reward: 60},
			{Level: "ideal", Value: 1000, Reward: 100},
		},
	})

	solver.AddGoal(MetaGoal{
		ID:          "B",
		VariableIdx: idxB,
		Priority:    2, // Lower priority
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 300, Reward: 30},
			{Level: "satisfactory", Value: 500, Reward: 60},
			{Level: "ideal", Value: 700, Reward: 100},
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Insufficient budget test:")
	t.Logf("  Goal A (priority 1): $%.2f (level: %s)", result.VariableValues[idA], result.GoalLevels["A"])
	t.Logf("  Goal B (priority 2): $%.2f (level: %s)", result.VariableValues[idB], result.GoalLevels["B"])
	t.Logf("  Total reward: %.1f / %.1f", result.TotalReward, result.MaxPossibleReward)

	// A should reach ideal (priority 1, $1000)
	// B gets nothing (no budget left)
	assert.Equal(t, "ideal", result.GoalLevels["A"])
	assert.GreaterOrEqual(t, result.VariableValues[idA], 1000.0)
}

func TestMetaGPSolver_MultiLevelAchievement(t *testing.T) {
	// Budget: $1500
	// Goal A: min=$500, satisfactory=$800, ideal=$1000
	// Goal B: min=$300, satisfactory=$500, ideal=$700
	// Should achieve A's ideal ($1000) + B's satisfactory ($500)

	solver := NewMetaGPSolver(1500)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(MetaVariable{ID: idA, Name: "A", MinValue: 0, MaxValue: 2000})
	idxB := solver.AddVariable(MetaVariable{ID: idB, Name: "B", MinValue: 0, MaxValue: 2000})

	solver.AddGoal(MetaGoal{
		ID:          "A",
		VariableIdx: idxA,
		Priority:    1,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 500, Reward: 30},
			{Level: "satisfactory", Value: 800, Reward: 60},
			{Level: "ideal", Value: 1000, Reward: 100},
		},
	})

	solver.AddGoal(MetaGoal{
		ID:          "B",
		VariableIdx: idxB,
		Priority:    2,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 300, Reward: 30},
			{Level: "satisfactory", Value: 500, Reward: 60},
			{Level: "ideal", Value: 700, Reward: 100},
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Multi-level achievement test:")
	t.Logf("  Goal A: $%.2f (level: %s)", result.VariableValues[idA], result.GoalLevels["A"])
	t.Logf("  Goal B: $%.2f (level: %s)", result.VariableValues[idB], result.GoalLevels["B"])
	t.Logf("  Total reward: %.1f", result.TotalReward)

	// A should reach ideal, B should reach satisfactory
	assert.Equal(t, "ideal", result.GoalLevels["A"])
	assert.Equal(t, "satisfactory", result.GoalLevels["B"])
}

func TestMetaGPSolver_SamePriorityRewardOptimization(t *testing.T) {
	// Budget: $1300
	// Goal A (priority 1): min=$500 (reward 50), ideal=$800 (reward 100)
	// Goal B (priority 1): min=$300 (reward 50), ideal=$500 (reward 100)
	// Same priority - should achieve both at high levels

	solver := NewMetaGPSolver(1300)

	idA := uuid.New()
	idB := uuid.New()

	idxA := solver.AddVariable(MetaVariable{ID: idA, Name: "A", MinValue: 0, MaxValue: 2000})
	idxB := solver.AddVariable(MetaVariable{ID: idB, Name: "B", MinValue: 0, MaxValue: 2000})

	solver.AddGoal(MetaGoal{
		ID:          "A",
		VariableIdx: idxA,
		Priority:    1,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 500, Reward: 50},
			{Level: "ideal", Value: 800, Reward: 100},
		},
	})

	solver.AddGoal(MetaGoal{
		ID:          "B",
		VariableIdx: idxB,
		Priority:    1,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 300, Reward: 50},
			{Level: "ideal", Value: 500, Reward: 100},
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)

	t.Logf("Same priority optimization test:")
	t.Logf("  Goal A: $%.2f (level: %s)", result.VariableValues[idA], result.GoalLevels["A"])
	t.Logf("  Goal B: $%.2f (level: %s)", result.VariableValues[idB], result.GoalLevels["B"])
	t.Logf("  Total reward: %.1f / %.1f", result.TotalReward, result.MaxPossibleReward)

	// Both should achieve ideal (budget $1300 >= $800 + $500)
	assert.Contains(t, result.AchievedGoals, "A")
	assert.Contains(t, result.AchievedGoals, "B")
	assert.Equal(t, "ideal", result.GoalLevels["A"])
	assert.Equal(t, "ideal", result.GoalLevels["B"])
}

func TestBuildMetaGPFromConstraintModel(t *testing.T) {
	mandatoryID := uuid.New()
	debtID := uuid.New()
	goalID := uuid.New()

	model := &domain.ConstraintModel{
		TotalIncome: 5000,
		MandatoryExpenses: map[uuid.UUID]domain.CategoryConstraint{
			mandatoryID: {
				CategoryID: mandatoryID,
				Minimum:    1500,
				Maximum:    1500,
			},
		},
		FlexibleExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
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

	solver := BuildMetaGPFromConstraintModel(model, params)

	assert.NotNil(t, solver)
	assert.Greater(t, len(solver.variables), 0)
	assert.Greater(t, len(solver.goals), 0)

	t.Logf("Built Meta GP solver with %d variables and %d goals", len(solver.variables), len(solver.goals))

	result, err := solver.Solve()
	assert.NoError(t, err)

	t.Logf("Total reward: %.1f / %.1f (%.1f%%)", result.TotalReward, result.MaxPossibleReward, result.RewardRatio*100)
	t.Logf("Achieved goals: %d", len(result.AchievedGoals))
	t.Logf("Ideal goals: %d", len(result.IdealGoals))

	for goalID, level := range result.GoalLevels {
		t.Logf("  %s: %s ($%.2f)", goalID, level, result.GoalValues[goalID])
	}
}

func TestMetaGPSolver_EmptyInput(t *testing.T) {
	solver := NewMetaGPSolver(1000)

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.VariableValues)
}

func TestMetaGPSolver_ZeroBudget(t *testing.T) {
	solver := NewMetaGPSolver(0)

	id := uuid.New()
	idx := solver.AddVariable(MetaVariable{ID: id, Name: "Test", MinValue: 0, MaxValue: 100})

	solver.AddGoal(MetaGoal{
		ID:          "test",
		VariableIdx: idx,
		Priority:    1,
		TargetLevels: []TargetLevel{
			{Level: "minimum", Value: 50, Reward: 50},
			{Level: "ideal", Value: 100, Reward: 100},
		},
	})

	result, err := solver.Solve()

	assert.NoError(t, err)
	assert.Equal(t, 0.0, result.VariableValues[id])
	assert.Equal(t, "none", result.GoalLevels["test"])
}

// Note: Integration tests with GoalProgrammingSolver should be in the parent solver package
// since NewGoalProgrammingSolver and SolveMeta are defined there.
