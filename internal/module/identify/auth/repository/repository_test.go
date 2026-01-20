package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"personalfinancedss/internal/module/identify/auth/domain"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Create tables manually for SQLite (avoiding PostgreSQL-specific UUID syntax)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS verification_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used_at DATETIME,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS token_blacklist (
			id TEXT PRIMARY KEY,
			token TEXT NOT NULL,
			user_id TEXT NOT NULL,
			reason TEXT,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

// ==================== TokenRepository Tests ====================

func TestTokenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	t.Run("successfully create token", func(t *testing.T) {
		token := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "test_token_" + uuid.New().String(),
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		// Verify token was created
		var result domain.VerificationToken
		err = db.First(&result, "id = ?", token.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, token.Token, result.Token)
	})
}

func TestTokenRepository_GetByToken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	t.Run("successfully get token", func(t *testing.T) {
		tokenStr := "get_token_" + uuid.New().String()
		token := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     tokenStr,
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(token).Error)

		result, err := repo.GetByToken(ctx, tokenStr)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ID, result.ID)
		assert.Equal(t, tokenStr, result.Token)
	})

	t.Run("token not found", func(t *testing.T) {
		result, err := repo.GetByToken(ctx, "non_existent_token")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTokenRepository_MarkAsUsed(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	t.Run("successfully mark token as used", func(t *testing.T) {
		token := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "mark_used_" + uuid.New().String(),
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(token).Error)

		err := repo.MarkAsUsed(ctx, token.ID.String())
		assert.NoError(t, err)

		// Verify token was marked as used
		var result domain.VerificationToken
		require.NoError(t, db.First(&result, "id = ?", token.ID).Error)
		assert.NotNil(t, result.UsedAt)
	})
}

func TestTokenRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	t.Run("delete only expired tokens", func(t *testing.T) {
		// Create expired token
		expiredToken := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "expired_" + uuid.New().String(),
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		}
		require.NoError(t, db.Create(expiredToken).Error)

		// Create valid token
		validToken := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "valid_" + uuid.New().String(),
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(24 * time.Hour), // valid
		}
		require.NoError(t, db.Create(validToken).Error)

		err := repo.DeleteExpired(ctx)
		assert.NoError(t, err)

		// Expired token should be deleted (soft delete)
		var count int64
		db.Model(&domain.VerificationToken{}).Where("id = ?", expiredToken.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		// Valid token should still exist
		db.Model(&domain.VerificationToken{}).Where("id = ?", validToken.ID).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

func TestTokenRepository_DeleteByUserIDAndType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	t.Run("delete tokens by user and type", func(t *testing.T) {
		userID := uuid.New()

		// Create tokens for this user
		token1 := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    userID,
			Token:     "user_token_1_" + uuid.New().String(),
			Type:      string(domain.TokenTypeEmailVerification),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(token1).Error)

		token2 := &domain.VerificationToken{
			ID:        uuid.New(),
			UserID:    userID,
			Token:     "user_token_2_" + uuid.New().String(),
			Type:      string(domain.TokenTypePasswordReset), // different type
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(token2).Error)

		// Delete only email verification tokens
		err := repo.DeleteByUserIDAndType(ctx, userID.String(), string(domain.TokenTypeEmailVerification))
		assert.NoError(t, err)

		// Email verification token should be deleted
		var count int64
		db.Model(&domain.VerificationToken{}).Where("id = ?", token1.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		// Password reset token should still exist
		db.Model(&domain.VerificationToken{}).Where("id = ?", token2.ID).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

// ==================== TokenBlacklistRepository Tests ====================

func TestTokenBlacklistRepository_Add(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("successfully add token to blacklist", func(t *testing.T) {
		token := "blacklist_token_" + uuid.New().String()
		userID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		err := repo.Add(ctx, token, userID, "logout", expiresAt)
		assert.NoError(t, err)

		// Verify token was added
		var result domain.TokenBlacklist
		err = db.First(&result, "token = ?", token).Error
		assert.NoError(t, err)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "logout", result.Reason)
	})
}

func TestTokenBlacklistRepository_IsBlacklisted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("token is blacklisted", func(t *testing.T) {
		token := "is_blacklisted_" + uuid.New().String()
		blacklisted := &domain.TokenBlacklist{
			ID:        uuid.New(),
			Token:     token,
			UserID:    uuid.New(),
			Reason:    "test",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(blacklisted).Error)

		isBlacklisted, err := repo.IsBlacklisted(ctx, token)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
	})

	t.Run("token is not blacklisted", func(t *testing.T) {
		isBlacklisted, err := repo.IsBlacklisted(ctx, "non_existent_token")
		assert.NoError(t, err)
		assert.False(t, isBlacklisted)
	})

	t.Run("expired blacklisted token returns false", func(t *testing.T) {
		token := "expired_blacklist_" + uuid.New().String()
		blacklisted := &domain.TokenBlacklist{
			ID:        uuid.New(),
			Token:     token,
			UserID:    uuid.New(),
			Reason:    "test",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		}
		require.NoError(t, db.Create(blacklisted).Error)

		isBlacklisted, err := repo.IsBlacklisted(ctx, token)
		assert.NoError(t, err)
		assert.False(t, isBlacklisted) // expired should not count
	})
}

func TestTokenBlacklistRepository_CleanupExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("cleanup expired tokens", func(t *testing.T) {
		// Create expired blacklisted token
		expiredToken := &domain.TokenBlacklist{
			ID:        uuid.New(),
			Token:     "cleanup_expired_" + uuid.New().String(),
			UserID:    uuid.New(),
			Reason:    "test",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		require.NoError(t, db.Create(expiredToken).Error)

		// Create valid blacklisted token
		validToken := &domain.TokenBlacklist{
			ID:        uuid.New(),
			Token:     "cleanup_valid_" + uuid.New().String(),
			UserID:    uuid.New(),
			Reason:    "test",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(validToken).Error)

		err := repo.CleanupExpired(ctx)
		assert.NoError(t, err)

		// Expired token should be deleted
		var count int64
		db.Unscoped().Model(&domain.TokenBlacklist{}).Where("id = ?", expiredToken.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		// Valid token should still exist
		db.Model(&domain.TokenBlacklist{}).Where("id = ?", validToken.ID).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

func TestTokenBlacklistRepository_BlacklistAllUserTokens(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("blacklist all user tokens", func(t *testing.T) {
		userID := uuid.New()

		err := repo.BlacklistAllUserTokens(ctx, userID, "password_changed")
		assert.NoError(t, err)

		// Verify marker token was created
		var result domain.TokenBlacklist
		err = db.First(&result, "token = ?", "USER_ALL_TOKENS_"+userID.String()).Error
		assert.NoError(t, err)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "password_changed", result.Reason)
	})
}
