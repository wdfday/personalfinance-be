package ahp

import (
	"math"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"sort"
)

// SensitivityResult contains results from AHP sensitivity analysis
type SensitivityResult struct {
	// CriteriaSensitivity shows how changes in criteria weights affect ranking
	CriteriaSensitivity []CriteriaSensitivityItem `json:"criteria_sensitivity"`

	// ComparisonSensitivity shows which comparisons are most critical
	ComparisonSensitivity []ComparisonSensitivityItem `json:"comparison_sensitivity"`

	// RankingStability indicates how stable the current ranking is
	RankingStability RankingStabilityResult `json:"ranking_stability"`

	// CriticalThresholds shows at what weight changes the ranking would change
	CriticalThresholds []CriticalThreshold `json:"critical_thresholds"`
}

// CriteriaSensitivityItem shows sensitivity for one criterion
type CriteriaSensitivityItem struct {
	CriterionID   string  `json:"criterion_id"`
	CriterionName string  `json:"criterion_name"`
	CurrentWeight float64 `json:"current_weight"`

	// Impact on top alternative if weight increases by 10%
	ImpactIfIncrease float64 `json:"impact_if_increase"`
	// Impact on top alternative if weight decreases by 10%
	ImpactIfDecrease float64 `json:"impact_if_decrease"`

	// Sensitivity score (higher = more sensitive)
	SensitivityScore float64 `json:"sensitivity_score"`
	SensitivityLevel string  `json:"sensitivity_level"` // "high", "medium", "low"
}

// ComparisonSensitivityItem shows sensitivity for one comparison
type ComparisonSensitivityItem struct {
	ElementA     string  `json:"element_a"`
	ElementB     string  `json:"element_b"`
	CurrentValue float64 `json:"current_value"`

	// How much the ranking changes if this comparison changes by ±1
	RankingImpact    float64 `json:"ranking_impact"`
	SensitivityLevel string  `json:"sensitivity_level"`
}

// RankingStabilityResult indicates overall ranking stability
type RankingStabilityResult struct {
	IsStable        bool    `json:"is_stable"`
	StabilityScore  float64 `json:"stability_score"`   // 0-100
	MinWeightChange float64 `json:"min_weight_change"` // Minimum change to flip ranking
	TopTwoGap       float64 `json:"top_two_gap"`       // Gap between #1 and #2
	Recommendation  string  `json:"recommendation"`
}

// CriticalThreshold shows when ranking would change
type CriticalThreshold struct {
	CriterionID     string  `json:"criterion_id"`
	CurrentWeight   float64 `json:"current_weight"`
	ThresholdWeight float64 `json:"threshold_weight"` // Weight at which ranking changes
	ChangeDirection string  `json:"change_direction"` // "increase" or "decrease"
	AffectedRanking string  `json:"affected_ranking"` // e.g., "A overtakes B"
}

// AnalyzeSensitivity performs sensitivity analysis on AHP results
func (m *AHPModel) AnalyzeSensitivity(
	input *dto.AHPInput,
	output *dto.AHPOutput,
) *SensitivityResult {
	result := &SensitivityResult{
		CriteriaSensitivity:   make([]CriteriaSensitivityItem, 0),
		ComparisonSensitivity: make([]ComparisonSensitivityItem, 0),
		CriticalThresholds:    make([]CriticalThreshold, 0),
	}

	// 1. Analyze criteria weight sensitivity
	result.CriteriaSensitivity = m.analyzeCriteriaSensitivity(input, output)

	// 2. Analyze comparison sensitivity
	result.ComparisonSensitivity = m.analyzeComparisonSensitivity(input, output)

	// 3. Calculate ranking stability
	result.RankingStability = m.calculateRankingStability(output)

	// 4. Find critical thresholds
	result.CriticalThresholds = m.findCriticalThresholds(input, output)

	return result
}

// analyzeCriteriaSensitivity analyzes how criteria weight changes affect ranking
func (m *AHPModel) analyzeCriteriaSensitivity(
	input *dto.AHPInput,
	output *dto.AHPOutput,
) []CriteriaSensitivityItem {
	items := make([]CriteriaSensitivityItem, 0, len(input.Criteria))

	// Get top alternative
	topAltID := output.Ranking[0].AlternativeID

	for _, criterion := range input.Criteria {
		currentWeight := output.CriteriaWeights[criterion.ID]

		// Simulate +10% weight change
		impactIncrease := m.simulateWeightChange(input, output, criterion.ID, 0.10, topAltID)

		// Simulate -10% weight change
		impactDecrease := m.simulateWeightChange(input, output, criterion.ID, -0.10, topAltID)

		// Calculate sensitivity score
		sensitivityScore := math.Abs(impactIncrease) + math.Abs(impactDecrease)

		// Determine sensitivity level
		level := "low"
		if sensitivityScore > 0.10 {
			level = "high"
		} else if sensitivityScore > 0.05 {
			level = "medium"
		}

		items = append(items, CriteriaSensitivityItem{
			CriterionID:      criterion.ID,
			CriterionName:    criterion.Name,
			CurrentWeight:    currentWeight,
			ImpactIfIncrease: impactIncrease,
			ImpactIfDecrease: impactDecrease,
			SensitivityScore: sensitivityScore,
			SensitivityLevel: level,
		})
	}

	// Sort by sensitivity score descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].SensitivityScore > items[j].SensitivityScore
	})

	return items
}

