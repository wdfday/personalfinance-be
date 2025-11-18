package service

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/identify/broker/client"
	"personalfinancedss/internal/module/identify/broker/client/okx"
	"personalfinancedss/internal/module/identify/broker/client/sepay"
	"personalfinancedss/internal/module/identify/broker/client/ssi"
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/repository"
	"personalfinancedss/internal/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type brokerConnectionService struct {
	repo              repository.BrokerConnectionRepository
	encryptionService service.EncryptionService
	ssiClient         *ssi.Client
	okxClient         *okx.Client
	sepayClient       *sepay.Client
}

// NewBrokerConnectionService creates a new broker connection service
func NewBrokerConnectionService(
	repo repository.BrokerConnectionRepository,
	encryptionService service.EncryptionService,
	ssiClient *ssi.Client,
	okxClient *okx.Client,
	sepayClient *sepay.Client,
) BrokerConnectionService {
	return &brokerConnectionService{
		repo:              repo,
		encryptionService: encryptionService,
		ssiClient:         ssiClient,
		okxClient:         okxClient,
		sepayClient:       sepayClient,
	}
}

func (s *brokerConnectionService) Create(ctx context.Context, req *CreateBrokerConnectionRequest) (*domain.BrokerConnection, error) {
	// Get broker client
	brokerClient, err := s.getBrokerClient(req.BrokerType)
	if err != nil {
		return nil, err
	}

	// Build credentials
	credentials := client.Credentials{
		APIKey:         req.APIKey,
		APISecret:      req.APISecret,
		Passphrase:     req.Passphrase,
		ConsumerID:     req.ConsumerID,
		ConsumerSecret: req.ConsumerSecret,
		OTPCode:        req.OTPCode,
		OTPMethod:      req.OTPMethod,
	}

	// Authenticate with broker
	authResp, err := brokerClient.Authenticate(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with broker: %w", err)
	}

	// Get portfolio info to validate connection and get external account info
	portfolio, err := brokerClient.GetPortfolio(ctx, authResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio from broker: %w", err)
	}

	// Encrypt credentials
	encryptedAPIKey, err := s.encryptionService.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}

	encryptedAPISecret, err := s.encryptionService.Encrypt(req.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API secret: %w", err)
	}

	// Encrypt access and refresh tokens
	encryptedAccessToken, err := s.encryptionService.Encrypt(authResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken, err := s.encryptionService.Encrypt(authResp.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	// Create broker connection entity
	connection := &domain.BrokerConnection{
		ID:         uuid.New(),
		UserID:     req.UserID,
		BrokerType: req.BrokerType,
		BrokerName: req.BrokerName,
		Status:     domain.BrokerConnectionStatusActive,

		// Encrypted credentials
		APIKey:    encryptedAPIKey,
		APISecret: encryptedAPISecret,

		// Encrypted tokens
		AccessToken:     &encryptedAccessToken,
		RefreshToken:    &encryptedRefreshToken,
		TokenExpiresAt:  &authResp.ExpiresAt,
		LastRefreshedAt: new(time.Time),

		// Sync settings
		AutoSync:         req.AutoSync,
		SyncFrequency:    req.SyncFrequency,
		SyncAssets:       req.SyncAssets,
		SyncTransactions: req.SyncTransactions,
		SyncPrices:       req.SyncPrices,
		SyncBalance:      req.SyncBalance,

		// External account info (from portfolio)
		ExternalAccountName: &portfolio.Currency,

		Notes: req.Notes,
	}
	*connection.LastRefreshedAt = time.Now()

	// Encrypt optional fields
	if req.Passphrase != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt passphrase: %w", err)
		}
		connection.Passphrase = &encrypted
	}

	if req.ConsumerID != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.ConsumerID)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt consumer ID: %w", err)
		}
		connection.ConsumerID = &encrypted
	}

	if req.ConsumerSecret != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.ConsumerSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt consumer secret: %w", err)
		}
		connection.ConsumerSecret = &encrypted
	}

	if req.OTPMethod != nil {
		connection.OTPMethod = req.OTPMethod // OTP method is not sensitive
	}

	// Save to database
	if err := s.repo.Create(ctx, connection); err != nil {
		return nil, fmt.Errorf("failed to save broker connection: %w", err)
	}

	return connection, nil
}

func (s *brokerConnectionService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.BrokerConnection, error) {
	connection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("broker connection not found")
		}
		return nil, err
	}

	// Verify ownership
	if connection.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to broker connection")
	}

	return connection, nil
}

func (s *brokerConnectionService) List(ctx context.Context, userID uuid.UUID, filters *ListFilters) ([]*domain.BrokerConnection, error) {
	if filters == nil {
		filters = &ListFilters{}
	}

	// Get all connections for user
	connections, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := make([]*domain.BrokerConnection, 0)
	for _, conn := range connections {
		if !s.matchesFilters(conn, filters) {
			continue
		}
		filtered = append(filtered, conn)
	}

	return filtered, nil
}

