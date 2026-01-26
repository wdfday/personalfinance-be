package service

import (
	"context"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"personalfinancedss/internal/module/analytics/models/ahp"

	"go.uber.org/zap"
)

// Service interface for AHP operations
type Service interface {
	// ExecuteAHP runs AHP model with given input (full pairwise comparisons)
	ExecuteAHP(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error)

	// ExecuteAutoScoring runs AHP with automatic criteria scoring from Goals
	ExecuteAutoScoring(ctx context.Context, input *dto.AutoScoringInput) (*dto.AHPOutput, error)

	// ExecuteDirectRating runs simplified AHP using direct ratings (18 inputs vs 33)
	ExecuteDirectRating(ctx context.Context, input *dto.DirectRatingInput) (*dto.AHPOutput, error)

	// GetAutoScores calculates auto-scores with reasons for user review
	// This is step 1 of "auto-score with user review" flow
	GetAutoScores(ctx context.Context, input *dto.AutoScoresRequest) (*dto.AutoScoresResponse, error)

	// ExecuteDirectRatingWithOverrides runs direct rating with user-modified scores
	// This is step 2 of "auto-score with user review" flow
	ExecuteDirectRatingWithOverrides(ctx context.Context, input *dto.DirectRatingWithOverridesInput) (*dto.AHPOutput, error)
}

// service implements Service interface
type service struct {
	model      *ahp.AHPModel
	autoScorer *AutoScorer
	logger     *zap.Logger
}

// NewService creates a new AHP service
func NewService(logger *zap.Logger) Service {
	return &service{
		model:      ahp.NewAHPModel(),
		autoScorer: NewAutoScorer(),
		logger:     logger,
	}
}

// ExecuteAHP runs the AHP model
func (s *service) ExecuteAHP(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
	s.logger.Info("Executing AHP model", zap.String("user_id", input.UserID))

	// Validate input
	if err := s.model.Validate(ctx, input); err != nil {
		s.logger.Error("AHP validation failed", zap.Error(err))
		return nil, err
	}

	// Execute model
	result, err := s.model.Execute(ctx, input)
	if err != nil {
		s.logger.Error("AHP execution failed", zap.Error(err))
		return nil, err
	}

	output := result.(*dto.AHPOutput)

	s.logger.Info("AHP execution completed",
		zap.String("user_id", input.UserID),
		zap.Float64("consistency_ratio", output.ConsistencyRatio),
		zap.Bool("is_consistent", output.IsConsistent))

	return output, nil
}

// GetAutoScores calculates auto-scores with reasons for user review
// This is step 1 of the "auto-score with user review" flow
func (s *service) GetAutoScores(ctx context.Context, input *dto.AutoScoresRequest) (*dto.AutoScoresResponse, error) {
	s.logger.Info("Calculating auto-scores for user review",
		zap.String("user_id", input.UserID),
		zap.Int("goal_count", len(input.Goals)))

	goalScores := make([]dto.GoalAutoScores, 0, len(input.Goals))

	for _, goal := range input.Goals {
		// Convert to GoalLike for auto-scorer
		goalLike := s.convertGoalForRatingToGoalLike(goal)

		// Calculate scores with reasons
		scores := s.autoScorer.CalculateAllCriteriaWithReasons(goalLike, input.MonthlyIncome)

		// Convert service.ScoreWithReason to dto.ScoreWithReason
		dtoScores := make(map[string]dto.ScoreWithReason)
		for k, v := range scores {
			dtoScores[k] = dto.ScoreWithReason{
				Score:  v.Score,
				Reason: v.Reason,
			}
		}

		goalScores = append(goalScores, dto.GoalAutoScores{
			GoalID:   goal.ID,
			GoalName: goal.Name,
			GoalType: goal.Type,
			Scores:   dtoScores,
		})
	}

	// Default criteria weights (equal by default)
	// NOTE: Impact is temporarily disabled - redistribute weights to other 3 criteria
	defaultWeights := map[string]float64{
		"urgency":     1.0 / 3.0, // ~0.333
		"feasibility": 1.0 / 3.0, // ~0.333
		"importance":  1.0 / 3.0, // ~0.333
		// "impact":      0.25, // Temporarily disabled
	}

	// Suggested ratings (all 5 by default)
	suggestedRatings := map[string]int{
		"urgency":     5,
		"feasibility": 5,
		"importance":  5,
		// "impact":      5, // Temporarily disabled
	}

	return &dto.AutoScoresResponse{
		Goals:                    goalScores,
		DefaultCriteriaWeights:   defaultWeights,
		SuggestedCriteriaRatings: suggestedRatings,
	}, nil
}

