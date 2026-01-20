package ahp

import (
	"context"
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
)

// GroupDecisionInput represents input from multiple decision makers
type GroupDecisionInput struct {
	// Individual inputs from each decision maker
	DecisionMakers []DecisionMakerInput `json:"decision_makers"`

	// Aggregation method: "geometric_mean" (default), "arithmetic_mean", "weighted"
	AggregationMethod string `json:"aggregation_method"`

	// Weights for each decision maker (only used if method is "weighted")
	DecisionMakerWeights map[string]float64 `json:"decision_maker_weights,omitempty"`
}

// DecisionMakerInput represents one decision maker's input
type DecisionMakerInput struct {
	DecisionMakerID   string        `json:"decision_maker_id"`
	DecisionMakerName string        `json:"decision_maker_name"`
	Input             *dto.AHPInput `json:"input"`
	Weight            float64       `json:"weight,omitempty"` // Optional weight (expertise level)
}

// GroupDecisionOutput represents aggregated group decision result
type GroupDecisionOutput struct {
	// Aggregated AHP output
	AggregatedResult *dto.AHPOutput `json:"aggregated_result"`

	// Individual results for comparison
	IndividualResults []IndividualResult `json:"individual_results"`

	// Consensus metrics
	ConsensusMetrics ConsensusMetrics `json:"consensus_metrics"`

	// Disagreement analysis
	DisagreementAnalysis []DisagreementItem `json:"disagreement_analysis,omitempty"`
}

// IndividualResult contains one decision maker's result
type IndividualResult struct {
	DecisionMakerID   string         `json:"decision_maker_id"`
	DecisionMakerName string         `json:"decision_maker_name"`
	Output            *dto.AHPOutput `json:"output"`
	Weight            float64        `json:"weight"`
}

// ConsensusMetrics measures agreement among decision makers
type ConsensusMetrics struct {
	// Overall consensus index (0-1, higher = more agreement)
	ConsensusIndex float64 `json:"consensus_index"`

	// Criteria weight agreement
	CriteriaConsensus float64 `json:"criteria_consensus"`

	// Ranking agreement (Kendall's W or similar)
	RankingConsensus float64 `json:"ranking_consensus"`

	// Interpretation
	ConsensusLevel string `json:"consensus_level"` // "high", "medium", "low"
	Recommendation string `json:"recommendation"`
}

// DisagreementItem shows where decision makers disagree most
type DisagreementItem struct {
	Type              string   `json:"type"` // "criteria_weight", "alternative_ranking", "comparison"
	ElementID         string   `json:"element_id"`
	ElementName       string   `json:"element_name"`
	Variance          float64  `json:"variance"`
	MinValue          float64  `json:"min_value"`
	MaxValue          float64  `json:"max_value"`
	DisagreementLevel string   `json:"disagreement_level"`
	AffectedBy        []string `json:"affected_by"` // Decision maker IDs who differ most
}

// ExecuteGroupDecision aggregates multiple decision makers' inputs and produces group result
func (m *AHPModel) ExecuteGroupDecision(
	ctx context.Context,
	groupInput *GroupDecisionInput,
) (*GroupDecisionOutput, error) {
	if len(groupInput.DecisionMakers) == 0 {
		return nil, fmt.Errorf("at least one decision maker required")
	}

	if len(groupInput.DecisionMakers) == 1 {
		// Single decision maker - just execute normally
		result, err := m.Execute(ctx, groupInput.DecisionMakers[0].Input)
		if err != nil {
			return nil, err
		}
		return &GroupDecisionOutput{
			AggregatedResult: result.(*dto.AHPOutput),
			IndividualResults: []IndividualResult{{
				DecisionMakerID:   groupInput.DecisionMakers[0].DecisionMakerID,
				DecisionMakerName: groupInput.DecisionMakers[0].DecisionMakerName,
				Output:            result.(*dto.AHPOutput),
				Weight:            1.0,
			}},
			ConsensusMetrics: ConsensusMetrics{
				ConsensusIndex:    1.0,
				CriteriaConsensus: 1.0,
				RankingConsensus:  1.0,
				ConsensusLevel:    "high",
				Recommendation:    "Single decision maker - no consensus needed",
			},
		}, nil
	}

	// Execute AHP for each decision maker
	individualResults := make([]IndividualResult, 0, len(groupInput.DecisionMakers))
	for _, dm := range groupInput.DecisionMakers {
		if err := m.Validate(ctx, dm.Input); err != nil {
			return nil, fmt.Errorf("validation failed for %s: %w", dm.DecisionMakerID, err)
		}

		result, err := m.Execute(ctx, dm.Input)
		if err != nil {
			return nil, fmt.Errorf("execution failed for %s: %w", dm.DecisionMakerID, err)
		}

		weight := dm.Weight
		if weight == 0 {
			weight = 1.0
		}

		individualResults = append(individualResults, IndividualResult{
			DecisionMakerID:   dm.DecisionMakerID,
			DecisionMakerName: dm.DecisionMakerName,
			Output:            result.(*dto.AHPOutput),
			Weight:            weight,
		})
	}

	// Aggregate results
	aggregatedResult := m.aggregateResults(groupInput, individualResults)

	// Calculate consensus metrics
	consensusMetrics := m.calculateConsensusMetrics(individualResults)

	// Analyze disagreements
	disagreements := m.analyzeDisagreements(individualResults)

	return &GroupDecisionOutput{
		AggregatedResult:     aggregatedResult,
		IndividualResults:    individualResults,
		ConsensusMetrics:     consensusMetrics,
		DisagreementAnalysis: disagreements,
	}, nil
}

