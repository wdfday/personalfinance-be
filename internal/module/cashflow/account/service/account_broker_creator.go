package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"personalfinancedss/internal/broker/client"
	"personalfinancedss/internal/module/cashflow/account/domain"
	accountdto "personalfinancedss/internal/module/cashflow/account/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateAccountWithBroker creates a new account with broker integration
// This method will:
// 1. Validate broker credentials by calling broker API
// 2. Fetch account/portfolio info from broker
// 3. Create account with real data from broker
// 4. Save encrypted broker credentials
func (s *accountService) CreateAccountWithBroker(
	ctx context.Context,
	userID string,
	req accountdto.CreateAccountWithBrokerRequest,
) (*domain.Account, error) {
	s.logger.Info("Starting create account with broker",
		zap.String("user_id", userID),
		zap.String("broker_type", string(req.BrokerType)),
		zap.String("account_name", req.AccountName),
	)

	// Validate request
	if err := req.Validate(); err != nil {
		s.logger.Error("Request validation failed",
			zap.Error(err),
			zap.String("broker_type", string(req.BrokerType)),
		)
		return nil, shared.ErrBadRequest.WithDetails("validation", err.Error())
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Error("Invalid user ID format",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, shared.ErrBadRequest.WithDetails("field", "user_id").WithDetails("reason", "invalid UUID format")
	}

	// Get broker client
	var brokerClient client.BrokerClient
	switch req.BrokerType {
	case domain.BrokerTypeSSI:
		brokerClient = s.getBrokerClient(domain.BrokerTypeSSI)
		s.logger.Debug("Using SSI broker client")
	case domain.BrokerTypeOKX:
		brokerClient = s.getBrokerClient(domain.BrokerTypeOKX)
		s.logger.Debug("Using OKX broker client")
	default:
		s.logger.Error("Unsupported broker type",
			zap.String("broker_type", string(req.BrokerType)),
		)
		return nil, shared.ErrBadRequest.WithDetails("broker_type", "unsupported")
	}

	// Prepare credentials for authentication
	credentials := client.Credentials{}

	if req.BrokerType == domain.BrokerTypeOKX {
		if req.APIKey != nil {
			credentials.APIKey = *req.APIKey
		}
		if req.APISecret != nil {
			credentials.APISecret = *req.APISecret
		}
		credentials.Passphrase = req.Passphrase
	} else if req.BrokerType == domain.BrokerTypeSSI {
		if req.ConsumerID != nil {
			credentials.APIKey = *req.ConsumerID
		}
		if req.ConsumerSecret != nil {
			credentials.APISecret = *req.ConsumerSecret
		}
		credentials.ConsumerID = req.ConsumerID
		credentials.ConsumerSecret = req.ConsumerSecret
		credentials.OTPCode = req.OTPCode
		credentials.OTPMethod = req.OTPMethod
	}

	// Step 1: Authenticate with broker to validate credentials
	s.logger.Info("Authenticating with broker",
		zap.String("broker_type", string(req.BrokerType)),
	)

	authResp, err := brokerClient.Authenticate(ctx, credentials)
	if err != nil {
		s.logger.Error("Broker authentication failed",
			zap.Error(err),
			zap.String("broker_type", string(req.BrokerType)),
		)
		return nil, shared.ErrBadRequest.WithDetails("broker_auth", fmt.Sprintf("Failed to authenticate with broker: %v", err))
	}

	s.logger.Info("Broker authentication successful",
		zap.String("broker_type", string(req.BrokerType)),
		zap.Time("token_expires_at", authResp.ExpiresAt),
	)

	// Step 2: Get portfolio/account info from broker
	s.logger.Info("Fetching portfolio from broker",
		zap.String("broker_type", string(req.BrokerType)),
	)

	portfolio, err := brokerClient.GetPortfolio(ctx, authResp.AccessToken)
	if err != nil {
		// If portfolio fetch fails, we can still create account with default values
		s.logger.Warn("Failed to fetch portfolio, will use default values",
			zap.Error(err),
			zap.String("broker_type", string(req.BrokerType)),
		)
	} else {
		s.logger.Info("Portfolio fetched successfully",
			zap.String("broker_type", string(req.BrokerType)),
			zap.Float64("total_value", portfolio.TotalValue),
			zap.Float64("cash_balance", portfolio.CashBalance),
			zap.String("currency", portfolio.Currency),
		)
	}

	// Step 3: Create account with info from broker
	accountType, err := parseAccountType(req.AccountType)
	if err != nil {
		s.logger.Error("Invalid account type",
			zap.Error(err),
			zap.String("account_type", req.AccountType),
		)
		return nil, err
	}

	account := &domain.Account{
		UserID:            userUUID,
		AccountName:       strings.TrimSpace(req.AccountName),
		AccountType:       accountType,
		CurrentBalance:    0,
		Currency:          domain.CurrencyVND, // Default
		IsActive:          true,
		IsPrimary:         false,
		IncludeInNetWorth: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Set institution name
	if req.InstitutionName != nil {
		account.InstitutionName = normalizeString(*req.InstitutionName)
	} else if req.BrokerName != "" {
		institutionName := req.BrokerName
		account.InstitutionName = &institutionName
	}

	// Update with portfolio data if available
	if portfolio != nil {
		account.CurrentBalance = portfolio.TotalValue
		availableBalance := portfolio.CashBalance
		account.AvailableBalance = &availableBalance

		if portfolio.Currency != "" {
			account.Currency = domain.Currency(strings.ToUpper(portfolio.Currency))
		}

		s.logger.Debug("Account populated with portfolio data",
			zap.Float64("balance", account.CurrentBalance),
			zap.String("currency", string(account.Currency)),
		)
	}

	// Step 4: Build broker integration with encrypted credentials
	s.logger.Debug("Building broker integration")

	integration, err := s.buildBrokerIntegration(req, authResp)
	if err != nil {
		s.logger.Error("Failed to build broker integration",
			zap.Error(err),
		)
		return nil, shared.ErrInternal.WithDetails("broker_integration", err.Error())
	}

	// Set broker integration
	if err := account.SetBrokerIntegration(integration); err != nil {
		s.logger.Error("Failed to set broker integration",
			zap.Error(err),
		)
		return nil, shared.ErrInternal.WithDetails("set_broker_integration", err.Error())
	}

	// Step 5: Save account to database
	s.logger.Info("Saving account to database",
		zap.String("account_name", account.AccountName),
		zap.String("broker_type", string(req.BrokerType)),
	)

	if err := s.repo.Create(ctx, account); err != nil {
		s.logger.Error("Failed to save account to database",
			zap.Error(err),
			zap.String("account_name", account.AccountName),
		)
		return nil, shared.ErrInternal.WithError(err)
	}

	s.logger.Info("Account with broker created successfully",
		zap.String("account_id", account.ID.String()),
		zap.String("account_name", account.AccountName),
		zap.String("broker_type", string(req.BrokerType)),
		zap.Float64("initial_balance", account.CurrentBalance),
	)

	return account, nil
}

// buildBrokerIntegration creates broker integration from request and auth response
func (s *accountService) buildBrokerIntegration(
	req accountdto.CreateAccountWithBrokerRequest,
	authResp *client.AuthResponse,
) (*domain.BrokerIntegration, error) {
	s.logger.Debug("Building broker integration with encryption")

	syncFrequency := req.SyncFrequency
	if syncFrequency == 0 {
		syncFrequency = 60 // Default 60 minutes
	}

	now := time.Now()
	expiresAt := authResp.ExpiresAt

	integration := domain.NewBrokerIntegration(req.BrokerType)
	integration.BrokerName = req.BrokerName
	integration.AutoSync = req.AutoSync
	integration.SyncFrequency = syncFrequency
	integration.SyncAssets = req.SyncAssets
	integration.SyncTransactions = req.SyncTransactions
	integration.SyncPrices = req.SyncPrices
	integration.SyncBalance = req.SyncBalance

	// Encrypt and set tokens
	s.logger.Debug("Encrypting access and refresh tokens")

	encryptedAccessToken, err := s.encryptionService.EncryptIfNotEmpty(&authResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to encrypt access token", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}
	integration.AccessToken = encryptedAccessToken

	encryptedRefreshToken, err := s.encryptionService.EncryptIfNotEmpty(&authResp.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to encrypt refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
	}
	integration.RefreshToken = encryptedRefreshToken

	integration.TokenExpiresAt = &expiresAt
	integration.LastRefreshedAt = &now

	// Encrypt and set credentials based on broker type
	if req.BrokerType == domain.BrokerTypeOKX {
		s.logger.Debug("Encrypting OKX credentials")

		encryptedAPIKey, err := s.encryptionService.EncryptIfNotEmpty(req.APIKey)
		if err != nil {
			s.logger.Error("Failed to encrypt OKX API key", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt OKX API key: %w", err)
		}
		integration.OKXAPIKey = encryptedAPIKey

		encryptedAPISecret, err := s.encryptionService.EncryptIfNotEmpty(req.APISecret)
		if err != nil {
			s.logger.Error("Failed to encrypt OKX API secret", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt OKX API secret: %w", err)
		}
		integration.OKXAPISecret = encryptedAPISecret

		encryptedPassphrase, err := s.encryptionService.EncryptIfNotEmpty(req.Passphrase)
		if err != nil {
			s.logger.Error("Failed to encrypt OKX passphrase", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt OKX passphrase: %w", err)
		}
		integration.OKXPassphrase = encryptedPassphrase

		s.logger.Info("OKX credentials encrypted successfully")

	} else if req.BrokerType == domain.BrokerTypeSSI {
		s.logger.Debug("Encrypting SSI credentials")

		encryptedConsumerID, err := s.encryptionService.EncryptIfNotEmpty(req.ConsumerID)
		if err != nil {
			s.logger.Error("Failed to encrypt SSI consumer ID", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt SSI consumer ID: %w", err)
		}
		integration.SSIConsumerID = encryptedConsumerID

		encryptedConsumerSecret, err := s.encryptionService.EncryptIfNotEmpty(req.ConsumerSecret)
		if err != nil {
			s.logger.Error("Failed to encrypt SSI consumer secret", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt SSI consumer secret: %w", err)
		}
		integration.SSIConsumerSecret = encryptedConsumerSecret

		// OTP method is not sensitive, no need to encrypt
		integration.SSIOTPMethod = req.OTPMethod

		s.logger.Info("SSI credentials encrypted successfully")
	}

	s.logger.Debug("Broker integration built successfully with all credentials encrypted")
	return integration, nil
}

// getBrokerClient returns the appropriate broker client
// TODO: This should use dependency injection
func (s *accountService) getBrokerClient(brokerType domain.BrokerType) client.BrokerClient {
	// This is a temporary solution
	// In production, inject broker clients or syncService
	switch brokerType {
	case domain.BrokerTypeSSI:
		return s.ssiClient // Need to add this field
	case domain.BrokerTypeOKX:
		return s.okxClient // Need to add this field
	default:
		return nil
	}
}
