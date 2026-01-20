package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/shared"
)

// ==================== Mock Repository ====================

type MockBudgetConstraintRepository struct {
	mock.Mock
}

func (m *MockBudgetConstraintRepository) Create(ctx context.Context, bc *domain.BudgetConstraint) error {
	args := m.Called(ctx, bc)
	return args.Error(0)
}

func (m *MockBudgetConstraintRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetConstraint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BudgetConstraint), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetArchivedByUser(ctx context.Context, userID uuid.UUID) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) (*domain.BudgetConstraint, error) {
	args := m.Called(ctx, userID, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BudgetConstraint), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetByStatus(ctx context.Context, userID uuid.UUID, status domain.ConstraintStatus) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetVersionHistory(ctx context.Context, constraintID uuid.UUID) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, constraintID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetLatestVersion(ctx context.Context, constraintID uuid.UUID) (*domain.BudgetConstraint, error) {
	args := m.Called(ctx, constraintID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BudgetConstraint), args.Error(1)
}

func (m *MockBudgetConstraintRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListBudgetConstraintsQuery) (domain.BudgetConstraints, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.BudgetConstraints), args.Error(1)
}

func (m *MockBudgetConstraintRepository) Update(ctx context.Context, bc *domain.BudgetConstraint) error {
	args := m.Called(ctx, bc)
	return args.Error(0)
}

func (m *MockBudgetConstraintRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBudgetConstraintRepository) DeleteByUserAndCategory(ctx context.Context, userID, categoryID uuid.UUID) error {
	args := m.Called(ctx, userID, categoryID)
	return args.Error(0)
}

func (m *MockBudgetConstraintRepository) Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	args := m.Called(ctx, id, archivedBy)
	return args.Error(0)
}

func (m *MockBudgetConstraintRepository) Exists(ctx context.Context, userID, categoryID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, categoryID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBudgetConstraintRepository) GetTotalMandatory(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

// ==================== CreateBudgetConstraint Tests ====================

func TestCreateBudgetConstraint(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully create constraint", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    categoryID.String(),
			MinimumAmount: 1000.00,
			StartDate:     time.Now(),
		}

		mockRepo.On("Exists", ctx, userID, categoryID).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.BudgetConstraint")).Return(nil)

		result, err := svc.CreateBudgetConstraint(ctx, userID.String(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, categoryID, result.CategoryID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    uuid.New().String(),
			MinimumAmount: 1000.00,
		}

		result, err := svc.CreateBudgetConstraint(ctx, "invalid-uuid", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid category ID", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    "invalid-uuid",
			MinimumAmount: 1000.00,
		}

		result, err := svc.CreateBudgetConstraint(ctx, uuid.New().String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - negative minimum amount", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    uuid.New().String(),
			MinimumAmount: -100.00,
		}

		result, err := svc.CreateBudgetConstraint(ctx, uuid.New().String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - max below min", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		maxAmount := 500.0
		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    uuid.New().String(),
			MinimumAmount: 1000.00,
			MaximumAmount: &maxAmount,
		}

		result, err := svc.CreateBudgetConstraint(ctx, uuid.New().String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - constraint already exists", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		req := dto.CreateBudgetConstraintRequest{
			CategoryID:    categoryID.String(),
			MinimumAmount: 1000.00,
		}

		mockRepo.On("Exists", ctx, userID, categoryID).Return(true, nil)

		result, err := svc.CreateBudgetConstraint(ctx, userID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetBudgetConstraint Tests ====================

func TestGetBudgetConstraint(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully get constraint", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()
		expected := &domain.BudgetConstraint{
			ID:            constraintID,
			UserID:        userID,
			MinimumAmount: 1000.00,
		}

		mockRepo.On("GetByID", ctx, constraintID).Return(expected, nil)

		result, err := svc.GetBudgetConstraint(ctx, userID.String(), constraintID.String())

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()

		mockRepo.On("GetByID", ctx, constraintID).Return(nil, shared.ErrNotFound)

		result, err := svc.GetBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetActiveConstraints Tests ====================

func TestGetActiveConstraints(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully get active constraints", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		expected := domain.BudgetConstraints{
			{ID: uuid.New(), MinimumAmount: 1000},
			{ID: uuid.New(), MinimumAmount: 2000},
		}

		mockRepo.On("GetActiveByUser", ctx, userID).Return(expected, nil)

		result, err := svc.GetActiveConstraints(ctx, userID.String())

		require.NoError(t, err)
		assert.Len(t, result, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		result, err := svc.GetActiveConstraints(ctx, "invalid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ==================== DeleteBudgetConstraint Tests ====================

func TestDeleteBudgetConstraint(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully delete constraint", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()
		existing := &domain.BudgetConstraint{
			ID:     constraintID,
			UserID: userID,
		}

		mockRepo.On("GetByID", ctx, constraintID).Return(existing, nil)
		mockRepo.On("Delete", ctx, constraintID).Return(nil)

		err := svc.DeleteBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()

		mockRepo.On("GetByID", ctx, constraintID).Return(nil, shared.ErrNotFound)

		err := svc.DeleteBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - belongs to another user", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		anotherUserID := uuid.New()
		constraintID := uuid.New()
		existing := &domain.BudgetConstraint{
			ID:     constraintID,
			UserID: anotherUserID, // Different user
		}

		mockRepo.On("GetByID", ctx, constraintID).Return(existing, nil)

		err := svc.DeleteBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== ArchiveBudgetConstraint Tests ====================

func TestArchiveBudgetConstraint(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully archive constraint", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()
		existing := &domain.BudgetConstraint{
			ID:     constraintID,
			UserID: userID,
			Status: domain.ConstraintStatusActive,
		}

		mockRepo.On("GetByID", ctx, constraintID).Return(existing, nil)
		mockRepo.On("Archive", ctx, constraintID, userID).Return(nil)

		err := svc.ArchiveBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - already archived", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraintID := uuid.New()
		existing := &domain.BudgetConstraint{
			ID:     constraintID,
			UserID: userID,
			Status: domain.ConstraintStatusArchived, // Already archived
		}

		mockRepo.On("GetByID", ctx, constraintID).Return(existing, nil)

		err := svc.ArchiveBudgetConstraint(ctx, userID.String(), constraintID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetBudgetConstraintSummary Tests ====================

func TestGetBudgetConstraintSummary(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully get summary", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		constraints := domain.BudgetConstraints{
			{ID: uuid.New(), UserID: userID, MinimumAmount: 1000, IsFlexible: false, Status: domain.ConstraintStatusActive, StartDate: time.Now().Add(-time.Hour)},
			{ID: uuid.New(), UserID: userID, MinimumAmount: 2000, IsFlexible: true, MaximumAmount: 3000, Status: domain.ConstraintStatusActive, StartDate: time.Now().Add(-time.Hour)},
		}

		mockRepo.On("GetByUser", ctx, userID).Return(constraints, nil)

		result, err := svc.GetBudgetConstraintSummary(ctx, userID.String())

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Count)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		result, err := svc.GetBudgetConstraintSummary(ctx, "invalid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - repo error", func(t *testing.T) {
		mockRepo := new(MockBudgetConstraintRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		mockRepo.On("GetByUser", ctx, userID).Return(nil, errors.New("db error"))

		result, err := svc.GetBudgetConstraintSummary(ctx, userID.String())

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}