// simulateWeightChange simulates changing a criterion weight and returns impact on top alternative
func (m *AHPModel) simulateWeightChange(
	input *dto.AHPInput,
	output *dto.AHPOutput,
	criterionID string,
	changePercent float64,
	topAltID string,
) float64 {
	// Create modified weights
	modifiedWeights := make(map[string]float64)
	totalOtherWeights := 0.0

	for cID, weight := range output.CriteriaWeights {
		if cID == criterionID {
			modifiedWeights[cID] = weight * (1 + changePercent)
		} else {
			modifiedWeights[cID] = weight
			totalOtherWeights += weight
		}
	}

	// Normalize weights to sum to 1
	newTotal := 0.0
	for _, w := range modifiedWeights {
		newTotal += w
	}
	for cID := range modifiedWeights {
		modifiedWeights[cID] /= newTotal
	}

	// Recalculate global priorities with modified weights
	newPriority := 0.0
	for _, criterion := range input.Criteria {
		newPriority += modifiedWeights[criterion.ID] * output.LocalPriorities[criterion.ID][topAltID]
	}

	// Return change in priority
	return newPriority - output.AlternativePriorities[topAltID]
}

// analyzeComparisonSensitivity analyzes which comparisons are most critical
func (m *AHPModel) analyzeComparisonSensitivity(
	input *dto.AHPInput,
	output *dto.AHPOutput,
) []ComparisonSensitivityItem {
	items := make([]ComparisonSensitivityItem, 0)

	// Analyze criteria comparisons
	for _, comp := range input.CriteriaComparisons {
		impact := m.calculateComparisonImpact(comp.Value)
		level := "low"
		if impact > 0.15 {
			level = "high"
		} else if impact > 0.08 {
			level = "medium"
		}

		items = append(items, ComparisonSensitivityItem{
			ElementA:         comp.ElementA,
			ElementB:         comp.ElementB,
			CurrentValue:     comp.Value,
			RankingImpact:    impact,
			SensitivityLevel: level,
		})
	}

	// Sort by impact descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].RankingImpact > items[j].RankingImpact
	})

	return items
}

// calculateComparisonImpact estimates impact of a comparison value
func (m *AHPModel) calculateComparisonImpact(value float64) float64 {
	// Higher values (further from 1) have more impact
	if value >= 1 {
		return (value - 1) / 8.0 // Normalize to 0-1 range
	}
	return (1/value - 1) / 8.0
}

// calculateRankingStability calculates overall ranking stability
func (m *AHPModel) calculateRankingStability(output *dto.AHPOutput) RankingStabilityResult {
	result := RankingStabilityResult{}

	if len(output.Ranking) < 2 {
		result.IsStable = true
		result.StabilityScore = 100
		result.Recommendation = "Only one alternative - ranking is trivially stable"
		return result
	}

	// Calculate gap between top two
	result.TopTwoGap = output.Ranking[0].Priority - output.Ranking[1].Priority

	// Estimate minimum weight change to flip ranking
	// Simplified: if gap is large, need large change
	result.MinWeightChange = result.TopTwoGap * 2

	// Calculate stability score (0-100)
	// Higher gap = more stable
	result.StabilityScore = math.Min(100, result.TopTwoGap*500)

	// Determine stability
	if result.TopTwoGap > 0.15 {
		result.IsStable = true
		result.Recommendation = "Ranking is highly stable. Top choice is clearly dominant."
	} else if result.TopTwoGap > 0.05 {
		result.IsStable = true
		result.Recommendation = "Ranking is moderately stable. Consider reviewing close alternatives."
	} else {
		result.IsStable = false
		result.Recommendation = "Ranking is sensitive. Small changes in judgments could flip the top choice."
	}

	return result
}

// findCriticalThresholds finds weight values where ranking would change
func (m *AHPModel) findCriticalThresholds(
	input *dto.AHPInput,
	output *dto.AHPOutput,
) []CriticalThreshold {
	thresholds := make([]CriticalThreshold, 0)

	if len(output.Ranking) < 2 {
		return thresholds
	}

	topAlt := output.Ranking[0]
	secondAlt := output.Ranking[1]

	// For each criterion, find threshold where #2 would overtake #1
	for _, criterion := range input.Criteria {
		currentWeight := output.CriteriaWeights[criterion.ID]

		// Local priorities for top two alternatives under this criterion
		topLocal := output.LocalPriorities[criterion.ID][topAlt.AlternativeID]
		secondLocal := output.LocalPriorities[criterion.ID][secondAlt.AlternativeID]

		// If second alternative is better under this criterion, increasing weight helps it
		if secondLocal > topLocal {
			// Calculate threshold (simplified)
			gap := topAlt.Priority - secondAlt.Priority
			localGap := secondLocal - topLocal

			if localGap > 0 {
				// Threshold = current + (gap / localGap)
				threshold := currentWeight + (gap / localGap)
				if threshold > 0 && threshold < 1 {
					thresholds = append(thresholds, CriticalThreshold{
						CriterionID:     criterion.ID,
						CurrentWeight:   currentWeight,
						ThresholdWeight: threshold,
						ChangeDirection: "increase",
						AffectedRanking: secondAlt.AlternativeName + " overtakes " + topAlt.AlternativeName,
					})
				}
			}
		}
	}

	return thresholds
}
