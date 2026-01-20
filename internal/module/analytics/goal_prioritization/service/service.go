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
	defaultWeights := map[string]float64{
		"urgency":     0.25,
		"feasibility": 0.25,
		"importance":  0.25,
		"impact":      0.25,
	}

	// Suggested ratings (all 5 by default)
	suggestedRatings := map[string]int{
		"urgency":     5,
		"feasibility": 5,
		"importance":  5,
		"impact":      5,
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
