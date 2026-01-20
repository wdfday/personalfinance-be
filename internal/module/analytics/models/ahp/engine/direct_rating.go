package engine

import (
	"context"
	"fmt"
	"sort"
	"time"

	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"personalfinancedss/internal/module/analytics/goal_prioritization/service"
)

// DirectRatingEngine implements direct rating method for goal prioritization
// This is a simpler alternative to full AHP that requires fewer user inputs
// Trade-off: Faster input (18 vs 33) but no consistency check
type DirectRatingEngine struct {
	autoScorer *service.AutoScorer
}

// NewDirectRatingEngine creates a new DirectRatingEngine
func NewDirectRatingEngine() *DirectRatingEngine {
	return &DirectRatingEngine{
		autoScorer: service.NewAutoScorer(),
	}
}

// Execute calculates goal priorities using direct rating method
// Algorithm:
// 1. Normalize criteria ratings to weights (sum = 1.0)
// 2. Auto-score each goal per criterion (0-1 scale)
// 3. Calculate weighted sum: priority = Σ(weight × score)
// 4. Sort by priority descending
//
// Returns AHPOutput-compatible result (but ConsistencyRatio = 0, IsConsistent = false)
func (dre *DirectRatingEngine) Execute(
	ctx context.Context,
	input *dto.DirectRatingInput,
) (*dto.AHPOutput, error) {
	startTime := time.Now()

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 1: Normalize criteria ratings to weights
	criteriaWeights := dre.normalizeCriteriaRatings(input.CriteriaRatings)

	// Step 2: Auto-score goals per criterion
	goalScores := make(map[string]map[string]float64)      // goal_id -> criterion -> score
	localPriorities := make(map[string]map[string]float64) // criterion -> goal_id -> score

	for criterionName := range criteriaWeights {
		localPriorities[criterionName] = make(map[string]float64)
	}

	for _, goal := range input.Goals {
		scores := dre.calculateGoalScores(goal, input.MonthlyIncome)
		goalScores[goal.ID] = scores

		// Populate local priorities (for AHP compatibility)
		for criterion, score := range scores {
			localPriorities[criterion][goal.ID] = score
		}
	}

	// Step 3: Calculate global priorities (weighted sum)
	globalPriorities := make(map[string]float64)

	for _, goal := range input.Goals {
		priority := 0.0
		for criterion, weight := range criteriaWeights {
			score := goalScores[goal.ID][criterion]
			priority += weight * score
		}
		globalPriorities[goal.ID] = priority
	}

	// Step 4: Normalize global priorities to sum to 1.0
	totalPriority := 0.0
	for _, priority := range globalPriorities {
		totalPriority += priority
	}
	if totalPriority > 0 {
		for goalID := range globalPriorities {
			globalPriorities[goalID] /= totalPriority
		}
		// Re-normalize local priorities too
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
	}

	// Step 5: Create ranking
	ranking := dre.createRanking(input.Goals, globalPriorities)

	// Build AHP-compatible output
	output := &dto.AHPOutput{
		AlternativePriorities: globalPriorities,
		CriteriaWeights:       criteriaWeights,
		LocalPriorities:       localPriorities,
		ConsistencyRatio:      0.0,   // No consistency check for direct rating
		IsConsistent:          false, // Cannot guarantee consistency
		Ranking:               ranking,
	}

	executionTime := time.Since(startTime).Milliseconds()
	fmt.Printf("DirectRatingEngine executed in %dms\n", executionTime)

	return output, nil
}

// normalizeCriteriaRatings converts ratings (1-10) to weights (sum = 1.0)
func (dre *DirectRatingEngine) normalizeCriteriaRatings(ratings map[string]int) map[string]float64 {
	weights := make(map[string]float64)
	total := 0

	// Sum all ratings
	for _, rating := range ratings {
		total += rating
	}

	// Normalize
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

// calculateGoalScores calculates scores for a goal across all criteria
func (dre *DirectRatingEngine) calculateGoalScores(
	goal dto.GoalForRating,
	monthlyIncome float64,
) map[string]float64 {
	// Convert GoalForRating to goal domain (for auto-scorer)
	goalDomain := dre.convertToGoalDomain(goal)

	// Calculate all criteria using auto-scorer
	return dre.autoScorer.CalculateAllCriteria(goalDomain, monthlyIncome)
}

// convertToGoalDomain converts GoalForRating to goal domain entity
func (dre *DirectRatingEngine) convertToGoalDomain(goal dto.GoalForRating) *service.GoalLike {
	// Parse goal category
	var goalCategory service.GoalCategory
	switch goal.Type {
	case "savings":
		goalCategory = service.GoalCategorySavings
	case "debt":
		goalCategory = service.GoalCategoryDebt
	case "investment":
		goalCategory = service.GoalCategoryInvestment
	case "purchase":
		goalCategory = service.GoalCategoryPurchase
	case "emergency":
		goalCategory = service.GoalCategoryEmergency
	case "retirement":
		goalCategory = service.GoalCategoryRetirement
	case "education":
		goalCategory = service.GoalCategoryEducation
	case "travel":
		goalCategory = service.GoalCategoryTravel
	default:
		goalCategory = service.GoalCategoryOther
	}

	// Parse priority
	var priority service.GoalPriority
	switch goal.Priority {
	case "critical":
		priority = service.GoalPriorityCritical
	case "high":
		priority = service.GoalPriorityHigh
	case "low":
		priority = service.GoalPriorityLow
	default:
		priority = service.GoalPriorityMedium
	}

	remainingAmount := goal.TargetAmount - goal.CurrentAmount
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	return &service.GoalLike{
		Category:        goalCategory,
		Priority:        priority,
		TargetAmount:    goal.TargetAmount,
		CurrentAmount:   goal.CurrentAmount,
		TargetDate:      &goal.TargetDate,
		RemainingAmount: remainingAmount,
		Status:          service.GoalStatusActive,
	}
}

// createRanking creates sorted ranking from priorities
func (dre *DirectRatingEngine) createRanking(
	goals []dto.GoalForRating,
	priorities map[string]float64,
) []domain.RankItem {
	ranking := make([]domain.RankItem, 0, len(goals))

	for _, goal := range goals {
		ranking = append(ranking, domain.RankItem{
			AlternativeID:   goal.ID,
			AlternativeName: goal.Name,
			Priority:        priorities[goal.ID],
		})
	}

	// Sort by priority descending
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].Priority > ranking[j].Priority
	})

	// Assign ranks
	for i := range ranking {
		ranking[i].Rank = i + 1
	}

	return ranking
}
