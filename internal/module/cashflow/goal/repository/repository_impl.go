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

func (r *repository) FindByType(ctx context.Context, userID uuid.UUID, goalType domain.GoalType) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, goalType).
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
