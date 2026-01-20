package dto

import (
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
)

// AHPInput là input hoàn chỉnh cho model
type AHPInput struct {
	UserID string `json:"user_id" binding:"required"`

	// Criteria và pairwise comparisons giữa criteria
	Criteria            []domain.Criteria           `json:"criteria" binding:"required,min=2"`
	CriteriaComparisons []domain.PairwiseComparison `json:"criteria_comparisons" binding:"required"`

	// Alternatives (Goals) và comparisons cho từng criterion
	Alternatives []domain.Alternative `json:"alternatives" binding:"required,min=2"`
	// Key = CriteriaID, Value = comparisons của alternatives theo criteria đó
	AlternativeComparisons map[string][]domain.PairwiseComparison `json:"alternative_comparisons" binding:"required"`
}

// AHPOutput là kết quả từ AHP
type AHPOutput struct {
	// Global priorities của alternatives (kết quả cuối cùng)
	// Key = AlternativeID, Value = Global priority weight
	AlternativePriorities map[string]float64 `json:"alternative_priorities"`

	// Weights của criteria
	CriteriaWeights map[string]float64 `json:"criteria_weights"`

	// Local priorities của alternatives cho từng criterion
	// First key = CriteriaID, Second key = AlternativeID, Value = Local priority
	LocalPriorities map[string]map[string]float64 `json:"local_priorities"`

	// Consistency metrics
	ConsistencyRatio float64 `json:"consistency_ratio"`
	IsConsistent     bool    `json:"is_consistent"` // CR < 0.10

	// Ranking của alternatives
	Ranking []domain.RankItem `json:"ranking"`
}

// ToAHPOutput converts internal result to DTO
func ToAHPOutput(
	altPriorities map[string]float64,
	criteriaWeights map[string]float64,
	localPriorities map[string]map[string]float64,
	cr float64,
	ranking []domain.RankItem,
) *AHPOutput {
	return &AHPOutput{
		AlternativePriorities: altPriorities,
		CriteriaWeights:       criteriaWeights,
		LocalPriorities:       localPriorities,
		ConsistencyRatio:      cr,
		IsConsistent:          cr < 0.10,
		Ranking:               ranking,
	}
}
