package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVerificationToken_TableName(t *testing.T) {
	token := VerificationToken{}
	assert.Equal(t, "verification_tokens", token.TableName())
}

func TestVerificationToken_IsExpired(t *testing.T) {
	t.Run("token is expired", func(t *testing.T) {
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, token.IsExpired())
	})

	t.Run("token is not expired", func(t *testing.T) {
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, token.IsExpired())
	})

	t.Run("token expires exactly now", func(t *testing.T) {
		token := &VerificationToken{
			ExpiresAt: time.Now(),
		}
		// Should be expired (or very close)
		assert.True(t, token.IsExpired())
	})
}

func TestVerificationToken_IsUsed(t *testing.T) {
	t.Run("token is used", func(t *testing.T) {
		usedAt := time.Now()
		token := &VerificationToken{
			UsedAt: &usedAt,
		}
		assert.True(t, token.IsUsed())
	})

	t.Run("token is not used", func(t *testing.T) {
		token := &VerificationToken{
			UsedAt: nil,
		}
		assert.False(t, token.IsUsed())
	})
}

func TestVerificationToken_IsValid(t *testing.T) {
	t.Run("valid token - not expired and not used", func(t *testing.T) {
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    nil,
		}
		assert.True(t, token.IsValid())
	})

	t.Run("invalid token - expired", func(t *testing.T) {
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			UsedAt:    nil,
		}
		assert.False(t, token.IsValid())
	})

	t.Run("invalid token - used", func(t *testing.T) {
		usedAt := time.Now()
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    &usedAt,
		}
		assert.False(t, token.IsValid())
	})

	t.Run("invalid token - expired and used", func(t *testing.T) {
		usedAt := time.Now()
		token := &VerificationToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			UsedAt:    &usedAt,
		}
		assert.False(t, token.IsValid())
	})
}

func TestTokenType_Constants(t *testing.T) {
	assert.Equal(t, TokenType("email_verification"), TokenTypeEmailVerification)
	assert.Equal(t, TokenType("password_reset"), TokenTypePasswordReset)
}

func TestVerificationToken_Fields(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()
	ip := "192.168.1.1"
	agent := "test-agent"
	now := time.Now()

	token := &VerificationToken{
		ID:        tokenID,
		UserID:    userID,
		Token:     "test_token_123",
		Type:      string(TokenTypeEmailVerification),
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    nil,
		IPAddress: &ip,
		UserAgent: &agent,
	}

	assert.Equal(t, tokenID, token.ID)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, "test_token_123", token.Token)
	assert.Equal(t, "email_verification", token.Type)
	assert.NotNil(t, token.IPAddress)
	assert.Equal(t, "192.168.1.1", *token.IPAddress)
	assert.NotNil(t, token.UserAgent)
	assert.Equal(t, "test-agent", *token.UserAgent)
}
