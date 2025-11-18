package dto

import (
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/service"

	"github.com/google/uuid"
)

// ToCreateServiceRequest converts API request to service request
func (r *CreateBrokerConnectionRequest) ToServiceRequest(userID uuid.UUID) *service.CreateBrokerConnectionRequest {
	// Apply defaults
	r.ApplyDefaults()

	return &service.CreateBrokerConnectionRequest{
		UserID:           userID,
		BrokerType:       domain.BrokerType(r.BrokerType),
		BrokerName:       r.BrokerName,
		APIKey:           r.APIKey,
		APISecret:        r.APISecret,
		Passphrase:       r.Passphrase,
		ConsumerID:       r.ConsumerID,
		ConsumerSecret:   r.ConsumerSecret,
		OTPCode:          r.OTPCode,
		OTPMethod:        r.OTPMethod,
		AutoSync:         *r.AutoSync,
		SyncFrequency:    *r.SyncFrequency,
		SyncAssets:       *r.SyncAssets,
		SyncTransactions: *r.SyncTransactions,
		SyncPrices:       *r.SyncPrices,
		SyncBalance:      *r.SyncBalance,
		Notes:            r.Notes,
	}
}

// ToUpdateServiceRequest converts API request to service request
func (r *UpdateBrokerConnectionRequest) ToServiceRequest() *service.UpdateBrokerConnectionRequest {
	return &service.UpdateBrokerConnectionRequest{
		BrokerName:       r.BrokerName,
		APIKey:           r.APIKey,
		APISecret:        r.APISecret,
		Passphrase:       r.Passphrase,
		ConsumerID:       r.ConsumerID,
		ConsumerSecret:   r.ConsumerSecret,
		OTPMethod:        r.OTPMethod,
		AutoSync:         r.AutoSync,
		SyncFrequency:    r.SyncFrequency,
		SyncAssets:       r.SyncAssets,
		SyncTransactions: r.SyncTransactions,
		SyncPrices:       r.SyncPrices,
		SyncBalance:      r.SyncBalance,
		Notes:            r.Notes,
	}
}

// ToListServiceFilters converts query parameters to service filters
func (q *ListBrokerConnectionsQuery) ToServiceFilters() *service.ListFilters {
	filters := &service.ListFilters{
		AutoSyncOnly:    q.AutoSyncOnly,
		ActiveOnly:      q.ActiveOnly,
		NeedingSyncOnly: q.NeedingSyncOnly,
	}

	if q.BrokerType != nil {
		brokerType := domain.BrokerType(*q.BrokerType)
		filters.BrokerType = &brokerType
	}

	if q.Status != nil {
		status := domain.BrokerConnectionStatus(*q.Status)
		filters.Status = &status
	}

	return filters
}

// ToBrokerConnectionResponse converts domain model to API response
func ToBrokerConnectionResponse(conn *domain.BrokerConnection) *BrokerConnectionResponse {
	return &BrokerConnectionResponse{
		ID:                    conn.ID,
		UserID:                conn.UserID,
		BrokerType:            string(conn.BrokerType),
		BrokerName:            conn.BrokerName,
		Status:                string(conn.Status),
		TokenExpiresAt:        conn.TokenExpiresAt,
		LastRefreshedAt:       conn.LastRefreshedAt,
		IsTokenValid:          conn.IsTokenValid(),
		AutoSync:              conn.AutoSync,
		SyncFrequency:         conn.SyncFrequency,
		SyncAssets:            conn.SyncAssets,
		SyncTransactions:      conn.SyncTransactions,
		SyncPrices:            conn.SyncPrices,
		SyncBalance:           conn.SyncBalance,
		LastSyncAt:            conn.LastSyncAt,
		LastSyncStatus:        conn.LastSyncStatus,
		LastSyncError:         conn.LastSyncError,
		TotalSyncs:            conn.TotalSyncs,
		SuccessfulSyncs:       conn.SuccessfulSyncs,
		FailedSyncs:           conn.FailedSyncs,
		ExternalAccountID:     conn.ExternalAccountID,
		ExternalAccountNumber: conn.ExternalAccountNumber,
		ExternalAccountName:   conn.ExternalAccountName,
		Notes:                 conn.Notes,
		CreatedAt:             conn.CreatedAt,
		UpdatedAt:             conn.UpdatedAt,
	}
}

// ToBrokerConnectionListResponse converts a list of domain models to API response
func ToBrokerConnectionListResponse(connections []*domain.BrokerConnection) *BrokerConnectionListResponse {
	responses := make([]*BrokerConnectionResponse, len(connections))
	for i, conn := range connections {
		responses[i] = ToBrokerConnectionResponse(conn)
	}

	return &BrokerConnectionListResponse{
		Connections: responses,
		Total:       len(connections),
	}
}

// ToSyncResultResponse converts service sync result to API response
func ToSyncResultResponse(result *service.SyncResult) *SyncResultResponse {
	return &SyncResultResponse{
		Success:            result.Success,
		SyncedAt:           result.SyncedAt,
		AssetsCount:        result.AssetsCount,
		TransactionsCount:  result.TransactionsCount,
		UpdatedPricesCount: result.UpdatedPricesCount,
		BalanceUpdated:     result.BalanceUpdated,
		Error:              result.Error,
		Details:            result.Details,
	}
}

// GetBrokerProviders returns information about available broker providers
func GetBrokerProviders() *ListBrokerProvidersResponse {
	return &ListBrokerProvidersResponse{
		Providers: []BrokerProviderInfo{
			{
				BrokerType:  string(domain.BrokerTypeSSI),
				DisplayName: "SSI Securities",
				Description: "Vietnam stock trading platform",
				RequiredFields: []string{
					"consumer_id",
					"consumer_secret",
					"otp_code",
					"otp_method",
				},
				SupportedFeatures: []string{
					"portfolio",
					"positions",
					"transactions",
					"market_prices",
				},
			},
			{
				BrokerType:  string(domain.BrokerTypeOKX),
				DisplayName: "OKX Exchange",
				Description: "Cryptocurrency trading platform",
				RequiredFields: []string{
					"api_key",
					"api_secret",
					"passphrase",
				},
				SupportedFeatures: []string{
					"portfolio",
					"positions",
					"transactions",
					"market_prices",
				},
			},
			{
				BrokerType:  string(domain.BrokerTypeSePay),
				DisplayName: "SePay",
				Description: "Vietnam banking and payment API",
				RequiredFields: []string{
					"api_key",
				},
				SupportedFeatures: []string{
					"portfolio",
					"transactions",
					"balance",
				},
			},
		},
	}
}
