package ssi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"personalfinancedss/internal/broker/client"
)

const (
	SSIBaseURL    = "https://fc-data.ssi.com.vn"
	SSITradingURL = "https://fc-trading.ssi.com.vn"
	SSIAuthURL    = "https://id.ssi.com.vn/connect/token"
)

// SSIClient implements the BrokerClient interface for SSI Securities
type SSIClient struct {
	httpClient *http.Client
	baseURL    string
	tradingURL string
	authURL    string
}

// NewSSIClient creates a new SSI client
func NewSSIClient() *SSIClient {
	return &SSIClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    SSIBaseURL,
		tradingURL: SSITradingURL,
		authURL:    SSIAuthURL,
	}
}

// Authenticate authenticates with SSI and returns an access token
func (c *SSIClient) Authenticate(ctx context.Context, credentials client.Credentials) (*client.AuthResponse, error) {
	if credentials.ConsumerID == nil || credentials.ConsumerSecret == nil {
		return nil, fmt.Errorf("consumerID and consumerSecret are required for SSI")
	}

	// Prepare request body
	data := map[string]string{
		"grant_type":    "password",
		"client_id":     *credentials.ConsumerID,
		"client_secret": *credentials.ConsumerSecret,
	}

	// Add OTP if provided
	if credentials.OTPCode != nil {
		data["pin"] = *credentials.OTPCode
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return &client.AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		TokenType:    authResp.TokenType,
	}, nil
}

// RefreshToken refreshes the SSI access token
func (c *SSIClient) RefreshToken(ctx context.Context, refreshToken string) (*client.AuthResponse, error) {
	data := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &client.AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		TokenType:    authResp.TokenType,
	}, nil
}

// GetPortfolio retrieves portfolio information from SSI
func (c *SSIClient) GetPortfolio(ctx context.Context, accessToken string) (*client.Portfolio, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.tradingURL+"/api/v2/portfolio", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("portfolio request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var portfolioResp struct {
		TotalValue     float64 `json:"totalValue"`
		CashBalance    float64 `json:"cashBalance"`
		StockValue     float64 `json:"stockValue"`
		UnrealizedGain float64 `json:"unrealizedGain"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&portfolioResp); err != nil {
		return nil, fmt.Errorf("failed to decode portfolio response: %w", err)
	}

	return &client.Portfolio{
		TotalValue:     portfolioResp.TotalValue,
		CashBalance:    portfolioResp.CashBalance,
		UnrealizedGain: portfolioResp.UnrealizedGain,
		Currency:       "VND",
		AssetAllocation: map[string]float64{
			"stock": portfolioResp.StockValue,
			"cash":  portfolioResp.CashBalance,
		},
		LastUpdated: time.Now(),
	}, nil
}

// GetPositions retrieves current positions from SSI
func (c *SSIClient) GetPositions(ctx context.Context, accessToken string) ([]client.Position, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.tradingURL+"/api/v2/positions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create positions request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("positions request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var positionsResp []struct {
		Symbol          string  `json:"symbol"`
		Name            string  `json:"stockName"`
		Quantity        float64 `json:"totalQty"`
		AvgPrice        float64 `json:"avgPrice"`
		CurrentPrice    float64 `json:"marketPrice"`
		MarketValue     float64 `json:"marketValue"`
		UnrealizedPL    float64 `json:"unrealizedPL"`
		UnrealizedPLPct float64 `json:"unrealizedPLPct"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&positionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode positions response: %w", err)
	}

	positions := make([]client.Position, 0, len(positionsResp))
	for _, pos := range positionsResp {
		positions = append(positions, client.Position{
			Symbol:             pos.Symbol,
			Name:               pos.Name,
			AssetType:          "stock",
			Quantity:           pos.Quantity,
			AverageCostPerUnit: pos.AvgPrice,
			CurrentPrice:       pos.CurrentPrice,
			CurrentValue:       pos.MarketValue,
			UnrealizedGain:     pos.UnrealizedPL,
			UnrealizedGainPct:  pos.UnrealizedPLPct,
			Currency:           "VND",
			Exchange:           "HOSE", // Default to HOSE, should be determined from symbol
			LastUpdated:        time.Now(),
		})
	}

	return positions, nil
}

// GetTransactions retrieves transaction history from SSI
func (c *SSIClient) GetTransactions(ctx context.Context, accessToken string, startDate, endDate time.Time) ([]client.Transaction, error) {
	url := fmt.Sprintf("%s/api/v2/transactions?startDate=%s&endDate=%s",
		c.tradingURL,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactions request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("transactions request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var transactionsResp []struct {
		ID              string  `json:"id"`
		Symbol          string  `json:"symbol"`
		Side            string  `json:"side"` // buy/sell
		Quantity        float64 `json:"quantity"`
		Price           float64 `json:"price"`
		Amount          float64 `json:"amount"`
		Fee             float64 `json:"fee"`
		Tax             float64 `json:"tax"`
		TransactionDate string  `json:"transactionDate"`
		Status          string  `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&transactionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode transactions response: %w", err)
	}

	transactions := make([]client.Transaction, 0, len(transactionsResp))
	for _, trans := range transactionsResp {
		transDate, _ := time.Parse("2006-01-02", trans.TransactionDate)

		transType := "buy"
		if trans.Side == "sell" || trans.Side == "SELL" {
			transType = "sell"
		}

		transactions = append(transactions, client.Transaction{
			ExternalID:      trans.ID,
			TransactionType: transType,
			Symbol:          trans.Symbol,
			Quantity:        trans.Quantity,
			Price:           trans.Price,
			Amount:          trans.Amount,
			Fee:             trans.Fee,
			Tax:             trans.Tax,
			Currency:        "VND",
			TransactionDate: transDate,
			Status:          trans.Status,
		})
	}

	return transactions, nil
}

// GetMarketPrice retrieves current market price for a symbol
func (c *SSIClient) GetMarketPrice(ctx context.Context, symbol string) (*client.MarketPrice, error) {
	url := fmt.Sprintf("%s/api/v2/market/securities/%s", c.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create market price request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get market price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market price request failed with status %d", resp.StatusCode)
	}

	var priceResp struct {
		Symbol    string  `json:"symbol"`
		Price     float64 `json:"lastPrice"`
		Change    float64 `json:"change"`
		ChangePct float64 `json:"changePct"`
		Volume    float64 `json:"totalVolume"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return nil, fmt.Errorf("failed to decode market price response: %w", err)
	}

	return &client.MarketPrice{
		Symbol:      priceResp.Symbol,
		Price:       priceResp.Price,
		Change:      priceResp.Change,
		ChangePct:   priceResp.ChangePct,
		Volume:      priceResp.Volume,
		Currency:    "VND",
		LastUpdated: time.Now(),
	}, nil
}

// GetBatchMarketPrices retrieves prices for multiple symbols
func (c *SSIClient) GetBatchMarketPrices(ctx context.Context, symbols []string) (map[string]*client.MarketPrice, error) {
	prices := make(map[string]*client.MarketPrice)

	for _, symbol := range symbols {
		price, err := c.GetMarketPrice(ctx, symbol)
		if err != nil {
			// Log error but continue with other symbols
			continue
		}
		prices[symbol] = price
	}

	return prices, nil
}