// aggregateResults combines individual results using specified method
func (m *AHPModel) aggregateResults(
	groupInput *GroupDecisionInput,
	results []IndividualResult,
) *dto.AHPOutput {
	method := groupInput.AggregationMethod
	if method == "" {
		method = "geometric_mean" // Default: geometric mean (recommended for AHP)
	}

	// Aggregate criteria weights
	aggregatedCriteriaWeights := m.aggregateWeights(results, method, "criteria")

	// Aggregate alternative priorities
	aggregatedAltPriorities := m.aggregateWeights(results, method, "alternatives")

	// Aggregate local priorities
	aggregatedLocalPriorities := m.aggregateLocalPriorities(results, method)

	// Calculate average consistency ratio
	avgCR := 0.0
	for _, r := range results {
		avgCR += r.Output.ConsistencyRatio
	}
	avgCR /= float64(len(results))

	// Create aggregated ranking
	ranking := m.createAggregatedRanking(results, aggregatedAltPriorities)

	return &dto.AHPOutput{
		AlternativePriorities: aggregatedAltPriorities,
		CriteriaWeights:       aggregatedCriteriaWeights,
		LocalPriorities:       aggregatedLocalPriorities,
		ConsistencyRatio:      avgCR,
		IsConsistent:          avgCR < 0.10,
		Ranking:               ranking,
	}
}

// aggregateWeights aggregates weights using specified method
func (m *AHPModel) aggregateWeights(
	results []IndividualResult,
	method string,
	weightType string,
) map[string]float64 {
	aggregated := make(map[string]float64)

	// Get all keys from first result
	var keys []string
	if weightType == "criteria" {
		for k := range results[0].Output.CriteriaWeights {
			keys = append(keys, k)
		}
	} else {
		for k := range results[0].Output.AlternativePriorities {
			keys = append(keys, k)
		}
	}

	for _, key := range keys {
		values := make([]float64, len(results))
		weights := make([]float64, len(results))

		for i, r := range results {
			if weightType == "criteria" {
				values[i] = r.Output.CriteriaWeights[key]
			} else {
				values[i] = r.Output.AlternativePriorities[key]
			}
			weights[i] = r.Weight
		}

		switch method {
		case "geometric_mean":
			aggregated[key] = geometricMean(values)
		case "arithmetic_mean":
			aggregated[key] = arithmeticMean(values)
		case "weighted":
			aggregated[key] = weightedMean(values, weights)
		default:
			aggregated[key] = geometricMean(values)
		}
	}

	// Normalize to sum to 1
	total := 0.0
	for _, v := range aggregated {
		total += v
	}
	for k := range aggregated {
		aggregated[k] /= total
	}

	return aggregated
}

// aggregateLocalPriorities aggregates local priorities for each criterion
func (m *AHPModel) aggregateLocalPriorities(
	results []IndividualResult,
	method string,
) map[string]map[string]float64 {
	aggregated := make(map[string]map[string]float64)

	// Get criteria IDs from first result
	for criterionID := range results[0].Output.LocalPriorities {
		aggregated[criterionID] = make(map[string]float64)

		// Get alternative IDs
		for altID := range results[0].Output.LocalPriorities[criterionID] {
			values := make([]float64, len(results))
			for i, r := range results {
				values[i] = r.Output.LocalPriorities[criterionID][altID]
			}

			switch method {
			case "geometric_mean":
				aggregated[criterionID][altID] = geometricMean(values)
			case "arithmetic_mean":
				aggregated[criterionID][altID] = arithmeticMean(values)
			default:
				aggregated[criterionID][altID] = geometricMean(values)
			}
		}

		// Normalize
		total := 0.0
		for _, v := range aggregated[criterionID] {
			total += v
		}
		for k := range aggregated[criterionID] {
			aggregated[criterionID][k] /= total
		}
	}

	return aggregated
}

// createAggregatedRanking creates ranking from aggregated priorities
func (m *AHPModel) createAggregatedRanking(
	results []IndividualResult,
	priorities map[string]float64,
) []domain.RankItem {
	// Get alternative names from first result
	altNames := make(map[string]string)
	for _, item := range results[0].Output.Ranking {
		altNames[item.AlternativeID] = item.AlternativeName
	}

	ranking := make([]domain.RankItem, 0, len(priorities))
	for id, priority := range priorities {
		ranking = append(ranking, domain.RankItem{
			AlternativeID:   id,
			AlternativeName: altNames[id],
			Priority:        priority,
		})
	}

	// Sort by priority descending
	for i := 0; i < len(ranking)-1; i++ {
		for j := i + 1; j < len(ranking); j++ {
			if ranking[j].Priority > ranking[i].Priority {
				ranking[i], ranking[j] = ranking[j], ranking[i]
			}
		}
	}

	// Assign ranks
	for i := range ranking {
		ranking[i].Rank = i + 1
	}

	return ranking
}

