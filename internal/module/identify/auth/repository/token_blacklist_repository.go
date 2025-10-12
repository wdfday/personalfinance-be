package repository

import (
	"context"
	"personalfinancedss/internal/module/identify/auth/domain"
	"time"

	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ITokenBlacklistRepository defines operations for token blacklist management
type ITokenBlacklistRepository interface {
	// Add adds a token to the blacklist
	Add(ctx context.Context, token string, userID uuid.UUID, reason string, expiresAt time.Time) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// CleanupExpired removes expired tokens from blacklist
	CleanupExpired(ctx context.Context) error

	// BlacklistAllUserTokens blacklists all tokens for a user (e.g., on password change)
	BlacklistAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error
}

// tokenBlacklistRepository implements ITokenBlacklistRepository
type tokenBlacklistRepository struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db *gorm.DB) ITokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

// Add adds a token to the blacklist
func (r *tokenBlacklistRepository) Add(ctx context.Context, token string, userID uuid.UUID, reason string, expiresAt time.Time) error {
	blacklistedToken := &domain.TokenBlacklist{
		Token:     token,
		UserID:    userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}

	if err := r.db.WithContext(ctx).Create(blacklistedToken).Error; err != nil {
		return shared.ErrInternal.WithError(err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.TokenBlacklist{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, shared.ErrInternal.WithError(err)
	}

	return count > 0, nil
}

// CleanupExpired removes expired tokens from blacklist
func (r *tokenBlacklistRepository) CleanupExpired(ctx context.Context) error {
	// Permanently delete expired tokens (past their expiration time)
	err := r.db.WithContext(ctx).
		Unscoped(). // bypass soft delete
		Where("expires_at < ?", time.Now()).
		Delete(&domain.TokenBlacklist{}).Error

	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	return nil
}

// BlacklistAllUserTokens blacklists all tokens for a user
// This is useful when user changes password or account is compromised
func (r *tokenBlacklistRepository) BlacklistAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error {
	// Since we don't store all active tokens, this is a marker operation
	// The middleware will need to check token creation time vs this timestamp
	// For now, we'll just add a marker entry
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // Match refresh token expiry

	markerToken := &domain.TokenBlacklist{
		Token:     "USER_ALL_TOKENS_" + userID.String(),
		UserID:    userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}

	if err := r.db.WithContext(ctx).Create(markerToken).Error; err != nil {
		return shared.ErrInternal.WithError(err)
	}

	return nil
}
