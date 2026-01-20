package sepay

import (
	"context"
	"personalfinancedss/internal/module/identify/broker/client"
	"testing"
	"time"
)

const testAPIToken = "QVRQGYKNE8QVGZMHFAR3KVU9BSTJPZS6JTCLOW3Z76BFQXN4OOTFKRXTRUCXD2WE"

func TestGetBankAccounts(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	accounts, err := c.GetBankAccounts(ctx, testAPIToken)
	if err != nil {
		t.Fatalf("Failed to get bank accounts: %v", err)
	}

	t.Logf("Found %d bank accounts", len(accounts))
	for _, acc := range accounts {
		t.Logf("  Account: %s - %s (%s) Balance: %.2f VND, Active: %v",
			acc.AccountNumber,
			acc.AccountHolderName,
			acc.BankName,
			acc.Balance,
			acc.IsActive,
		)
	}

	if len(accounts) == 0 {
		t.Error("Expected at least one bank account")
	}
}

func TestGetAccountTransactions(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	// First get accounts
	accounts, err := c.GetBankAccounts(ctx, testAPIToken)
	if err != nil {
		t.Fatalf("Failed to get bank accounts: %v", err)
	}

	if len(accounts) == 0 {
		t.Skip("No bank accounts found")
	}

	// Get transactions for first active account
	var accountNumber string
	for _, acc := range accounts {
		if acc.IsActive {
			accountNumber = acc.AccountNumber
			break
		}
	}

	if accountNumber == "" {
		t.Skip("No active bank accounts found")
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // Last month

	transactions, err := c.GetAccountTransactions(ctx, testAPIToken, accountNumber, startDate, endDate)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	t.Logf("Found %d transactions for account %s", len(transactions), accountNumber)
	for _, txn := range transactions {
		t.Logf("  [%s] %s: %.2f VND - %s",
			txn.TransactionDate.Format("2006-01-02"),
			txn.TransactionType,
			txn.Amount,
			txn.Notes,
		)
	}
}

func TestAuthenticate(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	authResp, err := c.Authenticate(ctx, client.Credentials{APIKey: testAPIToken})
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	t.Logf("Auth successful:")
	t.Logf("  AccessToken: %s...", authResp.AccessToken[:20])
	t.Logf("  ExpiresAt: %v", authResp.ExpiresAt)
	t.Logf("  TokenType: %s", authResp.TokenType)
}

func TestGetPortfolio(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	portfolio, err := c.GetPortfolio(ctx, testAPIToken)
	if err != nil {
		t.Fatalf("Failed to get portfolio: %v", err)
	}

	t.Logf("Portfolio:")
	t.Logf("  TotalValue: %.2f %s", portfolio.TotalValue, portfolio.Currency)
	t.Logf("  CashBalance: %.2f", portfolio.CashBalance)
	t.Logf("  LastUpdated: %v", portfolio.LastUpdated)
}
