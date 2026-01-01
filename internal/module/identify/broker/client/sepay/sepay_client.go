package sepay

import (
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
	_, err := c.GetBankAccounts(ctx, credentials.APIKey)
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
	accounts, err := c.GetBankAccounts(ctx, accessToken)
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
	accounts, err := c.GetBankAccounts(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	var allTransactions []client.Transaction

	// Get transactions for each account
	for _, account := range accounts {
		accountTransactions, err := c.GetAccountTransactions(ctx, accessToken, account.AccountNumber, startDate, endDate)
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
	ID                string `json:"id"`
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	Accumulated       string `json:"accumulated"`
	LastTransaction   string `json:"last_transaction"`
	Label             string `json:"label"`
	Active            string `json:"active"`
	CreatedAt         string `json:"created_at"`
	BankShortName     string `json:"bank_short_name"`
	BankFullName      string `json:"bank_full_name"`
	BankBin           string `json:"bank_bin"`
	BankCode          string `json:"bank_code"`
}

type transactionResponse struct {
	ID                 string `json:"id"`
	TransactionDate    string `json:"transaction_date"`
	AccountNumber      string `json:"account_number"`
	AmountIn           string `json:"amount_in"`
	AmountOut          string `json:"amount_out"`
	Accumulated        string `json:"accumulated"`
	TransactionContent string `json:"transaction_content"`
	ReferenceNumber    string `json:"reference_number"`
	BankBrandName      string `json:"bank_brand_name"`
}

// GetBankAccounts retrieves all bank accounts from SePay (implements BankingBrokerClient)
func (c *Client) GetBankAccounts(ctx context.Context, accessToken string) ([]client.BankAccount, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/bankaccounts/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
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
		Status       int           `json:"status"`
		Error        interface{}   `json:"error"`
		Messages     interface{}   `json:"messages"`
		BankAccounts []bankAccount `json:"bankaccounts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("API returned error status: %d", result.Status)
	}

	// Convert to client.BankAccount
	accounts := make([]client.BankAccount, 0, len(result.BankAccounts))
	for _, acc := range result.BankAccounts {
		var lastTxTime *time.Time
		if acc.LastTransaction != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", acc.LastTransaction); err == nil {
				lastTxTime = &t
			}
		}

		accounts = append(accounts, client.BankAccount{
			AccountNumber:     acc.AccountNumber,
			AccountHolderName: acc.AccountHolderName,
			BankCode:          acc.BankCode,
			BankName:          acc.BankFullName,
			Balance:           parseFloat(acc.Accumulated),
			LastTransaction:   lastTxTime,
			IsActive:          acc.Active == "1",
		})
	}

	return accounts, nil
}

// parseFloat safely converts string to float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// GetAccountTransactions retrieves transactions for a specific account (implements BankingBrokerClient)
func (c *Client) GetAccountTransactions(ctx context.Context, accessToken string, accountNumber string, startDate, endDate time.Time) ([]client.Transaction, error) {
	// Build URL with query parameters
	url := fmt.Sprintf("%s/transactions/list?account_number=%s&transaction_date_min=%s&transaction_date_max=%s&limit=5000",
		baseURL,
		accountNumber,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
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
		Status       int                   `json:"status"`
		Error        interface{}           `json:"error"`
		Messages     interface{}           `json:"messages"`
		Transactions []transactionResponse `json:"transactions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("API returned error status: %d", result.Status)
	}

	// Convert to client.Transaction
	transactions := make([]client.Transaction, 0, len(result.Transactions))
	for _, t := range result.Transactions {
		amountIn := parseFloat(t.AmountIn)
		amountOut := parseFloat(t.AmountOut)

		transactionType := "deposit"
		amount := amountIn
		if amountOut > 0 {
			transactionType = "withdrawal"
			amount = amountOut
		}

		txDate, _ := time.Parse("2006-01-02 15:04:05", t.TransactionDate)

		transactions = append(transactions, client.Transaction{
			ExternalID:      fmt.Sprintf("sepay_%s", t.ID),
			TransactionType: transactionType,
			Symbol:          t.BankBrandName,
			Amount:          amount,
			Currency:        "VND",
			TransactionDate: txDate,
			Status:          "completed",
			Notes:           t.TransactionContent,
			AccountNumber:   t.AccountNumber,
			ReferenceCode:   t.ReferenceNumber,
			RunningBalance:  parseFloat(t.Accumulated),
		})
	}

	return transactions, nil
}
