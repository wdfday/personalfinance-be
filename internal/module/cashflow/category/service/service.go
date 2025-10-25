package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/category/domain"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/module/cashflow/category/repository"

	"go.uber.org/zap"
)

// CategoryCreator defines category creation operations
type CategoryCreator interface {
	CreateCategory(ctx context.Context, userID string, req dto.CreateCategoryRequest) (*domain.Category, error)
	InitializeDefaultCategories(ctx context.Context, userID string) error
}

// CategoryReader defines category read operations
type CategoryReader interface {
	GetCategory(ctx context.Context, userID string, categoryID string) (*domain.Category, error)
	ListCategories(ctx context.Context, userID string, query dto.ListCategoriesQuery) ([]*domain.Category, error)
	ListCategoriesWithChildren(ctx context.Context, userID string, query dto.ListCategoriesQuery) ([]*domain.Category, error)
	GetCategoryStats(ctx context.Context, userID string, categoryID string) (*domain.Category, error)
	GetSystemCategories(ctx context.Context) ([]*domain.Category, error)
}

// CategoryUpdater defines category update operations
type CategoryUpdater interface {
	UpdateCategory(ctx context.Context, userID string, categoryID string, req dto.UpdateCategoryRequest) (*domain.Category, error)
}

// CategoryDeleter defines category delete operations
type CategoryDeleter interface {
	DeleteCategory(ctx context.Context, userID string, categoryID string) error
}

// Service is the composite interface for all category operations
type Service interface {
	CategoryCreator
	CategoryReader
	CategoryUpdater
	CategoryDeleter
}

// categoryService implements all category use cases
type categoryService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewService creates a new category service
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &categoryService{
		repo:   repo,
		logger: logger,
	}
}
