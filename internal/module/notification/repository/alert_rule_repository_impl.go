package repository

import (
	"context"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/shared"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type alertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository creates a new alert rule repository
func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &alertRuleRepository{db: db}
}

func (r *alertRuleRepository) Create(ctx context.Context, rule *domain.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return err
	}
	return nil
}

func (r *alertRuleRepository) GetByID(ctx context.Context, id string) (*domain.AlertRule, error) {
	ruleID, err := uuid.Parse(id)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	var rule domain.AlertRule
	if err := r.db.WithContext(ctx).First(&rule, "id = ?", ruleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &rule, nil
}

func (r *alertRuleRepository) Update(ctx context.Context, rule *domain.AlertRule) error {
	result := r.db.WithContext(ctx).Save(rule)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *alertRuleRepository) Delete(ctx context.Context, id string) error {
	ruleID, err := uuid.Parse(id)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	result := r.db.WithContext(ctx).Delete(&domain.AlertRule{}, "id = ?", ruleID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *alertRuleRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.AlertRule, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var rules []domain.AlertRule
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *alertRuleRepository) ListByUserAndType(ctx context.Context, userID string, ruleType domain.AlertRuleType, limit, offset int) ([]domain.AlertRule, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var rules []domain.AlertRule
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userUUID, ruleType).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *alertRuleRepository) ListEnabled(ctx context.Context, userID string) ([]domain.AlertRule, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID")
	}

	var rules []domain.AlertRule
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND enabled = ?", userUUID, true).
		Order("created_at DESC").
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *alertRuleRepository) ListScheduled(ctx context.Context) ([]domain.AlertRule, error) {
	now := time.Now()
	var rules []domain.AlertRule

	// Find all enabled alert rules with schedules that are due
	if err := r.db.WithContext(ctx).
		Where("enabled = ? AND schedule IS NOT NULL", true).
		Where("next_trigger_at IS NULL OR next_trigger_at <= ?", now).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *alertRuleRepository) UpdateLastTriggered(ctx context.Context, id string, lastTriggered, nextTrigger *time.Time) error {
	ruleID, err := uuid.Parse(id)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("field", "id").WithDetails("reason", "invalid UUID")
	}

	updates := map[string]interface{}{
		"last_triggered_at": lastTriggered,
		"next_trigger_at":   nextTrigger,
	}

	result := r.db.WithContext(ctx).
		Model(&domain.AlertRule{}).
		Where("id = ?", ruleID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
