package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"go.uber.org/zap"
)

// ExecuteAutoScoring runs AHP with automatic criteria scoring from Goals
// This is the simplest method - user provides goals and monthly income,
// system auto-calculates all criteria scores
func (s *service) ExecuteAutoScoring(ctx context.Context, input *dto.AutoScoringInput) (*dto.AHPOutput, error) {
	s.logger.Info("Executing AHP with auto-scoring",
		zap.String("user_id", input.UserID),
		zap.Int("goal_count", len(input.Goals)),
		zap.Float64("monthly_income", input.MonthlyIncome))

	// Validate input
	if len(input.Goals) < 2 {
		return nil, fmt.Errorf("at least 2 goals required for prioritization")
	}

	// Default criteria weights if not provided
	criteriaWeights := input.CriteriaWeights
	if len(criteriaWeights) == 0 {
		criteriaWeights = map[string]float64{
			"urgency":     0.25,
			"importance":  0.25,
			"feasibility": 0.25,
			"impact":      0.25,
		}
		s.logger.Info("Using default equal criteria weights")
	}

	// Build alternatives from goals
	alternatives := make([]domain.Alternative, 0, len(input.Goals))
	for _, goal := range input.Goals {
		alternatives = append(alternatives, domain.Alternative{
			ID:          goal.ID,
			Name:        goal.Name,
			Description: "",
		})
	}

	// Auto-score each goal for each criterion
	localPriorities := make(map[string]map[string]float64)
	for criterionName := range criteriaWeights {
		localPriorities[criterionName] = make(map[string]float64)
	}

	for _, goal := range input.Goals {
		// Convert to GoalLike
		goalLike := convertToGoalLike(goal)

		// Calculate all criteria scores
		scores := s.autoScorer.CalculateAllCriteria(goalLike, input.MonthlyIncome)

		// Store in local priorities
		for criterion, score := range scores {
			localPriorities[criterion][goal.ID] = score
		}
	}

	// Normalize local priorities to sum to 1.0 per criterion
	for criterion := range localPriorities {
		total := 0.0
		for _, score := range localPriorities[criterion] {
			total += score
		}
		if total > 0 {
			for goalID := range localPriorities[criterion] {
				localPriorities[criterion][goalID] /= total
			}
		}
	}

	// Calculate global priorities (weighted sum)
	globalPriorities := make(map[string]float64)
	for _, goal := range input.Goals {
		priority := 0.0
		for criterion, weight := range criteriaWeights {
			priority += weight * localPriorities[criterion][goal.ID]
		}
		globalPriorities[goal.ID] = priority
	}

	// Create ranking
	ranking := createRanking(input.Goals, globalPriorities)

	// Build output
	output := &dto.AHPOutput{
		AlternativePriorities: globalPriorities,
		CriteriaWeights:       criteriaWeights,
		LocalPriorities:       localPriorities,
		ConsistencyRatio:      0.0,   // No consistency check for auto-scoring
		IsConsistent:          false, // Not applicable
		Ranking:               ranking,
	}

	s.logger.Info("Auto-scoring AHP completed",
		zap.String("user_id", input.UserID),
		zap.Int("alternatives_ranked", len(ranking)))

	return output, nil
}

// ExecuteDirectRating runs simplified AHP using direct ratings
// User provides 1-10 ratings for criteria importance
// System auto-scores goals using those criteria
func (s *service) ExecuteDirectRating(ctx context.Context, input *dto.DirectRatingInput) (*dto.AHPOutput, error) {
	s.logger.Info("Executing AHP with direct rating",
		zap.String("user_id", input.UserID),
		zap.Int("goal_count", len(input.Goals)))

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Normalize criteria ratings to weights
	criteriaWeights := normalizeCriteriaRatings(input.CriteriaRatings)

	// Auto-score goals per criterion
	localPriorities := make(map[string]map[string]float64)
	for criterionName := range criteriaWeights {
		localPriorities[criterionName] = make(map[string]float64)
	}

	for _, goal := range input.Goals {
		goalLike := convertToGoalLike(goal)
		scores := s.autoScorer.CalculateAllCriteria(goalLike, input.MonthlyIncome)

		for criterion, score := range scores {
			localPriorities[criterion][goal.ID] = score
		}
	}

	// Normalize local priorities
	for criterion := range localPriorities {
		total := 0.0
		for _, score := range localPriorities[criterion] {
			total += score
		}
		if total > 0 {
			for goalID := range localPriorities[criterion] {
				localPriorities[criterion][goalID] /= total
			}
		}
	}

	// Calculate global priorities
	globalPriorities := make(map[string]float64)
	for _, goal := range input.Goals {
		priority := 0.0
		for criterion, weight := range criteriaWeights {
			priority += weight * localPriorities[criterion][goal.ID]
		}
		globalPriorities[goal.ID] = priority
	}

	// Create ranking
	ranking := createRanking(input.Goals, globalPriorities)

	output := &dto.AHPOutput{
		AlternativePriorities: globalPriorities,
		CriteriaWeights:       criteriaWeights,
		LocalPriorities:       localPriorities,
		ConsistencyRatio:      0.0,
		IsConsistent:          false,
		Ranking:               ranking,
	}

	s.logger.Info("Direct rating AHP completed",
		zap.String("user_id", input.UserID),
		zap.Int("alternatives_ranked", len(ranking)))

	return output, nil
}

// Helper functions

func convertToGoalLike(goal dto.GoalForRating) *GoalLike {
	// Parse goal category
	var goalCategory GoalCategory
	switch goal.Type {
	case "emergency":
		goalCategory = GoalCategoryEmergency
	case "debt":
		goalCategory = GoalCategoryDebt
	case "retirement":
		goalCategory = GoalCategoryRetirement
	case "education":
		goalCategory = GoalCategoryEducation
	case "purchase":
		goalCategory = GoalCategoryPurchase
	case "investment":
		goalCategory = GoalCategoryInvestment
	case "savings":
		goalCategory = GoalCategorySavings
	case "travel":
		goalCategory = GoalCategoryTravel
	default:
		goalCategory = GoalCategoryOther
	}

	// Parse priority
	var priority GoalPriority
	switch goal.Priority {
	case "critical":
		priority = GoalPriorityCritical
	case "high":
		priority = GoalPriorityHigh
	case "low":
		priority = GoalPriorityLow
	default:
		priority = GoalPriorityMedium
	}

	remainingAmount := goal.TargetAmount - goal.CurrentAmount
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	return &GoalLike{
		Category:        goalCategory,
		Priority:        priority,
		TargetAmount:    goal.TargetAmount,
		CurrentAmount:   goal.CurrentAmount,
		TargetDate:      &goal.TargetDate,
		RemainingAmount: remainingAmount,
		Status:          GoalStatusActive,
	}
}

func normalizeCriteriaRatings(ratings map[string]int) map[string]float64 {
	weights := make(map[string]float64)
	total := 0

	for _, rating := range ratings {
		total += rating
	}

	if total > 0 {
		for criterion, rating := range ratings {
			weights[criterion] = float64(rating) / float64(total)
		}
	} else {
		// Equal weights if all ratings are 0
		equalWeight := 1.0 / float64(len(ratings))
		for criterion := range ratings {
			weights[criterion] = equalWeight
		}
	}

	return weights
}

func createRanking(goals []dto.GoalForRating, priorities map[string]float64) []domain.RankItem {
	ranking := make([]domain.RankItem, 0, len(goals))

	for _, goal := range goals {
		ranking = append(ranking, domain.RankItem{
			AlternativeID:   goal.ID,
			AlternativeName: goal.Name,
			Priority:        priorities[goal.ID],
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
