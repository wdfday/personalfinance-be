package service

import (
	"context"

	"personalfinancedss/internal/module/identify/auth/domain"
	"personalfinancedss/internal/module/identify/auth/dto"
	"personalfinancedss/internal/module/identify/auth/helper"
	"personalfinancedss/internal/module/identify/auth/repository"
	userservice "personalfinancedss/internal/module/identify/user/service"
	notificationservice "personalfinancedss/internal/module/notification/service"
	"personalfinancedss/internal/shared"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing, validation, and management operations
type PasswordService struct {
	cost           int // bcrypt cost factor
	userService    userservice.IUserService
	tokenRepo      repository.TokenRepository
	tokenGenerator ITokenService
	emailService   notificationservice.EmailService
	logger         *zap.Logger
}

// NewPasswordService creates a new password service
func NewPasswordService(
	userService userservice.IUserService,
	tokenRepo repository.TokenRepository,
	tokenGenerator ITokenService,
	emailService notificationservice.EmailService,
	logger *zap.Logger,
) *PasswordService {
	return &PasswordService{
		cost:           bcrypt.DefaultCost, // cost 10
		userService:    userService,
		tokenRepo:      tokenRepo,
		tokenGenerator: tokenGenerator,
		emailService:   emailService,
		logger:         logger,
	}
}

// HashPassword hashes a plaintext password using bcrypt
func (s *PasswordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a plaintext password against a bcrypt hash
func (s *PasswordService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *PasswordService) IsValidPassword(password string) bool {
	return helper.IsPasswordStrong(password)
}

func (s *PasswordService) ValidatePasswordStrength(password string) []string {
	return helper.PasswordValidationErrors(password)
}

// ChangePassword allows authenticated users to update their password
func (s *PasswordService) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.VerifyPassword(user.Password, req.CurrentPassword); err != nil {
		return shared.ErrUnauthorized.WithDetails("message", "current password is incorrect")
	}

	if !s.IsValidPassword(req.NewPassword) {
		return shared.ErrBadRequest.WithDetails("message", "password does not meet security requirements")
	}

	hashedPassword, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	return s.userService.UpdatePassword(ctx, userID, hashedPassword)
}

// ForgotPassword initiates the password reset flow by issuing a reset token
func (s *PasswordService) ForgotPassword(ctx context.Context, email, ipAddress, userAgent string) error {
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		// Do not leak if the user exists
		return nil
	}

	if err := s.tokenRepo.DeleteByUserIDAndType(ctx, user.ID.String(), string(domain.TokenTypePasswordReset)); err != nil {
		return shared.ErrInternal.WithError(err)
	}

	tokenStr, err := s.tokenGenerator.GenerateTokenWithPrefix("pwd_reset")
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	token := &domain.VerificationToken{
		UserID:    user.ID,
		Token:     tokenStr,
		Type:      string(domain.TokenTypePasswordReset),
		ExpiresAt: s.tokenGenerator.GetPasswordResetExpiry(),
		IPAddress: helper.StringPtr(ipAddress),
		UserAgent: helper.StringPtr(userAgent),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return shared.ErrInternal.WithError(err)
	}

	if s.emailService != nil {
		if err := s.emailService.SendPasswordResetEmail(user.Email, user.FullName, tokenStr); err != nil {
			s.logger.Error(
				"Failed to send password reset email",
				zap.String("email", user.Email),
				zap.Error(err),
			)
		}
	}

	return nil
}

// ResetPassword validates a reset token and updates the user's password
func (s *PasswordService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if err == shared.ErrTokenNotFound {
			return shared.ErrTokenNotFound.WithDetails("type", "password reset")
		}
		return shared.ErrInternal.WithError(err)
	}

	if token.Type != string(domain.TokenTypePasswordReset) {
		return shared.ErrTokenInvalid.WithDetails("message", "wrong token type")
	}

	if token.IsUsed() {
		return shared.ErrTokenUsed
	}

	if token.IsExpired() {
		return shared.ErrTokenExpired
	}

	if !s.IsValidPassword(newPassword) {
		return shared.ErrBadRequest.WithDetails("message", "password does not meet security requirements")
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	if err := s.userService.UpdatePassword(ctx, token.UserID.String(), hashedPassword); err != nil {
		return err
	}

	if err := s.tokenRepo.MarkAsUsed(ctx, token.ID.String()); err != nil {
		return shared.ErrInternal.WithError(err)
	}

	_ = s.userService.ResetLoginAttempts(ctx, token.UserID.String())

	return nil
}
