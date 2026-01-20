package repository

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// New creates a new goal repository
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, goal *domain.Goal) error {
	return r.db.WithContext(ctx).Create(goal).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	var goal domain.Goal
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&goal).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("goal not found")
		}
		return nil, err
	}
	return &goal, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("priority DESC, created_at DESC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.GoalStatusActive).
		Order("priority DESC, target_date ASC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) FindByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category).
		Order("created_at DESC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) FindByStatus(ctx context.Context, userID uuid.UUID, status domain.GoalStatus) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) FindCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.GoalStatusCompleted).
		Order("completed_at DESC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) FindOverdueGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	var goals []domain.Goal
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_date < ? AND status NOT IN (?)", userID, now, []string{
			string(domain.GoalStatusCompleted),
			string(domain.GoalStatusCancelled),
		}).
		Order("target_date ASC").
		Find(&goals).Error
	return goals, err
}

func (r *repository) Update(ctx context.Context, goal *domain.Goal) error {
	return r.db.WithContext(ctx).Save(goal).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Goal{}, id).Error
}

func (r *repository) AddContribution(ctx context.Context, id uuid.UUID, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Goal{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_amount": gorm.Expr("current_amount + ?", amount),
		}).Error
}

// ============================================================
// Contribution Methods
// ============================================================

func (r *repository) CreateContribution(ctx context.Context, contribution *domain.GoalContribution) error {
	return r.db.WithContext(ctx).Create(contribution).Error
}

func (r *repository) FindContributionsByGoalID(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	var contributions []domain.GoalContribution
	err := r.db.WithContext(ctx).
		Where("goal_id = ?", goalID).
		Order("created_at DESC").
		Find(&contributions).Error
	return contributions, err
}

func (r *repository) FindContributionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]domain.GoalContribution, error) {
	var contributions []domain.GoalContribution
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Find(&contributions).Error
	return contributions, err
}

func (r *repository) GetNetContributionsByAccountID(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var result struct {
		NetAmount float64
	}

	// Calculate: SUM(deposits) - SUM(withdrawals)
	err := r.db.WithContext(ctx).
		Model(&domain.GoalContribution{}).
		Select(`
			COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'withdrawal' THEN amount ELSE 0 END), 0) as net_amount
		`).
		Where("account_id = ?", accountID).
		Scan(&result).Error

	return result.NetAmount, err
}

func (r *repository) GetNetContributionsByGoalID(ctx context.Context, goalID uuid.UUID) (float64, error) {
	var result struct {
		NetAmount float64
	}

	err := r.db.WithContext(ctx).
		Model(&domain.GoalContribution{}).
		Select(`
			COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'withdrawal' THEN amount ELSE 0 END), 0) as net_amount
		`).
		Where("goal_id = ?", goalID).
		Scan(&result).Error

	return result.NetAmount, err
}

// GetContributionsByDateRange retrieves contributions for a goal within a date range
func (r *repository) GetContributionsByDateRange(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error) {
	var contributions []domain.GoalContribution
	err := r.db.WithContext(ctx).
		Where("goal_id = ? AND created_at BETWEEN ? AND ?", goalID, startDate, endDate).
		Order("created_at ASC").
		Find(&contributions).Error
	return contributions, err
}
