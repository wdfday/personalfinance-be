package repository

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// New creates a new debt repository
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, debt *domain.Debt) error {
	return r.db.WithContext(ctx).Create(debt).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Debt, error) {
	var debt domain.Debt
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&debt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}
	return &debt, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	var debts []domain.Debt
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("status ASC, next_payment_date ASC, created_at DESC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	var debts []domain.Debt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.DebtStatusActive).
		Order("next_payment_date ASC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) FindByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error) {
	var debts []domain.Debt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, debtType).
		Order("created_at DESC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) FindByStatus(ctx context.Context, userID uuid.UUID, status domain.DebtStatus) ([]domain.Debt, error) {
	var debts []domain.Debt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) FindPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	var debts []domain.Debt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.DebtStatusPaidOff).
		Order("paid_off_date DESC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) FindOverdueDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	var debts []domain.Debt
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND next_payment_date < ? AND status NOT IN (?)", userID, now, []string{
			string(domain.DebtStatusPaidOff),
			string(domain.DebtStatusSettled),
			string(domain.DebtStatusInactive),
		}).
		Order("next_payment_date ASC").
		Find(&debts).Error
	return debts, err
}

func (r *repository) Update(ctx context.Context, debt *domain.Debt) error {
	return r.db.WithContext(ctx).Save(debt).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Debt{}, id).Error
}

func (r *repository) AddPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Debt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_balance": gorm.Expr("current_balance - ?", amount),
			"total_paid":      gorm.Expr("total_paid + ?", amount),
		}).Error
}
