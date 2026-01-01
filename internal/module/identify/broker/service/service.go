package service

import (
	"context"
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/dto"
	"time"

	"github.com/google/uuid"
)

// BrokerConnectionService defines the business logic for broker connections
type BrokerConnectionService interface {
	// Create creates a new broker connection and authenticates with the broker
	Create(ctx context.Context, req *dto.CreateBrokerConnectionServiceRequest) (*domain.BrokerConnection, error)

	// GetByID retrieves a broker connection by ID
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.BrokerConnection, error)

	// List retrieves all broker connections for a user
	List(ctx context.Context, userID uuid.UUID, filters *ListFilters) ([]*domain.BrokerConnection, error)

	// Update updates a broker connection
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateBrokerConnectionRequest) (*domain.BrokerConnection, error)

	// Delete soft-deletes a broker connection
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Activate activates a broker connection
	Activate(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Deactivate deactivates a broker connection
	Deactivate(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// RefreshToken refreshes the access token for a broker connection
	RefreshToken(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.BrokerConnection, error)

	// TestConnection tests the broker connection by authenticating
	TestConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// SyncNow manually triggers a sync for a broker connection
	SyncNow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*SyncResult, error)
}

// UpdateBrokerConnectionRequest represents the request to update a broker connection
type UpdateBrokerConnectionRequest struct {
	BrokerName *string

	// Update credentials (optional - will be encrypted)
	APIKey         *string
	APISecret      *string
	Passphrase     *string
	ConsumerID     *string
	ConsumerSecret *string
	OTPMethod      *string

	// Sync settings
	AutoSync         *bool
	SyncFrequency    *int
	SyncAssets       *bool
	SyncTransactions *bool
	SyncPrices       *bool
	SyncBalance      *bool

	Notes *string
}

// ListFilters represents filters for listing broker connections
type ListFilters struct {
	BrokerType      *domain.BrokerType
	Status          *domain.BrokerConnectionStatus
	AutoSyncOnly    bool
	ActiveOnly      bool
	NeedingSyncOnly bool
}

// SyncResult represents the result of a manual sync operation
type SyncResult struct {
	Success            bool
	SyncedAt           time.Time
	AssetsCount        int
	TransactionsCount  int
	UpdatedPricesCount int
	BalanceUpdated     bool
	Error              *string
	Details            map[string]interface{}
}
