package okx

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"personalfinancedss/internal/broker/client"
)

const (
	OKXBaseURL = "https://www.okx.com"
)

// OKXClient implements the BrokerClient interface for OKX Exchange
type OKXClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewOKXClient creates a new OKX client
func NewOKXClient() *OKXClient {
	return &OKXClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: OKXBaseURL,
	}
}

// sign creates HMAC SHA256 signature for OKX API
func (c *OKXClient) sign(timestamp, method, requestPath, body, secretKey string) string {
	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// makeRequest makes an authenticated request to OKX API
func (c *OKXClient) makeRequest(ctx context.Context, method, path string, body interface{}, apiKey, apiSecret, passphrase string) (*http.Response, error) {
	var bodyBytes []byte
	var err error
	bodyStr := ""

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyStr = string(bodyBytes)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	signature := c.sign(timestamp, method, path, bodyStr, apiSecret)

	url := c.baseURL + path
	var req *http.Request
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OK-ACCESS-KEY", apiKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", passphrase)

	return c.httpClient.Do(req)
}

// Authenticate - OKX uses API Key authentication, no separate auth endpoint
func (c *OKXClient) Authenticate(ctx context.Context, credentials client.Credentials) (*client.AuthResponse, error) {
	if credentials.Passphrase == nil || *credentials.Passphrase == "" {
		return nil, fmt.Errorf("passphrase is required for OKX")
	}

	// OKX doesn't have a separate auth endpoint, API key is used directly
	// We'll validate the credentials by making a test request
	resp, err := c.makeRequest(ctx, "GET", "/api/v5/account/balance", nil,
		credentials.APIKey, credentials.APISecret, *credentials.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to validate credentials: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: %s", string(body))
	}

	// For OKX, we don't get tokens, we just use the API key
	// Return a dummy auth response
	expiresAt := time.Now().Add(365 * 24 * time.Hour) // API keys don't expire
	return &client.AuthResponse{
		AccessToken:  credentials.APIKey,
		RefreshToken: "",
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshToken - OKX doesn't need token refresh, API keys are permanent
func (c *OKXClient) RefreshToken(ctx context.Context, refreshToken string) (*client.AuthResponse, error) {
	return nil, fmt.Errorf("OKX uses API keys and doesn't support token refresh")
}

// GetPortfolio retrieves portfolio information from OKX
func (c *OKXClient) GetPortfolio(ctx context.Context, accessToken string) (*client.Portfolio, error) {
	// accessToken is actually the API key for OKX
	// We need to extract credentials from somewhere or pass them differently
	// For now, return an error indicating we need the full credentials
	return nil, fmt.Errorf("OKX GetPortfolio requires full credentials, not just access token")
}

// GetPortfolioWithCredentials retrieves portfolio using full credentials
func (c *OKXClient) GetPortfolioWithCredentials(ctx context.Context, apiKey, apiSecret, passphrase string) (*client.Portfolio, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v5/account/balance", nil, apiKey, apiSecret, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("portfolio request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var balanceResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			TotalEq string `json:"totalEq"` // Total equity in USD
			IsoEq   string `json:"isoEq"`   // Isolated margin equity in USD
			AdjEq   string `json:"adjEq"`   // Adjusted equity in USD
			Details []struct {
				Ccy      string `json:"ccy"`      // Currency
				Eq       string `json:"eq"`       // Equity
				AvailBal string `json:"availBal"` // Available balance
				UplNow   string `json:"uplNow"`   // Unrealized P&L
			} `json:"details"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return nil, fmt.Errorf("failed to decode balance response: %w", err)
	}

	if balanceResp.Code != "0" {
		return nil, fmt.Errorf("OKX API error: %s", balanceResp.Msg)
	}

	if len(balanceResp.Data) == 0 {
		return &client.Portfolio{
			Currency:    "USD",
			LastUpdated: time.Now(),
		}, nil
	}

	data := balanceResp.Data[0]
	totalValue := parseFloat(data.TotalEq)

	// Calculate asset allocation
	assetAllocation := make(map[string]float64)
	cashBalance := 0.0
	unrealizedGain := 0.0

	for _, detail := range data.Details {
		value := parseFloat(detail.Eq)
		assetAllocation[detail.Ccy] = value

		if detail.Ccy == "USDT" || detail.Ccy == "USD" || detail.Ccy == "USDC" {
			cashBalance += value
		}

		unrealizedGain += parseFloat(detail.UplNow)
	}

	return &client.Portfolio{
		TotalValue:      totalValue,
		CashBalance:     cashBalance,
		UnrealizedGain:  unrealizedGain,
		Currency:        "USD",
		AssetAllocation: assetAllocation,
		LastUpdated:     time.Now(),
	}, nil
}

// GetPositions retrieves current positions from OKX
func (c *OKXClient) GetPositions(ctx context.Context, accessToken string) ([]client.Position, error) {
	return nil, fmt.Errorf("OKX GetPositions requires full credentials, not just access token")
}

// GetPositionsWithCredentials retrieves positions using full credentials
func (c *OKXClient) GetPositionsWithCredentials(ctx context.Context, apiKey, apiSecret, passphrase string) ([]client.Position, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v5/account/positions", nil, apiKey, apiSecret, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("positions request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var positionsResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId      string `json:"instId"`      // Instrument ID (e.g., BTC-USDT)
			Pos         string `json:"pos"`         // Position quantity
			AvgPx       string `json:"avgPx"`       // Average open price
			MarkPx      string `json:"markPx"`      // Mark price
			UplNow      string `json:"uplNow"`      // Unrealized P&L
			UplRatio    string `json:"uplRatio"`    // Unrealized P&L ratio
			NotionalUsd string `json:"notionalUsd"` // Position value in USD
			Lever       string `json:"lever"`       // Leverage
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&positionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode positions response: %w", err)
	}

	if positionsResp.Code != "0" {
		return nil, fmt.Errorf("OKX API error: %s", positionsResp.Msg)
	}

	positions := make([]client.Position, 0, len(positionsResp.Data))
	for _, pos := range positionsResp.Data {
		if parseFloat(pos.Pos) == 0 {
			continue // Skip zero positions
		}

		// Parse symbol (e.g., BTC-USDT -> BTC)
		symbol := strings.Split(pos.InstId, "-")[0]

		quantity := parseFloat(pos.Pos)
		avgPrice := parseFloat(pos.AvgPx)
		currentPrice := parseFloat(pos.MarkPx)
		unrealizedPL := parseFloat(pos.UplNow)
		unrealizedPLPct := parseFloat(pos.UplRatio) * 100

		positions = append(positions, client.Position{
			Symbol:             symbol,
			Name:               pos.InstId,
			AssetType:          "crypto",
			Quantity:           quantity,
			AverageCostPerUnit: avgPrice,
			CurrentPrice:       currentPrice,
			CurrentValue:       parseFloat(pos.NotionalUsd),
			UnrealizedGain:     unrealizedPL,
			UnrealizedGainPct:  unrealizedPLPct,
			Currency:           "USD",
			Exchange:           "OKX",
			ExternalID:         pos.InstId,
			LastUpdated:        time.Now(),
		})
	}

	return positions, nil
}

// GetTransactions retrieves transaction history from OKX
func (c *OKXClient) GetTransactions(ctx context.Context, accessToken string, startDate, endDate time.Time) ([]client.Transaction, error) {
	return nil, fmt.Errorf("OKX GetTransactions requires full credentials, not just access token")
}

// GetMarketPrice retrieves current market price for a symbol
func (c *OKXClient) GetMarketPrice(ctx context.Context, symbol string) (*client.MarketPrice, error) {
	// Format: BTC-USDT for spot
	instId := symbol + "-USDT"
	url := fmt.Sprintf("%s/api/v5/market/ticker?instId=%s", c.baseURL, instId)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create market price request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get market price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market price request failed with status %d", resp.StatusCode)
	}

	var priceResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId  string `json:"instId"`
			Last    string `json:"last"`    // Latest price
			LastSz  string `json:"lastSz"`  // Latest size
			AskPx   string `json:"askPx"`   // Ask price
			BidPx   string `json:"bidPx"`   // Bid price
			Open24h string `json:"open24h"` // 24h open
			High24h string `json:"high24h"` // 24h high
			Low24h  string `json:"low24h"`  // 24h low
			Vol24h  string `json:"vol24h"`  // 24h volume
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return nil, fmt.Errorf("failed to decode market price response: %w", err)
	}

	if priceResp.Code != "0" || len(priceResp.Data) == 0 {
		return nil, fmt.Errorf("OKX API error: %s", priceResp.Msg)
	}

	data := priceResp.Data[0]
	price := parseFloat(data.Last)
	open := parseFloat(data.Open24h)
	change := price - open
	changePct := 0.0
	if open > 0 {
		changePct = (change / open) * 100
	}

	return &client.MarketPrice{
		Symbol:      symbol,
		Price:       price,
		Change:      change,
		ChangePct:   changePct,
		Volume:      parseFloat(data.Vol24h),
		Currency:    "USD",
		LastUpdated: time.Now(),
	}, nil
}

// GetBatchMarketPrices retrieves prices for multiple symbols
func (c *OKXClient) GetBatchMarketPrices(ctx context.Context, symbols []string) (map[string]*client.MarketPrice, error) {
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

// parseFloat converts string to float64, returns 0 on error
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
