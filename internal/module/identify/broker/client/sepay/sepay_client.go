package sepay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"personalfinancedss/internal/module/identify/broker/client"
	"time"
)

const (
	baseURL        = "https://my.sepay.vn/userapi"
	timeout        = 30 * time.Second
	defaultRefresh = 24 * time.Hour // SePay API Token doesn't expire, use 24h as refresh interval
)

// Client implements the BrokerClient interface for SePay
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new SePay client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Authenticate validates the API token and returns account info
// For SePay, the API token is permanent and doesn't need OAuth flow
func (c *Client) Authenticate(ctx context.Context, credentials client.Credentials) (*client.AuthResponse, error) {
	// Validate API token by getting account info
	_, err := c.getBankAccounts(ctx, credentials.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with SePay: %w", err)
	}

	// SePay API tokens don't expire, but we set a refresh interval
	expiresAt := time.Now().Add(defaultRefresh)

	return &client.AuthResponse{
		AccessToken:  credentials.APIKey,
		RefreshToken: credentials.APIKey, // Same as access token
		ExpiresIn:    int(defaultRefresh.Seconds()),
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken refreshes the access token (for SePay, just validates it)
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*client.AuthResponse, error) {
	// For SePay, refreshing is just re-validating the token
	return c.Authenticate(ctx, client.Credentials{APIKey: refreshToken})
}

// GetPortfolio retrieves account balance and portfolio value
func (c *Client) GetPortfolio(ctx context.Context, accessToken string) (*client.Portfolio, error) {
	accounts, err := c.getBankAccounts(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	totalBalance := 0.0
	for _, acc := range accounts {
		totalBalance += acc.Balance
	}

	return &client.Portfolio{
		TotalValue:  totalBalance,
		CashBalance: totalBalance,
		Currency:    "VND",
		LastUpdated: time.Now(),
	}, nil
}

// GetPositions retrieves current positions (not applicable for SePay banking)
func (c *Client) GetPositions(ctx context.Context, accessToken string) ([]client.Position, error) {
	// SePay is a banking/payment API, not an investment platform
	// Return empty positions
	return []client.Position{}, nil
}

// GetTransactions retrieves transaction history
func (c *Client) GetTransactions(ctx context.Context, accessToken string, startDate, endDate time.Time) ([]client.Transaction, error) {
	// Get all bank accounts first
	accounts, err := c.getBankAccounts(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	var allTransactions []client.Transaction

	// Get transactions for each account
	for _, account := range accounts {
		accountTransactions, err := c.getAccountTransactions(ctx, accessToken, account.ID, startDate, endDate)
		if err != nil {
			// Log error but continue with other accounts
			continue
		}
		allTransactions = append(allTransactions, accountTransactions...)
	}

	return allTransactions, nil
}

// GetMarketPrice is not applicable for SePay (banking API)
func (c *Client) GetMarketPrice(ctx context.Context, symbol string) (*client.MarketPrice, error) {
	return nil, fmt.Errorf("market price not supported by SePay banking API")
}

// GetBatchMarketPrices is not applicable for SePay (banking API)
func (c *Client) GetBatchMarketPrices(ctx context.Context, symbols []string) (map[string]*client.MarketPrice, error) {
	return nil, fmt.Errorf("market prices not supported by SePay banking API")
}

// Internal API methods

type bankAccount struct {
	ID            int     `json:"id"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	BankName      string  `json:"bank_name"`
	BankShortName string  `json:"bank_short_name"`
	Balance       float64 `json:"balance"`
	Status        int     `json:"status"`
}

type transactionResponse struct {
	ID              int     `json:"id"`
	TransactionDate string  `json:"transaction_date"`
	AccountNumber   string  `json:"account_number"`
	TransactionType string  `json:"transaction_type"` // in, out
	Amount          float64 `json:"amount_in"`
	AmountOut       float64 `json:"amount_out"`
	Content         string  `json:"content"`
	Code            string  `json:"code"`
	SubAccount      string  `json:"sub_account"`
	BankBrandName   string  `json:"bank_brand_name"`
	TransferAt      string  `json:"transfer_at"`
	ReferenceNumber string  `json:"reference_number"`
	Balance         float64 `json:"balance"`
}

func (c *Client) getBankAccounts(ctx context.Context, apiToken string) ([]bankAccount, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/bank-accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status  int           `json:"status"`
		Message string        `json:"messages"`
		Data    []bankAccount `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("API returned error: %s", result.Message)
	}

	return result.Data, nil
}

func (c *Client) getAccountTransactions(ctx context.Context, apiToken string, accountID int, startDate, endDate time.Time) ([]client.Transaction, error) {
	// Build request payload
	payload := map[string]interface{}{
		"account_id": accountID,
		"from_date":  startDate.Format("2006-01-02"),
		"to_date":    endDate.Format("2006-01-02"),
		"limit":      1000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/transactions/list", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status   int                   `json:"status"`
		Message  string                `json:"messages"`
		Data     []transactionResponse `json:"transactions"`
		Total    int                   `json:"total"`
		NextPage *int                  `json:"next_page"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("API returned error: %s", result.Message)
	}

	// Convert to client.Transaction
	transactions := make([]client.Transaction, 0, len(result.Data))
	for _, t := range result.Data {
		transactionType := "deposit"
		amount := t.Amount
		if t.TransactionType == "out" || t.AmountOut > 0 {
			transactionType = "withdrawal"
			amount = t.AmountOut
		}

		txDate, _ := time.Parse("2006-01-02 15:04:05", t.TransferAt)

		transactions = append(transactions, client.Transaction{
			ExternalID:      fmt.Sprintf("sepay_%d", t.ID),
			TransactionType: transactionType,
			Symbol:          t.BankBrandName,
			Amount:          amount,
			Currency:        "VND",
			TransactionDate: txDate,
			Status:          "completed",
			Notes:           t.Content,
		})
	}

	return transactions, nil
}
