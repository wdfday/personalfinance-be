package service

import (
	"context"
	accountservice "personalfinancedss/internal/module/cashflow/account/service"
	categoryservice "personalfinancedss/internal/module/cashflow/category/service"
	profileservice "personalfinancedss/internal/module/identify/profile/service"
	"personalfinancedss/internal/module/identify/user/domain"
	"personalfinancedss/internal/module/identify/user/repository"
	"personalfinancedss/internal/shared"
	"time"

	"go.uber.org/zap"
)

// UserCreator interface for user creation use case
type UserCreator interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
}

// UserReader defines the interface for reading user
type UserReader interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, filter domain.ListUsersFilter, pagination shared.Pagination) (shared.Page[domain.User], error)
}

// UserUpdater defines the interface for updating user
type UserUpdater interface {
	Update(ctx context.Context, user *domain.User) error
	UpdateColumns(ctx context.Context, id string, cols map[string]any) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id string, at time.Time, ip *string) error
}

// UserDeleter defines the interface for deleting user
type UserDeleter interface {
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
}

// UserExistenceChecker defines the interface for checking user existence
type UserExistenceChecker interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// IUserService represents the complete user service interface
// It composes all use case interfaces for convenience
type IUserService interface {
	UserCreator
	UserReader
	UserUpdater
	UserDeleter
	UserExistenceChecker

	// Auth-related methods
	MarkEmailVerified(ctx context.Context, id string, at time.Time) error
	IncLoginAttempts(ctx context.Context, id string) error
	ResetLoginAttempts(ctx context.Context, id string) error
	SetLockedUntil(ctx context.Context, id string, until *time.Time) error
}

type UserService struct {
	repo            repository.Repository
	profileService  profileservice.Service
	categoryService categoryservice.Service
	accountService  accountservice.Service
	logger          *zap.Logger
}

func NewUserService(
	repo repository.Repository,
	profileService profileservice.Service,
	categoryService categoryservice.Service,
	accountService accountservice.Service,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		repo:            repo,
		profileService:  profileService,
		categoryService: categoryService,
		accountService:  accountService,
		logger:          logger,
	}
}
