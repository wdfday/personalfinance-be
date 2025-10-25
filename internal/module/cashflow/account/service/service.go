package service

import (
	"context"
	"personalfinancedss/internal/broker/client"
	"personalfinancedss/internal/module/cashflow/account/domain"
	accountdto "personalfinancedss/internal/module/cashflow/account/dto"
	"personalfinancedss/internal/module/cashflow/account/repository"
	encryptionService "personalfinancedss/internal/service"

	"go.uber.org/zap"
)

// AccountCreator defines account creation operations
type AccountCreator interface {
	CreateAccount(ctx context.Context, userID string, req accountdto.CreateAccountRequest) (*domain.Account, error)
	CreateDefaultCashAccount(ctx context.Context, userID string) error
	CreateAccountWithBroker(ctx context.Context, userID string, req accountdto.CreateAccountWithBrokerRequest) (*domain.Account, error)
}

// AccountReader defines account read operations
type AccountReader interface {
	GetByID(ctx context.Context, id, userID string) (*domain.Account, error)
	GetByUserID(ctx context.Context, userID string, req accountdto.ListAccountsRequest) ([]domain.Account, int64, error)
}

// AccountUpdater defines account update operations
type AccountUpdater interface {
	UpdateAccount(ctx context.Context, id, userID string, req accountdto.UpdateAccountRequest) (*domain.Account, error)
	unsetPrimaryAccount(ctx context.Context, userID string) error
}

// AccountDeleter defines account delete operations
type AccountDeleter interface {
	DeleteAccount(ctx context.Context, id, userID string) error
}

// Service is the composite interface for all account operations
type Service interface {
	AccountCreator
	AccountReader
	AccountUpdater
	AccountDeleter
}

// accountService implements all account use cases
type accountService struct {
	repo              repository.Repository
	ssiClient         client.BrokerClient
	okxClient         client.BrokerClient
	encryptionService *encryptionService.EncryptionService
	logger            *zap.Logger
}

// NewService creates a new account service
func NewService(
	repo repository.Repository,
	ssiClient client.BrokerClient,
	okxClient client.BrokerClient,
	encryptionSvc *encryptionService.EncryptionService,
	logger *zap.Logger,
) Service {
	return &accountService{
		repo:              repo,
		ssiClient:         ssiClient,
		okxClient:         okxClient,
		encryptionService: encryptionSvc,
		logger:            logger.Named("account.service"),
	}
}