// ExecuteDirectRatingWithOverrides runs direct rating with user-modified scores
// This is step 2 of the "auto-score with user review" flow
func (s *service) ExecuteDirectRatingWithOverrides(ctx context.Context, input *dto.DirectRatingWithOverridesInput) (*dto.AHPOutput, error) {
	s.logger.Info("Executing direct rating with user overrides",
		zap.String("user_id", input.UserID),
		zap.Int("override_count", len(input.GoalScoreOverrides)))

	// Validate input
	if err := input.Validate(); err != nil {
		s.logger.Error("Validation failed", zap.Error(err))
		return nil, err
	}

	// For now, pass through to regular DirectRating
	// The overrides will be handled in the DirectRatingEngine
	return s.ExecuteDirectRating(ctx, &input.DirectRatingInput)
}

// convertGoalForRatingToGoalLike converts DTO to internal type
func (s *service) convertGoalForRatingToGoalLike(goal dto.GoalForRating) *GoalLike {
	// Parse goal category
	var goalCategory GoalCategory
	switch goal.Type {
	case "savings":
		goalCategory = GoalCategorySavings
	case "debt":
		goalCategory = GoalCategoryDebt
	case "investment":
		goalCategory = GoalCategoryInvestment
	case "purchase":
		goalCategory = GoalCategoryPurchase
	case "emergency":
		goalCategory = GoalCategoryEmergency
	case "retirement":
		goalCategory = GoalCategoryRetirement
	case "education":
		goalCategory = GoalCategoryEducation
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
		// NOTE: Impact is temporarily disabled - redistribute to 3 criteria
		criteriaWeights = map[string]float64{
			"urgency":     1.0 / 3.0, // ~0.333
			"importance":  1.0 / 3.0, // ~0.333
			"feasibility": 1.0 / 3.0, // ~0.333
			// "impact":      0.25, // Temporarily disabled
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

	// Normalize local priorities using improved min-max normalization
	// This prevents clustering and provides better distribution
	for criterion := range localPriorities {
		scores := localPriorities[criterion]

		// Find min and max scores
		minScore := math.MaxFloat64
		maxScore := -math.MaxFloat64
		for _, score := range scores {
			if score < minScore {
				minScore = score
			}
			if score > maxScore {
				maxScore = score
			}
		}

		// If all scores are the same, use simple normalization
		if maxScore-minScore < 0.0001 {
			total := 0.0
			for _, score := range scores {
				total += score
			}
			if total > 0 {
				for goalID := range scores {
					scores[goalID] /= total
				}
			} else {
				// Equal distribution if all scores are 0
				equalWeight := 1.0 / float64(len(scores))
				for goalID := range scores {
					scores[goalID] = equalWeight
				}
			}
		} else {
			// Min-max normalization with small epsilon to avoid division issues
			rangeSize := maxScore - minScore
			for goalID := range scores {
				// Normalize to [0, 1] range
				normalized := (scores[goalID] - minScore) / rangeSize
				scores[goalID] = normalized
			}

			// Then normalize to sum to 1.0 (AHP requirement)
			total := 0.0
			for _, score := range scores {
				total += score
			}
			if total > 0 {
				for goalID := range scores {
					scores[goalID] /= total
				}
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

	// Normalize local priorities using improved min-max normalization
	// This prevents clustering and provides better distribution
	for criterion := range localPriorities {
		scores := localPriorities[criterion]

		// Find min and max scores
		minScore := math.MaxFloat64
		maxScore := -math.MaxFloat64
		for _, score := range scores {
			if score < minScore {
				minScore = score
			}
			if score > maxScore {
				maxScore = score
			}
		}

		// If all scores are the same, use simple normalization
		if maxScore-minScore < 0.0001 {
			total := 0.0
			for _, score := range scores {
				total += score
			}
			if total > 0 {
				for goalID := range scores {
					scores[goalID] /= total
				}
			} else {
				// Equal distribution if all scores are 0
				equalWeight := 1.0 / float64(len(scores))
				for goalID := range scores {
					scores[goalID] = equalWeight
				}
			}
		} else {
			// Min-max normalization with small epsilon to avoid division issues
			rangeSize := maxScore - minScore
			for goalID := range scores {
				// Normalize to [0, 1] range
				normalized := (scores[goalID] - minScore) / rangeSize
				scores[goalID] = normalized
			}

			// Then normalize to sum to 1.0 (AHP requirement)
			total := 0.0
			for _, score := range scores {
				total += score
			}
			if total > 0 {
				for goalID := range scores {
					scores[goalID] /= total
				}
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
