package repository

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/identify/auth/domain"
	"time"

	"personalfinancedss/internal/shared"

	"gorm.io/gorm"
)

// TokenRepository handles token persistence
type TokenRepository interface {
	Create(ctx context.Context, token *domain.VerificationToken) error
	GetByToken(ctx context.Context, tokenStr string) (*domain.VerificationToken, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserIDAndType(ctx context.Context, userID, tokenType string) error
}

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

// Create creates a new token
func (r *tokenRepository) Create(ctx context.Context, token *domain.VerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByToken retrieves a token by its string value
func (r *tokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.VerificationToken, error) {
	var token domain.VerificationToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND deleted_at IS NULL", tokenStr).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrTokenNotFound
		}
		return nil, err
	}

	return &token, nil
}

// MarkAsUsed marks a token as used
func (r *tokenRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.VerificationToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error
}

// DeleteExpired deletes all expired tokens
func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.VerificationToken{}).Error
}

// DeleteByUserIDAndType deletes all tokens for a user by type
func (r *tokenRepository) DeleteByUserIDAndType(ctx context.Context, userID, tokenType string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, tokenType).
		Delete(&domain.VerificationToken{}).Error
}
