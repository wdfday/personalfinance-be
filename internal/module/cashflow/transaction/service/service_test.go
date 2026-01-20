package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/shared"
)

// ==================== Mock Repository ====================

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByExternalID(ctx context.Context, userID uuid.UUID, externalID string) (*domain.Transaction, error) {
	args := m.Called(ctx, userID, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) ([]*domain.Transaction, int64, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]interface{}) error {
	args := m.Called(ctx, id, columns)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) GetTransactionsByDateRange(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, accountID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetSummary(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionSummary), args.Error(1)
}

func (m *MockTransactionRepository) GetRecurringTransactions(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

// ==================== GetTransaction Tests ====================

func TestGetTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get transaction", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		txID := uuid.New()
		expected := &domain.Transaction{
			ID:     txID,
			UserID: userID,
			Amount: 100000,
		}

		mockRepo.On("GetByUserID", ctx, txID, userID).Return(expected, nil)

		result, err := svc.GetTransaction(ctx, userID.String(), txID.String())

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		result, err := svc.GetTransaction(ctx, "invalid", uuid.New().String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid transaction ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		result, err := svc.GetTransaction(ctx, uuid.New().String(), "invalid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		txID := uuid.New()

		mockRepo.On("GetByUserID", ctx, txID, userID).Return(nil, shared.ErrNotFound)

		result, err := svc.GetTransaction(ctx, userID.String(), txID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, shared.ErrNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== ListTransactions Tests ====================

func TestListTransactions(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully list transactions", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		query := dto.ListTransactionsQuery{}

		transactions := []*domain.Transaction{
			{ID: uuid.New(), Amount: 100000},
			{ID: uuid.New(), Amount: 200000},
		}

		// Need to match the modified query with defaults
		mockRepo.On("List", ctx, userID, mock.AnythingOfType("dto.ListTransactionsQuery")).Return(transactions, int64(2), nil)
		mockRepo.On("GetSummary", ctx, userID, mock.AnythingOfType("dto.ListTransactionsQuery")).Return(&dto.TransactionSummary{
			TotalCredit: 200000,
			TotalDebit:  100000,
		}, nil)

		result, err := svc.ListTransactions(ctx, userID.String(), query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Transactions, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		result, err := svc.ListTransactions(ctx, "invalid", dto.ListTransactionsQuery{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with pagination defaults", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		query := dto.ListTransactionsQuery{Page: 0, PageSize: 0} // Will be set to defaults

		mockRepo.On("List", ctx, userID, mock.MatchedBy(func(q dto.ListTransactionsQuery) bool {
			return q.Page == 1 && q.PageSize == 20
		})).Return([]*domain.Transaction{}, int64(0), nil)
		mockRepo.On("GetSummary", ctx, userID, mock.AnythingOfType("dto.ListTransactionsQuery")).Return(nil, nil)

		result, err := svc.ListTransactions(ctx, userID.String(), query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 20, result.Pagination.PageSize)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetTransactionSummary Tests ====================

func TestGetTransactionSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get summary", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		query := dto.ListTransactionsQuery{}
		expected := &dto.TransactionSummary{
			TotalCredit: 500000,
			TotalDebit:  300000,
			NetAmount:   200000,
		}

		mockRepo.On("GetSummary", ctx, userID, query).Return(expected, nil)

		result, err := svc.GetTransactionSummary(ctx, userID.String(), query)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		result, err := svc.GetTransactionSummary(ctx, "invalid", dto.ListTransactionsQuery{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ==================== DeleteTransaction Tests ====================

func TestDeleteTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully delete transaction", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		txID := uuid.New()
		existing := &domain.Transaction{
			ID:     txID,
			UserID: userID,
		}

		mockRepo.On("GetByUserID", ctx, txID, userID).Return(existing, nil)
		mockRepo.On("Delete", ctx, txID).Return(nil)

		err := svc.DeleteTransaction(ctx, userID.String(), txID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		txID := uuid.New()

		mockRepo.On("GetByUserID", ctx, txID, userID).Return(nil, shared.ErrNotFound)

		err := svc.DeleteTransaction(ctx, userID.String(), txID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== CreateTransaction Tests ====================

func TestCreateTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully create transaction", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		userID := uuid.New()
		accountID := uuid.New()
		req := dto.CreateTransactionRequest{
			AccountID:   accountID.String(),
			Direction:   "DEBIT",
			Instrument:  "CASH",
			Source:      "MANUAL",
			Amount:      100000,
			BookingDate: time.Now(),
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)
		mockRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&domain.Transaction{
			ID:         uuid.New(),
			UserID:     userID,
			AccountID:  accountID,
			Amount:     100000,
			Direction:  domain.DirectionDebit,
			Instrument: domain.InstrumentCash,
		}, nil)

		result, err := svc.CreateTransaction(ctx, userID.String(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(100000), result.Amount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		req := dto.CreateTransactionRequest{
			AccountID: uuid.New().String(),
		}

		result, err := svc.CreateTransaction(ctx, "invalid", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid account ID", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		req := dto.CreateTransactionRequest{
			AccountID: "invalid",
		}

		result, err := svc.CreateTransaction(ctx, uuid.New().String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid direction", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		svc := NewService(mockRepo, nil)

		req := dto.CreateTransactionRequest{
			AccountID: uuid.New().String(),
			Direction: "INVALID",
		}

		result, err := svc.CreateTransaction(ctx, uuid.New().String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
