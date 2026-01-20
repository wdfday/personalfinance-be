package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/shared"
)

// ==================== Mock Repository ====================

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Category, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) ListWithChildren(ctx context.Context, userID uuid.UUID, query dto.ListCategoriesQuery) ([]*domain.Category, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByType(ctx context.Context, userID uuid.UUID, categoryType domain.CategoryType) ([]*domain.Category, error) {
	args := m.Called(ctx, userID, categoryType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]any) error {
	args := m.Called(ctx, id, columns)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetCategoryStats(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID) (*domain.Category, error) {
	args := m.Called(ctx, categoryID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) CheckNameExists(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) GetSystemCategories(ctx context.Context) ([]*domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) BulkCreate(ctx context.Context, categories []*domain.Category) error {
	args := m.Called(ctx, categories)
	return args.Error(0)
}

// ==================== CreateCategory Tests ====================

func TestCreateCategory(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully create category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		req := dto.CreateCategoryRequest{
			Name: "Groceries",
			Type: "expense",
		}

		mockRepo.On("CheckNameExists", ctx, userID, "Groceries", (*uuid.UUID)(nil)).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Category")).Return(nil)
		mockRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&domain.Category{
			ID:   uuid.New(),
			Name: "Groceries",
		}, nil)

		result, err := svc.CreateCategory(ctx, userID.String(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Groceries", result.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		req := dto.CreateCategoryRequest{Name: "Test"}

		result, err := svc.CreateCategory(ctx, "invalid-uuid", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - name already exists", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		req := dto.CreateCategoryRequest{
			Name: "Existing Category",
			Type: "expense",
		}

		mockRepo.On("CheckNameExists", ctx, userID, "Existing Category", (*uuid.UUID)(nil)).Return(true, nil)

		result, err := svc.CreateCategory(ctx, userID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - parent not found", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		parentID := uuid.New()
		parentIDStr := parentID.String()
		req := dto.CreateCategoryRequest{
			Name:     "Child Category",
			Type:     "expense",
			ParentID: &parentIDStr,
		}

		mockRepo.On("CheckNameExists", ctx, userID, "Child Category", (*uuid.UUID)(nil)).Return(false, nil)
		mockRepo.On("GetByUserID", ctx, parentID, userID).Return(nil, shared.ErrNotFound)

		result, err := svc.CreateCategory(ctx, userID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetCategory Tests ====================

func TestGetCategory(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully get category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()

		expectedCategory := &domain.Category{
			ID:   categoryID,
			Name: "Test Category",
		}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(expectedCategory, nil)

		result, err := svc.GetCategory(ctx, userID.String(), categoryID.String())

		require.NoError(t, err)
		assert.Equal(t, expectedCategory, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		result, err := svc.GetCategory(ctx, "invalid", uuid.New().String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - invalid category ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		result, err := svc.GetCategory(ctx, uuid.New().String(), "invalid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - category not found", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(nil, shared.ErrNotFound)

		result, err := svc.GetCategory(ctx, userID.String(), categoryID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, shared.ErrNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== ListCategories Tests ====================

func TestListCategories(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully list categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		query := dto.ListCategoriesQuery{}

		expectedCategories := []*domain.Category{
			{ID: uuid.New(), Name: "Category 1"},
			{ID: uuid.New(), Name: "Category 2"},
		}

		mockRepo.On("List", ctx, userID, query).Return(expectedCategories, nil)

		result, err := svc.ListCategories(ctx, userID.String(), query)

		require.NoError(t, err)
		assert.Len(t, result, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		result, err := svc.ListCategories(ctx, "invalid", dto.ListCategoriesQuery{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("list with stats", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		catID := uuid.New()
		query := dto.ListCategoriesQuery{IncludeStats: true}

		categories := []*domain.Category{{ID: catID, Name: "Category"}}

		mockRepo.On("List", ctx, userID, query).Return(categories, nil)
		mockRepo.On("GetCategoryStats", ctx, catID, userID).Return(&domain.Category{
			TransactionCount: 10,
			TotalAmount:      50000,
		}, nil)

		result, err := svc.ListCategories(ctx, userID.String(), query)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 10, result[0].TransactionCount)
		assert.Equal(t, float64(50000), result[0].TotalAmount)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== DeleteCategory Tests ====================

func TestDeleteCategory(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully delete category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		userIDPtr := userID

		category := &domain.Category{
			ID:     categoryID,
			UserID: &userIDPtr,
			Name:   "Test",
		}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(category, nil)
		mockRepo.On("GetChildren", ctx, categoryID).Return([]*domain.Category{}, nil)
		mockRepo.On("Delete", ctx, categoryID).Return(nil)

		err := svc.DeleteCategory(ctx, userID.String(), categoryID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		err := svc.DeleteCategory(ctx, "invalid", uuid.New().String())

		assert.Error(t, err)
	})

	t.Run("error - cannot delete system category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()

		// System category has nil UserID
		category := &domain.Category{
			ID:     categoryID,
			UserID: nil, // System category
			Name:   "System Category",
		}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(category, nil)

		err := svc.DeleteCategory(ctx, userID.String(), categoryID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - cannot delete category with children", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		userIDPtr := userID

		category := &domain.Category{
			ID:     categoryID,
			UserID: &userIDPtr,
		}

		children := []*domain.Category{{ID: uuid.New(), Name: "Child"}}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(category, nil)
		mockRepo.On("GetChildren", ctx, categoryID).Return(children, nil)

		err := svc.DeleteCategory(ctx, userID.String(), categoryID.String())

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== UpdateCategory Tests ====================

func TestUpdateCategory(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully update category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		userIDPtr := userID
		newName := "Updated Name"

		existingCategory := &domain.Category{
			ID:     categoryID,
			UserID: &userIDPtr,
			Name:   "Old Name",
		}

		req := dto.UpdateCategoryRequest{
			Name: &newName,
		}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(existingCategory, nil).Once()
		mockRepo.On("CheckNameExists", ctx, userID, newName, &categoryID).Return(false, nil)
		mockRepo.On("UpdateColumns", ctx, categoryID, mock.Anything).Return(nil)
		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(&domain.Category{
			ID:     categoryID,
			UserID: &userIDPtr,
			Name:   newName,
		}, nil).Once()

		result, err := svc.UpdateCategory(ctx, userID.String(), categoryID.String(), req)

		require.NoError(t, err)
		assert.Equal(t, newName, result.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - cannot update system category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()

		systemCategory := &domain.Category{
			ID:     categoryID,
			UserID: nil, // System category
		}

		req := dto.UpdateCategoryRequest{}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(systemCategory, nil)

		result, err := svc.UpdateCategory(ctx, userID.String(), categoryID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - circular reference", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()
		categoryID := uuid.New()
		userIDPtr := userID
		categoryIDStr := categoryID.String()

		existingCategory := &domain.Category{
			ID:     categoryID,
			UserID: &userIDPtr,
		}

		req := dto.UpdateCategoryRequest{
			ParentID: &categoryIDStr, // Self reference
		}

		mockRepo.On("GetByUserID", ctx, categoryID, userID).Return(existingCategory, nil)

		result, err := svc.UpdateCategory(ctx, userID.String(), categoryID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== GetSystemCategories Tests ====================

func TestGetSystemCategories(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully get system categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		systemCategories := []*domain.Category{
			{ID: uuid.New(), Name: "Food"},
			{ID: uuid.New(), Name: "Transport"},
		}

		mockRepo.On("GetSystemCategories", ctx).Return(systemCategories, nil)

		result, err := svc.GetSystemCategories(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - repository error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		mockRepo.On("GetSystemCategories", ctx).Return(nil, errors.New("db error"))

		result, err := svc.GetSystemCategories(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// ==================== InitializeDefaultCategories Tests ====================

func TestInitializeDefaultCategories(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("successfully initialize default categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()

		// User has no categories
		mockRepo.On("List", ctx, userID, dto.ListCategoriesQuery{IsRootOnly: true}).Return([]*domain.Category{}, nil)

		// System categories exist
		systemCategories := []*domain.Category{
			{ID: uuid.New(), Name: "Food", Level: 0},
			{ID: uuid.New(), Name: "Transport", Level: 0},
		}
		mockRepo.On("GetSystemCategories", ctx).Return(systemCategories, nil)

		// Bulk create succeeds
		mockRepo.On("BulkCreate", ctx, mock.AnythingOfType("[]*domain.Category")).Return(nil)

		err := svc.InitializeDefaultCategories(ctx, userID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("skip - user already has categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		userID := uuid.New()

		// User already has categories
		existingCategories := []*domain.Category{{ID: uuid.New(), Name: "Existing"}}
		mockRepo.On("List", ctx, userID, dto.ListCategoriesQuery{IsRootOnly: true}).Return(existingCategories, nil)

		err := svc.InitializeDefaultCategories(ctx, userID.String())

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		svc := NewService(mockRepo, logger)

		err := svc.InitializeDefaultCategories(ctx, "invalid-uuid")

		assert.Error(t, err)
	})
}
