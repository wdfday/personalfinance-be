package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"

	"go.uber.org/zap"
)

type alertRuleService struct {
	repo   repository.AlertRuleRepository
	logger *zap.Logger
	// TODO: Add budget and goal repositories when those modules are implemented
}

// NewAlertRuleService creates a new alert rule service
func NewAlertRuleService(
	repo repository.AlertRuleRepository,
	logger *zap.Logger,
) AlertRuleService {
	return &alertRuleService{
		repo:   repo,
		logger: logger,
	}
}

func (s *alertRuleService) CreateRule(ctx context.Context, rule *domain.AlertRule) error {
	s.logger.Info("Creating alert rule",
		zap.String("user_id", rule.UserID.String()),
		zap.String("type", string(rule.Type)),
		zap.String("name", rule.Name),
	)

	// Validate rule type
	if !rule.Type.IsValid() {
		return fmt.Errorf("invalid alert rule type: %s", rule.Type)
	}

	// Validate conditions based on rule type
	if err := s.validateRuleConditions(rule); err != nil {
		return fmt.Errorf("invalid rule conditions: %w", err)
	}

	return s.repo.Create(ctx, rule)
}

func (s *alertRuleService) GetRule(ctx context.Context, ruleID string) (*domain.AlertRule, error) {
	return s.repo.GetByID(ctx, ruleID)
}

func (s *alertRuleService) UpdateRule(ctx context.Context, rule *domain.AlertRule) error {
	s.logger.Info("Updating alert rule",
		zap.String("id", rule.ID.String()),
		zap.String("name", rule.Name),
	)

	// Validate conditions
	if err := s.validateRuleConditions(rule); err != nil {
		return fmt.Errorf("invalid rule conditions: %w", err)
	}

	return s.repo.Update(ctx, rule)
}

func (s *alertRuleService) DeleteRule(ctx context.Context, ruleID string) error {
	s.logger.Info("Deleting alert rule", zap.String("id", ruleID))
	return s.repo.Delete(ctx, ruleID)
}

func (s *alertRuleService) ListUserRules(ctx context.Context, userID string, limit, offset int) ([]domain.AlertRule, error) {
	return s.repo.ListByUserID(ctx, userID, limit, offset)
}

func (s *alertRuleService) ListEnabledRules(ctx context.Context, userID string) ([]domain.AlertRule, error) {
	return s.repo.ListEnabled(ctx, userID)
}

func (s *alertRuleService) EvaluateRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	s.logger.Debug("Evaluating alert rule",
		zap.String("rule_id", rule.ID.String()),
		zap.String("type", string(rule.Type)),
	)

	if !rule.Enabled {
		s.logger.Debug("Rule is disabled, skipping evaluation", zap.String("rule_id", rule.ID.String()))
		return false, nil
	}

	switch rule.Type {
	case domain.AlertRuleBudgetThreshold:
		return s.evaluateBudgetThresholdRule(ctx, rule)

	case domain.AlertRuleGoalMilestone:
		return s.evaluateGoalMilestoneRule(ctx, rule)

	case domain.AlertRuleGoalOffTrack:
		return s.evaluateGoalOffTrackRule(ctx, rule)

	case domain.AlertRuleScheduledReport:
		// Scheduled reports are handled by the scheduler service
		return false, nil

	case domain.AlertRuleLargeTransaction:
		return s.evaluateLargeTransactionRule(ctx, rule)

	case domain.AlertRuleUnusualActivity:
		return s.evaluateUnusualActivityRule(ctx, rule)

	default:
		return false, fmt.Errorf("unsupported alert rule type: %s", rule.Type)
	}
}

func (s *alertRuleService) EvaluateAllUserRules(ctx context.Context, userID string) error {
	s.logger.Info("Evaluating all rules for user", zap.String("user_id", userID))

	rules, err := s.repo.ListEnabled(ctx, userID)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		triggered, err := s.EvaluateRule(ctx, &rule)
		if err != nil {
			s.logger.Error("Failed to evaluate rule",
				zap.String("rule_id", rule.ID.String()),
				zap.Error(err),
			)
			continue
		}

		if triggered {
			s.logger.Info("Rule triggered",
				zap.String("rule_id", rule.ID.String()),
				zap.String("type", string(rule.Type)),
			)
			// TODO: Send notification based on rule configuration
		}
	}

	return nil
}

// validateRuleConditions validates rule conditions based on rule type
func (s *alertRuleService) validateRuleConditions(rule *domain.AlertRule) error {
	if rule.Conditions == nil || len(rule.Conditions) == 0 {
		return fmt.Errorf("conditions cannot be empty")
	}

	switch rule.Type {
	case domain.AlertRuleBudgetThreshold:
		// Required: budget_id, threshold_percentage
		if _, ok := rule.Conditions["budget_id"]; !ok {
			return fmt.Errorf("budget_id is required for budget threshold rule")
		}
		if _, ok := rule.Conditions["threshold_percentage"]; !ok {
			return fmt.Errorf("threshold_percentage is required for budget threshold rule")
		}

	case domain.AlertRuleGoalMilestone:
		// Required: goal_id, milestone_percentage
		if _, ok := rule.Conditions["goal_id"]; !ok {
			return fmt.Errorf("goal_id is required for goal milestone rule")
		}
		if _, ok := rule.Conditions["milestone_percentage"]; !ok {
			return fmt.Errorf("milestone_percentage is required for goal milestone rule")
		}

	case domain.AlertRuleScheduledReport:
		// Required: frequency
		if _, ok := rule.Conditions["frequency"]; !ok {
			return fmt.Errorf("frequency is required for scheduled report rule")
		}
		frequency := rule.Conditions["frequency"].(string)
		if frequency != "daily" && frequency != "weekly" && frequency != "monthly" {
			return fmt.Errorf("invalid frequency: must be daily, weekly, or monthly")
		}
	}

	return nil
}

// evaluateBudgetThresholdRule evaluates budget threshold rules
func (s *alertRuleService) evaluateBudgetThresholdRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	// TODO: Implement when budget module is available
	s.logger.Warn("Budget threshold rule evaluation not implemented yet (budget module not available)",
		zap.String("rule_id", rule.ID.String()),
	)
	return false, nil
}

// evaluateGoalMilestoneRule evaluates goal milestone rules
func (s *alertRuleService) evaluateGoalMilestoneRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	// TODO: Implement when goal module is available
	s.logger.Warn("Goal milestone rule evaluation not implemented yet (goal module not available)",
		zap.String("rule_id", rule.ID.String()),
	)
	return false, nil
}

// evaluateGoalOffTrackRule evaluates goal off-track rules
func (s *alertRuleService) evaluateGoalOffTrackRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	// TODO: Implement when goal module is available
	s.logger.Warn("Goal off-track rule evaluation not implemented yet (goal module not available)",
		zap.String("rule_id", rule.ID.String()),
	)
	return false, nil
}

// evaluateLargeTransactionRule evaluates large transaction rules
func (s *alertRuleService) evaluateLargeTransactionRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	// TODO: Implement large transaction detection
	s.logger.Debug("Large transaction rule evaluation",
		zap.String("rule_id", rule.ID.String()),
	)
	return false, nil
}

// evaluateUnusualActivityRule evaluates unusual activity rules
func (s *alertRuleService) evaluateUnusualActivityRule(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	// TODO: Implement unusual activity detection
	s.logger.Debug("Unusual activity rule evaluation",
		zap.String("rule_id", rule.ID.String()),
	)
	return false, nil
}
