package dto

import (
	"testing"

	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"

	"github.com/stretchr/testify/assert"
)

func TestToAHPOutput(t *testing.T) {
	tests := []struct {
		name               string
		altPriorities      map[string]float64
		criteriaWeights    map[string]float64
		localPriorities    map[string]map[string]float64
		cr                 float64
		ranking            []domain.RankItem
		expectedConsistent bool
	}{
		{
			name: "Consistent result with CR < 0.10",
			altPriorities: map[string]float64{
				"goal1": 0.45,
				"goal2": 0.35,
				"goal3": 0.20,
			},
			criteriaWeights: map[string]float64{
				"criteria1": 0.6,
				"criteria2": 0.4,
			},
			localPriorities: map[string]map[string]float64{
				"criteria1": {
					"goal1": 0.5,
					"goal2": 0.3,
					"goal3": 0.2,
				},
				"criteria2": {
					"goal1": 0.4,
					"goal2": 0.4,
					"goal3": 0.2,
				},
			},
			cr: 0.05,
			ranking: []domain.RankItem{
				{AlternativeID: "goal1", AlternativeName: "Goal 1", Priority: 0.45, Rank: 1},
				{AlternativeID: "goal2", AlternativeName: "Goal 2", Priority: 0.35, Rank: 2},
				{AlternativeID: "goal3", AlternativeName: "Goal 3", Priority: 0.20, Rank: 3},
			},
			expectedConsistent: true,
		},
		{
			name: "Inconsistent result with CR >= 0.10",
			altPriorities: map[string]float64{
				"goal1": 0.50,
				"goal2": 0.30,
				"goal3": 0.20,
			},
			criteriaWeights: map[string]float64{
				"criteria1": 0.7,
				"criteria2": 0.3,
			},
			localPriorities: map[string]map[string]float64{
				"criteria1": {
					"goal1": 0.6,
					"goal2": 0.25,
					"goal3": 0.15,
				},
			},
			cr: 0.15,
			ranking: []domain.RankItem{
				{AlternativeID: "goal1", AlternativeName: "Goal 1", Priority: 0.50, Rank: 1},
			},
			expectedConsistent: false,
		},
		{
			name: "Edge case with CR exactly 0.10",
			altPriorities: map[string]float64{
				"goal1": 0.50,
			},
			criteriaWeights: map[string]float64{
				"criteria1": 1.0,
			},
			localPriorities: map[string]map[string]float64{},
			cr:              0.10,
			ranking: []domain.RankItem{
				{AlternativeID: "goal1", AlternativeName: "Goal 1", Priority: 0.50, Rank: 1},
			},
			expectedConsistent: false,
		},
		{
			name:               "Zero CR - perfect consistency",
			altPriorities:      map[string]float64{"goal1": 1.0},
			criteriaWeights:    map[string]float64{"criteria1": 1.0},
			localPriorities:    map[string]map[string]float64{},
			cr:                 0.0,
			ranking:            []domain.RankItem{},
			expectedConsistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ToAHPOutput(
				tt.altPriorities,
				tt.criteriaWeights,
				tt.localPriorities,
				tt.cr,
				tt.ranking,
			)

			assert.NotNil(t, output)
			assert.Equal(t, tt.altPriorities, output.AlternativePriorities)
			assert.Equal(t, tt.criteriaWeights, output.CriteriaWeights)
			assert.Equal(t, tt.localPriorities, output.LocalPriorities)
			assert.Equal(t, tt.cr, output.ConsistencyRatio)
			assert.Equal(t, tt.expectedConsistent, output.IsConsistent)
			assert.Equal(t, tt.ranking, output.Ranking)
		})
	}
}

func TestToAHPOutput_NilMaps(t *testing.T) {
	output := ToAHPOutput(nil, nil, nil, 0.05, nil)

	assert.NotNil(t, output)
	assert.Nil(t, output.AlternativePriorities)
	assert.Nil(t, output.CriteriaWeights)
	assert.Nil(t, output.LocalPriorities)
	assert.Equal(t, 0.05, output.ConsistencyRatio)
	assert.True(t, output.IsConsistent)
	assert.Nil(t, output.Ranking)
}