func (s *brokerConnectionService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateBrokerConnectionRequest) (*domain.BrokerConnection, error) {
	// Get existing connection
	connection, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.BrokerName != nil {
		connection.BrokerName = *req.BrokerName
	}

	if req.AutoSync != nil {
		connection.AutoSync = *req.AutoSync
	}

	if req.SyncFrequency != nil {
		connection.SyncFrequency = *req.SyncFrequency
	}

	if req.SyncAssets != nil {
		connection.SyncAssets = *req.SyncAssets
	}

	if req.SyncTransactions != nil {
		connection.SyncTransactions = *req.SyncTransactions
	}

	if req.SyncPrices != nil {
		connection.SyncPrices = *req.SyncPrices
	}

	if req.SyncBalance != nil {
		connection.SyncBalance = *req.SyncBalance
	}

	if req.Notes != nil {
		connection.Notes = req.Notes
	}

	// Update credentials if provided
	if req.APIKey != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		connection.APIKey = encrypted
	}

	if req.APISecret != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.APISecret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API secret: %w", err)
		}
		connection.APISecret = encrypted
	}

	if req.Passphrase != nil {
		encrypted, err := s.encryptionService.Encrypt(*req.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt passphrase: %w", err)
		}
		connection.Passphrase = &encrypted
	}

	// Save updates
	if err := s.repo.Update(ctx, connection); err != nil {
		return nil, fmt.Errorf("failed to update broker connection: %w", err)
	}

	return connection, nil
}

func (s *brokerConnectionService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Verify ownership
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *brokerConnectionService) Activate(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get connection
	connection, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	connection.Status = domain.BrokerConnectionStatusActive
	return s.repo.Update(ctx, connection)
}

func (s *brokerConnectionService) Deactivate(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get connection
	connection, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	connection.Status = domain.BrokerConnectionStatusDisconnected
	return s.repo.Update(ctx, connection)
}

func (s *brokerConnectionService) RefreshToken(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.BrokerConnection, error) {
	// Get connection
	connection, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Get broker client
	brokerClient, err := s.getBrokerClient(connection.BrokerType)
	if err != nil {
		return nil, err
	}

	// Decrypt refresh token
	if connection.RefreshToken == nil {
		return nil, fmt.Errorf("no refresh token available")
	}

	refreshToken, err := s.encryptionService.Decrypt(*connection.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Refresh token with broker
	authResp, err := brokerClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update connection with new token
	connection.RefreshAccessToken(authResp.AccessToken, authResp.ExpiresIn)

	// Encrypt new tokens
	encryptedAccessToken, err := s.encryptionService.Encrypt(authResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt new access token: %w", err)
	}

	encryptedRefreshToken, err := s.encryptionService.Encrypt(authResp.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt new refresh token: %w", err)
	}

	connection.AccessToken = &encryptedAccessToken
	connection.RefreshToken = &encryptedRefreshToken

	// Save to database
	if err := s.repo.Update(ctx, connection); err != nil {
		return nil, fmt.Errorf("failed to update tokens: %w", err)
	}

	return connection, nil
}

func (s *brokerConnectionService) TestConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get connection
	connection, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Get broker client
	brokerClient, err := s.getBrokerClient(connection.BrokerType)
	if err != nil {
		return err
	}

	// Decrypt access token
	if connection.AccessToken == nil {
		return fmt.Errorf("no access token available")
	}

	accessToken, err := s.encryptionService.Decrypt(*connection.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	// Try to get portfolio to test connection
	_, err = brokerClient.GetPortfolio(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

func (s *brokerConnectionService) SyncNow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*SyncResult, error) {
	// This will be implemented in the sync service
	// For now, just validate the connection exists
	_, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return &SyncResult{
		Success: false,
		Error:   stringPtr("sync service not yet implemented"),
	}, fmt.Errorf("sync service not yet implemented")
}

// Helper methods

func (s *brokerConnectionService) getBrokerClient(brokerType domain.BrokerType) (client.BrokerClient, error) {
	switch brokerType {
	case domain.BrokerTypeSSI:
		return s.ssiClient, nil
	case domain.BrokerTypeOKX:
		return s.okxClient, nil
	case domain.BrokerTypeSePay:
		return s.sepayClient, nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", brokerType)
	}
}

func (s *brokerConnectionService) matchesFilters(conn *domain.BrokerConnection, filters *ListFilters) bool {
	if filters.BrokerType != nil && conn.BrokerType != *filters.BrokerType {
		return false
	}

	if filters.Status != nil && conn.Status != *filters.Status {
		return false
	}

	if filters.ActiveOnly && conn.Status != domain.BrokerConnectionStatusActive {
		return false
	}

	if filters.AutoSyncOnly && !conn.AutoSync {
		return false
	}

	if filters.NeedingSyncOnly && !conn.NeedsSync() {
		return false
	}

	return true
}

func stringPtr(s string) *string {
	return &s
}
