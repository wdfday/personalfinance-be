package ahp

import (
	"context"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAHPModel_Name(t *testing.T) {
	model := NewAHPModel()
	assert.Equal(t, "goal_prioritization", model.Name())
}

func TestAHPModel_Description(t *testing.T) {
	model := NewAHPModel()
	assert.Equal(t, "Analytic Hierarchy Process for prioritizing financial goals", model.Description())
}

func TestAHPModel_Dependencies(t *testing.T) {
	model := NewAHPModel()
	deps := model.Dependencies()
	assert.Empty(t, deps)
}

func TestAHPModel_Validate(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid input type",
			input:       "invalid",
			expectError: true,
			errorMsg:    "input must be *dto.AHPInput type",
		},
		{
			name: "Less than 2 criteria",
			input: &dto.AHPInput{
				Criteria:               []domain.Criteria{{ID: "c1"}},
				Alternatives:           []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons:    []domain.PairwiseComparison{},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{},
			},
			expectError: true,
			errorMsg:    "at least 2 criteria required",
		},
		{
			name: "Less than 2 alternatives",
			input: &dto.AHPInput{
				Criteria:               []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives:           []domain.Alternative{{ID: "a1"}},
				CriteriaComparisons:    []domain.PairwiseComparison{},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{},
			},
			expectError: true,
			errorMsg:    "at least 2 alternatives required",
		},
		{
			name: "Incorrect number of criteria comparisons",
			input: &dto.AHPInput{
				Criteria:            []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives:        []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					// Expected 1 comparison for 2 criteria, but providing 0
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{},
			},
			expectError: true,
			errorMsg:    "criteria comparisons: expected 1, got 0",
		},
		{
			name: "Missing alternative comparisons for criterion",
			input: &dto.AHPInput{
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 3.0},
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					// Missing comparisons for both c1 and c2
				},
			},
			expectError: true,
			errorMsg:    "missing alternative comparisons for criterion: c1",
		},
		{
			name: "Incorrect number of alternative comparisons",
			input: &dto.AHPInput{
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 3.0},
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {}, // Expected 1 comparison for 2 alternatives, but providing 0
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
				},
			},
			expectError: true,
			errorMsg:    "criterion c1: expected 1 comparisons, got 0",
		},
		{
			name: "Comparison value out of range - too high",
			input: &dto.AHPInput{
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 10.0}, // Out of range
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
				},
			},
			expectError: true,
			errorMsg:    "comparison value out of range: 10.000000",
		},
		{
			name: "Comparison value out of range - too low",
			input: &dto.AHPInput{
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 0.1}, // Out of range (< 1/9)
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
				},
			},
			expectError: true,
			errorMsg:    "comparison value out of range: 0.100000",
		},
		{
			name: "Valid input - minimum case",
			input: &dto.AHPInput{
				UserID:       "user1",
				Criteria:     []domain.Criteria{{ID: "c1"}, {ID: "c2"}},
				Alternatives: []domain.Alternative{{ID: "a1"}, {ID: "a2"}},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 3.0},
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {{ElementA: "a1", ElementB: "a2", Value: 2.0}},
					"c2": {{ElementA: "a1", ElementB: "a2", Value: 0.5}},
				},
			},
			expectError: false,
		},
		{
			name: "Valid input - 3 criteria, 3 alternatives",
			input: &dto.AHPInput{
				UserID: "user1",
				Criteria: []domain.Criteria{
					{ID: "c1"}, {ID: "c2"}, {ID: "c3"},
				},
				Alternatives: []domain.Alternative{
					{ID: "a1"}, {ID: "a2"}, {ID: "a3"},
				},
				CriteriaComparisons: []domain.PairwiseComparison{
					{ElementA: "c1", ElementB: "c2", Value: 3.0},
					{ElementA: "c1", ElementB: "c3", Value: 5.0},
					{ElementA: "c2", ElementB: "c3", Value: 2.0},
				},
				AlternativeComparisons: map[string][]domain.PairwiseComparison{
					"c1": {
						{ElementA: "a1", ElementB: "a2", Value: 2.0},
						{ElementA: "a1", ElementB: "a3", Value: 4.0},
						{ElementA: "a2", ElementB: "a3", Value: 2.0},
					},
					"c2": {
						{ElementA: "a1", ElementB: "a2", Value: 1.0},
						{ElementA: "a1", ElementB: "a3", Value: 3.0},
						{ElementA: "a2", ElementB: "a3", Value: 3.0},
					},
					"c3": {
						{ElementA: "a1", ElementB: "a2", Value: 0.5},
						{ElementA: "a1", ElementB: "a3", Value: 2.0},
						{ElementA: "a2", ElementB: "a3", Value: 4.0},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.Validate(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAHPModel_BuildMatrix(t *testing.T) {
	model := NewAHPModel()

	t.Run("Build 2x2 matrix", func(t *testing.T) {
		ids := []string{"a", "b"}
		comparisons := []domain.PairwiseComparison{
			{ElementA: "a", ElementB: "b", Value: 3.0},
		}

		matrix := model.buildMatrix(ids, comparisons)

		assert.Equal(t, 2, matrix.Size)
		assert.Equal(t, ids, matrix.Elements)
		assert.Equal(t, 1.0, matrix.Matrix[0][0])     // Diagonal
		assert.Equal(t, 1.0, matrix.Matrix[1][1])     // Diagonal
		assert.Equal(t, 3.0, matrix.Matrix[0][1])     // a > b
		assert.Equal(t, 1.0/3.0, matrix.Matrix[1][0]) // b < a (reciprocal)
	})

	t.Run("Build 3x3 matrix", func(t *testing.T) {
		ids := []string{"a", "b", "c"}
		comparisons := []domain.PairwiseComparison{
			{ElementA: "a", ElementB: "b", Value: 2.0},
			{ElementA: "a", ElementB: "c", Value: 5.0},
			{ElementA: "b", ElementB: "c", Value: 3.0},
		}

		matrix := model.buildMatrix(ids, comparisons)

		assert.Equal(t, 3, matrix.Size)
		assert.Equal(t, ids, matrix.Elements)

		// Check diagonal
		assert.Equal(t, 1.0, matrix.Matrix[0][0])
		assert.Equal(t, 1.0, matrix.Matrix[1][1])
		assert.Equal(t, 1.0, matrix.Matrix[2][2])

		// Check comparisons
		assert.Equal(t, 2.0, matrix.Matrix[0][1])
		assert.Equal(t, 0.5, matrix.Matrix[1][0])

		assert.Equal(t, 5.0, matrix.Matrix[0][2])
		assert.Equal(t, 0.2, matrix.Matrix[2][0])

		assert.Equal(t, 3.0, matrix.Matrix[1][2])
		assert.InDelta(t, 1.0/3.0, matrix.Matrix[2][1], 0.0001)
	})
}

func TestAHPModel_CalculatePriorities(t *testing.T) {
	model := NewAHPModel()

	t.Run("Perfect consistency matrix", func(t *testing.T) {
		// Matrix where a:b:c = 4:2:1
		matrix := &domain.ComparisonMatrix{
			Size: 3,
			Matrix: [][]float64{
				{1.0, 2.0, 4.0},
				{0.5, 1.0, 2.0},
				{0.25, 0.5, 1.0},
			},
			Elements: []string{"a", "b", "c"},
		}

		priorities, cr, err := model.calculatePriorities(matrix)

		require.NoError(t, err)
		assert.InDelta(t, 0.571, priorities["a"], 0.01) // ~4/7
		assert.InDelta(t, 0.286, priorities["b"], 0.01) // ~2/7
		assert.InDelta(t, 0.143, priorities["c"], 0.01) // ~1/7

		// Sum should be 1.0
		sum := priorities["a"] + priorities["b"] + priorities["c"]
		assert.InDelta(t, 1.0, sum, 0.001)

		// CR should be very low for consistent matrix
		assert.Less(t, cr, 0.1)
	})

	t.Run("2x2 matrix", func(t *testing.T) {
		matrix := &domain.ComparisonMatrix{
			Size: 2,
			Matrix: [][]float64{
				{1.0, 3.0},
				{1.0 / 3.0, 1.0},
			},
			Elements: []string{"a", "b"},
		}

		priorities, cr, err := model.calculatePriorities(matrix)

		require.NoError(t, err)
		assert.InDelta(t, 0.75, priorities["a"], 0.01)
		assert.InDelta(t, 0.25, priorities["b"], 0.01)

		// For 2x2, CR should be 0
		assert.Equal(t, 0.0, cr)
	})
}

func TestAHPModel_GetRandomIndex(t *testing.T) {
	model := NewAHPModel()

	tests := []struct {
		n        int
		expected float64
	}{
		{1, 0.00},
		{2, 0.00},
		{3, 0.58},
		{4, 0.90},
		{5, 1.12},
		{6, 1.24},
		{7, 1.32},
		{8, 1.41},
		{9, 1.45},
		{10, 1.49},
		{11, 1.50}, // 1.49 + 0.01 * (11-10)
		{15, 1.54}, // 1.49 + 0.01 * (15-10)
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.n)), func(t *testing.T) {
			ri := model.getRandomIndex(tt.n)
			assert.InDelta(t, tt.expected, ri, 0.001)
		})
	}
}

func TestAHPModel_CreateRanking(t *testing.T) {
	model := NewAHPModel()

	alternatives := []domain.Alternative{
		{ID: "goal1", Name: "Emergency Fund"},
		{ID: "goal2", Name: "House Down Payment"},
		{ID: "goal3", Name: "Retirement"},
	}

	priorities := map[string]float64{
		"goal1": 0.5,
		"goal2": 0.3,
		"goal3": 0.2,
	}

	ranking := model.createRanking(alternatives, priorities)

	assert.Len(t, ranking, 3)

	// Check sorting (highest priority first)
	assert.Equal(t, "goal1", ranking[0].AlternativeID)
	assert.Equal(t, "Emergency Fund", ranking[0].AlternativeName)
	assert.Equal(t, 0.5, ranking[0].Priority)
	assert.Equal(t, 1, ranking[0].Rank)

	assert.Equal(t, "goal2", ranking[1].AlternativeID)
	assert.Equal(t, 2, ranking[1].Rank)

	assert.Equal(t, "goal3", ranking[2].AlternativeID)
	assert.Equal(t, 3, ranking[2].Rank)
}

func TestAHPModel_Execute(t *testing.T) {
	model := NewAHPModel()
	ctx := context.Background()

	t.Run("Full execution - simple case", func(t *testing.T) {
		input := &dto.AHPInput{
			UserID: "user1",
			Criteria: []domain.Criteria{
				{ID: "importance", Name: "Importance"},
				{ID: "urgency", Name: "Urgency"},
			},
			CriteriaComparisons: []domain.PairwiseComparison{
				{ElementA: "importance", ElementB: "urgency", Value: 2.0},
			},
			Alternatives: []domain.Alternative{
				{ID: "goal1", Name: "Emergency Fund"},
				{ID: "goal2", Name: "Vacation"},
			},
			AlternativeComparisons: map[string][]domain.PairwiseComparison{
				"importance": {
					{ElementA: "goal1", ElementB: "goal2", Value: 5.0},
				},
				"urgency": {
					{ElementA: "goal1", ElementB: "goal2", Value: 3.0},
				},
			},
		}

		t.Logf("\n========== AHP EXECUTION: Simple Case ==========")
		t.Logf("\n📋 INPUT:")
		t.Logf("  Criteria:")
		for _, c := range input.Criteria {
			t.Logf("    • %s", c.Name)
		}
		t.Logf("  Alternatives (Goals):")
		for _, a := range input.Alternatives {
			t.Logf("    • %s", a.Name)
		}
		t.Logf("  Pairwise Comparisons:")
		t.Logf("    Criteria: Importance is 2x more important than Urgency")
		t.Logf("    For Importance: Emergency Fund is 5x better than Vacation")
		t.Logf("    For Urgency: Emergency Fund is 3x more urgent than Vacation")

		result, err := model.Execute(ctx, input)

		require.NoError(t, err)
		output := result.(*dto.AHPOutput)

		t.Logf("\n📊 OUTPUT:")
		t.Logf("  Criteria Weights:")
		for id, weight := range output.CriteriaWeights {
			t.Logf("    • %-15s: %.4f (%.1f%%)", id, weight, weight*100)
		}
		t.Logf("  \n  Final Goal Priorities:")
		for _, item := range output.Ranking {
			t.Logf("    #%d: %-20s %.4f (%.1f%%)", item.Rank, item.AlternativeName, item.Priority, item.Priority*100)
		}
		t.Logf("  \n  Consistency Check:")
		t.Logf("    • Consistency Ratio: %.4f", output.ConsistencyRatio)
		t.Logf("    • Is Consistent: %v (threshold: CR < 0.10)", output.IsConsistent)
		t.Logf("\n================================================\n")

		// Check criteria weights
		assert.NotNil(t, output.CriteriaWeights)
		assert.Len(t, output.CriteriaWeights, 2)

		importanceWeight := output.CriteriaWeights["importance"]
		urgencyWeight := output.CriteriaWeights["urgency"]

		// importance should be higher (2x)
		assert.Greater(t, importanceWeight, urgencyWeight)

		// Weights should sum to 1
		assert.InDelta(t, 1.0, importanceWeight+urgencyWeight, 0.001)

		// Check alternative priorities
		assert.NotNil(t, output.AlternativePriorities)
		assert.Len(t, output.AlternativePriorities, 2)

		goal1Priority := output.AlternativePriorities["goal1"]
		goal2Priority := output.AlternativePriorities["goal2"]

		// goal1 should be higher
		assert.Greater(t, goal1Priority, goal2Priority)

		// Priorities should sum to 1
		assert.InDelta(t, 1.0, goal1Priority+goal2Priority, 0.001)

		// Check local priorities
		assert.NotNil(t, output.LocalPriorities)
		assert.Len(t, output.LocalPriorities, 2)

		// Check ranking
		assert.NotNil(t, output.Ranking)
		assert.Len(t, output.Ranking, 2)
		assert.Equal(t, "goal1", output.Ranking[0].AlternativeID)
		assert.Equal(t, 1, output.Ranking[0].Rank)
		assert.Equal(t, "goal2", output.Ranking[1].AlternativeID)
		assert.Equal(t, 2, output.Ranking[1].Rank)

		// Check consistency
		assert.Less(t, output.ConsistencyRatio, 0.1)
		assert.True(t, output.IsConsistent)
	})

	t.Run("Full execution - 3 criteria, 3 alternatives", func(t *testing.T) {
		input := &dto.AHPInput{
			UserID: "user1",
			Criteria: []domain.Criteria{
				{ID: "c1", Name: "Impact"},
				{ID: "c2", Name: "Urgency"},
				{ID: "c3", Name: "Cost"},
			},
			CriteriaComparisons: []domain.PairwiseComparison{
				{ElementA: "c1", ElementB: "c2", Value: 3.0},
				{ElementA: "c1", ElementB: "c3", Value: 5.0},
				{ElementA: "c2", ElementB: "c3", Value: 2.0},
			},
			Alternatives: []domain.Alternative{
				{ID: "a1", Name: "Goal A"},
				{ID: "a2", Name: "Goal B"},
				{ID: "a3", Name: "Goal C"},
			},
			AlternativeComparisons: map[string][]domain.PairwiseComparison{
				"c1": {
					{ElementA: "a1", ElementB: "a2", Value: 2.0},
					{ElementA: "a1", ElementB: "a3", Value: 4.0},
					{ElementA: "a2", ElementB: "a3", Value: 2.0},
				},
				"c2": {
					{ElementA: "a1", ElementB: "a2", Value: 1.0},
					{ElementA: "a1", ElementB: "a3", Value: 3.0},
					{ElementA: "a2", ElementB: "a3", Value: 3.0},
				},
				"c3": {
					{ElementA: "a1", ElementB: "a2", Value: 0.5},
					{ElementA: "a1", ElementB: "a3", Value: 2.0},
					{ElementA: "a2", ElementB: "a3", Value: 4.0},
				},
			},
		}

		t.Logf("\n========== AHP EXECUTION: Complex 3x3 Case ==========")
		t.Logf("\n📋 INPUT:")
		t.Logf("  Criteria (3):")
		for i, c := range input.Criteria {
			t.Logf("    %d. %s", i+1, c.Name)
		}
		t.Logf("  Alternatives (3):")
		for i, a := range input.Alternatives {
			t.Logf("    %d. %s", i+1, a.Name)
		}
		t.Logf("  Criteria Comparisons:")
		t.Logf("    • Impact vs Urgency: 3.0 (Impact is 3x more important)")
		t.Logf("    • Impact vs Cost: 5.0 (Impact is 5x more important)")
		t.Logf("    • Urgency vs Cost: 2.0 (Urgency is 2x more important)")

		result, err := model.Execute(ctx, input)

		require.NoError(t, err)
		output := result.(*dto.AHPOutput)

		t.Logf("\n📊 OUTPUT:")
		t.Logf("  Criteria Weights:")
		criteriaNames := map[string]string{"c1": "Impact", "c2": "Urgency", "c3": "Cost"}
		for _, c := range input.Criteria {
			weight := output.CriteriaWeights[c.ID]
			t.Logf("    • %-10s: %.4f (%.1f%%)", criteriaNames[c.ID], weight, weight*100)
		}
		t.Logf("\n  Local Priorities by Criterion:")
		for _, c := range input.Criteria {
			t.Logf("    %s:", criteriaNames[c.ID])
			for _, a := range input.Alternatives {
				localPriority := output.LocalPriorities[c.ID][a.ID]
				t.Logf("      • %-10s: %.4f", a.Name, localPriority)
			}
		}
		t.Logf("\n  🏆 FINAL RANKING (Global Priorities):")
		for _, item := range output.Ranking {
			bar := ""
			barLength := int(item.Priority * 50)
			for i := 0; i < barLength; i++ {
				bar += "█"
			}
			t.Logf("    #%d %-10s %.4f (%.1f%%) %s",
				item.Rank, item.AlternativeName, item.Priority, item.Priority*100, bar)
		}
		t.Logf("\n  Consistency Check:")
		t.Logf("    • Consistency Ratio: %.4f", output.ConsistencyRatio)
		t.Logf("    • Status: %v (CR < 0.10 = consistent)", output.IsConsistent)
		t.Logf("\n====================================================\n")

		// Verify output structure
		assert.NotNil(t, output.CriteriaWeights)
		assert.Len(t, output.CriteriaWeights, 3)

		assert.NotNil(t, output.AlternativePriorities)
		assert.Len(t, output.AlternativePriorities, 3)

		assert.NotNil(t, output.LocalPriorities)
		assert.Len(t, output.LocalPriorities, 3)

		assert.NotNil(t, output.Ranking)
		assert.Len(t, output.Ranking, 3)

		// Verify all weights sum to 1
		criteriaSum := output.CriteriaWeights["c1"] +
			output.CriteriaWeights["c2"] +
			output.CriteriaWeights["c3"]
		assert.InDelta(t, 1.0, criteriaSum, 0.001)

		alternativeSum := output.AlternativePriorities["a1"] +
			output.AlternativePriorities["a2"] +
			output.AlternativePriorities["a3"]
		assert.InDelta(t, 1.0, alternativeSum, 0.001)

		// Verify ranking order (ranks should be 1, 2, 3)
		ranks := make(map[int]bool)
		for _, item := range output.Ranking {
			ranks[item.Rank] = true
		}
		assert.True(t, ranks[1])
		assert.True(t, ranks[2])
		assert.True(t, ranks[3])
	})

	t.Run("Execute with invalid input type", func(t *testing.T) {
		// Should panic because we don't check type in Execute
		// (assuming Validate was called first)
		assert.Panics(t, func() {
			model.Execute(ctx, "invalid")
		})
	})
}

func TestAHPModel_BuildComparisonMatrixForCriteria(t *testing.T) {
	model := NewAHPModel()

	criteria := []domain.Criteria{
		{ID: "c1", Name: "Criteria 1"},
		{ID: "c2", Name: "Criteria 2"},
	}

	comparisons := []domain.PairwiseComparison{
		{ElementA: "c1", ElementB: "c2", Value: 3.0},
	}

	matrix := model.buildComparisonMatrixForCriteria(criteria, comparisons)

	assert.Equal(t, 2, matrix.Size)
	assert.Equal(t, []string{"c1", "c2"}, matrix.Elements)
	assert.Equal(t, 3.0, matrix.Matrix[0][1])
	assert.InDelta(t, 1.0/3.0, matrix.Matrix[1][0], 0.0001)
}

func TestAHPModel_BuildComparisonMatrixForAlternatives(t *testing.T) {
	model := NewAHPModel()

	alternatives := []domain.Alternative{
		{ID: "a1", Name: "Alternative 1"},
		{ID: "a2", Name: "Alternative 2"},
	}

	comparisons := []domain.PairwiseComparison{
		{ElementA: "a1", ElementB: "a2", Value: 2.0},
	}

	matrix := model.buildComparisonMatrixForAlternatives(alternatives, comparisons)

	assert.Equal(t, 2, matrix.Size)
	assert.Equal(t, []string{"a1", "a2"}, matrix.Elements)
	assert.Equal(t, 2.0, matrix.Matrix[0][1])
	assert.Equal(t, 0.5, matrix.Matrix[1][0])
}

func TestAHPModel_CalculateConsistencyRatio(t *testing.T) {
	model := NewAHPModel()

	t.Run("Perfect consistency - 2x2 matrix", func(t *testing.T) {
		matrix := [][]float64{
			{1.0, 3.0},
			{1.0 / 3.0, 1.0},
		}
		priorityVector := []float64{0.75, 0.25}

		cr := model.calculateConsistencyRatio(matrix, priorityVector)

		// For 2x2, CR should always be 0 (RI = 0)
		assert.Equal(t, 0.0, cr)
	})

	t.Run("Consistent 3x3 matrix", func(t *testing.T) {
		// Perfectly consistent matrix
		matrix := [][]float64{
			{1.0, 2.0, 4.0},
			{0.5, 1.0, 2.0},
			{0.25, 0.5, 1.0},
		}
		priorityVector := []float64{0.571, 0.286, 0.143}

		cr := model.calculateConsistencyRatio(matrix, priorityVector)

		// Should be very low for consistent matrix
		assert.Less(t, cr, 0.05)
	})
}
