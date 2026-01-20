package tests

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, debt *domain.Debt) error {
	args := m.Called(ctx, debt)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Debt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Debt), args.Error(1)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) FindByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error) {
	args := m.Called(ctx, userID, debtType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) FindByStatus(ctx context.Context, userID uuid.UUID, status domain.DebtStatus) ([]domain.Debt, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) FindPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) FindOverdueDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Debt), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, debt *domain.Debt) error {
	args := m.Called(ctx, debt)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) AddPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	args := m.Called(ctx, id, amount)
	return args.Error(0)
}