// calculateConsensusMetrics calculates agreement metrics
func (m *AHPModel) calculateConsensusMetrics(results []IndividualResult) ConsensusMetrics {
	metrics := ConsensusMetrics{}

	// Calculate criteria weight consensus (using coefficient of variation)
	criteriaCV := m.calculateWeightConsensus(results, "criteria")
	metrics.CriteriaConsensus = 1.0 - math.Min(1.0, criteriaCV)

	// Calculate ranking consensus (simplified Kendall's W)
	metrics.RankingConsensus = m.calculateRankingConsensus(results)

	// Overall consensus index
	metrics.ConsensusIndex = (metrics.CriteriaConsensus + metrics.RankingConsensus) / 2

	// Determine level and recommendation
	if metrics.ConsensusIndex >= 0.8 {
		metrics.ConsensusLevel = "high"
		metrics.Recommendation = "Strong agreement among decision makers. Group decision is reliable."
	} else if metrics.ConsensusIndex >= 0.6 {
		metrics.ConsensusLevel = "medium"
		metrics.Recommendation = "Moderate agreement. Consider discussing areas of disagreement."
	} else {
		metrics.ConsensusLevel = "low"
		metrics.Recommendation = "Significant disagreement. Recommend facilitated discussion before finalizing."
	}

	return metrics
}

// calculateWeightConsensus calculates consensus on weights using coefficient of variation
func (m *AHPModel) calculateWeightConsensus(results []IndividualResult, weightType string) float64 {
	if len(results) < 2 {
		return 0
	}

	// Get all keys
	var keys []string
	if weightType == "criteria" {
		for k := range results[0].Output.CriteriaWeights {
			keys = append(keys, k)
		}
	} else {
		for k := range results[0].Output.AlternativePriorities {
			keys = append(keys, k)
		}
	}

	totalCV := 0.0
	for _, key := range keys {
		values := make([]float64, len(results))
		for i, r := range results {
			if weightType == "criteria" {
				values[i] = r.Output.CriteriaWeights[key]
			} else {
				values[i] = r.Output.AlternativePriorities[key]
			}
		}

		// Calculate coefficient of variation
		mean := arithmeticMean(values)
		if mean > 0 {
			variance := 0.0
			for _, v := range values {
				variance += (v - mean) * (v - mean)
			}
			variance /= float64(len(values))
			stdDev := math.Sqrt(variance)
			totalCV += stdDev / mean
		}
	}

	return totalCV / float64(len(keys))
}

// calculateRankingConsensus calculates agreement on rankings
func (m *AHPModel) calculateRankingConsensus(results []IndividualResult) float64 {
	if len(results) < 2 {
		return 1.0
	}

	// Compare each pair of rankings
	agreements := 0
	comparisons := 0

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			// Compare top choice
			if results[i].Output.Ranking[0].AlternativeID == results[j].Output.Ranking[0].AlternativeID {
				agreements++
			}
			comparisons++
		}
	}

	if comparisons == 0 {
		return 1.0
	}

	return float64(agreements) / float64(comparisons)
}

// analyzeDisagreements identifies where decision makers disagree most
func (m *AHPModel) analyzeDisagreements(results []IndividualResult) []DisagreementItem {
	items := make([]DisagreementItem, 0)

	// Analyze criteria weight disagreements
	for criterionID := range results[0].Output.CriteriaWeights {
		values := make([]float64, len(results))
		for i, r := range results {
			values[i] = r.Output.CriteriaWeights[criterionID]
		}

		variance := calculateVariance(values)
		minVal, maxVal := minMax(values)

		if variance > 0.01 { // Threshold for significant disagreement
			level := "low"
			if variance > 0.05 {
				level = "high"
			} else if variance > 0.02 {
				level = "medium"
			}

			items = append(items, DisagreementItem{
				Type:              "criteria_weight",
				ElementID:         criterionID,
				Variance:          variance,
				MinValue:          minVal,
				MaxValue:          maxVal,
				DisagreementLevel: level,
			})
		}
	}

	return items
}

// Helper functions

func geometricMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	product := 1.0
	for _, v := range values {
		if v > 0 {
			product *= v
		}
	}
	return math.Pow(product, 1.0/float64(len(values)))
}

func arithmeticMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func weightedMean(values, weights []float64) float64 {
	if len(values) == 0 || len(values) != len(weights) {
		return 0
	}
	sum := 0.0
	weightSum := 0.0
	for i, v := range values {
		sum += v * weights[i]
		weightSum += weights[i]
	}
	if weightSum == 0 {
		return 0
	}
	return sum / weightSum
}

func calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := arithmeticMean(values)
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	return variance / float64(len(values))
}

func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	return minVal, maxVal
}
