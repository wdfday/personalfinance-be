package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/identify/auth/dto"
	"personalfinancedss/internal/module/identify/auth/repository"
	"personalfinancedss/internal/module/identify/user/domain"
	"personalfinancedss/internal/module/identify/user/service"
	notificationservice "personalfinancedss/internal/module/notification/service"
	"time"

	"personalfinancedss/internal/config"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles authentication operations
type Service struct {
	userService        service.IUserService
	jwtService         IJWTService
	passwordService    IPasswordService
	googleOAuthService *GoogleOAuthService
	tokenBlacklistRepo repository.ITokenBlacklistRepository
	securityLogger     notificationservice.SecurityLogger
	config             *config.Config
	logger             *zap.Logger
}

// NewService creates a new auth service
func NewService(
	userService service.IUserService,
	jwtService IJWTService,
	passwordService IPasswordService,
	googleOAuthService *GoogleOAuthService,
	tokenBlacklistRepo repository.ITokenBlacklistRepository,
	securityLogger notificationservice.SecurityLogger,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		userService:        userService,
		jwtService:         jwtService,
		passwordService:    passwordService,
		googleOAuthService: googleOAuthService,
		tokenBlacklistRepo: tokenBlacklistRepo,
		securityLogger:     securityLogger,
		config:             cfg,
		logger:             logger,
	}
}

// Register registers a new user
func (s *Service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResult, error) {
	// Check if user already exists
	exists, err := s.userService.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}
	if exists {
		return nil, shared.ErrConflict.WithDetails("field", "email")
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Create user
	user := &domain.User{
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Role:     domain.UserRoleUser,
		Status:   domain.UserStatusPendingVerification,
	}

	if req.Phone != "" {
		user.PhoneNumber = &req.Phone
	}

	createdUser, err := s.userService.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, expiresAt, err := s.jwtService.GenerateAccessToken(
		createdUser.ID.String(),
		createdUser.Email,
		createdUser.Role,
	)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	refreshToken, _, err := s.jwtService.GenerateRefreshToken(createdUser.ID.String())
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Log registration event
	s.securityLogger.LogRegistration(ctx, createdUser.ID.String(), createdUser.Email, "")

	return &dto.AuthResult{
		User:         createdUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login authenticates a user with email and password
func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResult, error) {
	// Get user by email
	user, err := s.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		// Log failed login attempt
		s.securityLogger.LogLoginFailed(ctx, req.Email, req.IP, "user not found")
		return nil, shared.ErrUnauthorized.WithDetails("message", "invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, shared.ErrUnauthorized.WithDetails("message", "account is locked")
	}

	// Verify password
	if err := s.passwordService.VerifyPassword(user.Password, req.Password); err != nil {
		// Increment login attempts
		_ = s.userService.IncLoginAttempts(ctx, user.ID.String())

		// Lock account after 5 failed attempts
		if user.LoginAttempts >= 4 {
			lockUntil := time.Now().Add(15 * time.Minute)
			_ = s.userService.SetLockedUntil(ctx, user.ID.String(), &lockUntil)

			// Log account locked
			s.securityLogger.LogAccountLocked(ctx, user.ID.String(), user.Email, req.IP, lockUntil)
		} else {
			// Log failed password attempt
			s.securityLogger.LogLoginFailed(ctx, req.Email, req.IP, "invalid password")
		}

		return nil, shared.ErrUnauthorized.WithDetails("message", "invalid credentials")
	}

	// Check account status
	if user.Status == domain.UserStatusSuspended {
		return nil, shared.ErrUnauthorized.WithDetails("message", "account is suspended")
	}

	// Reset login attempts
	_ = s.userService.ResetLoginAttempts(ctx, user.ID.String())

	// Update last login
	ip := req.IP
	_ = s.userService.UpdateLastLogin(ctx, user.ID.String(), time.Now(), &ip)

	// Generate tokens
	accessToken, expiresAt, err := s.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	refreshToken, _, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Log successful login
	s.securityLogger.LogLoginSuccess(ctx, user.ID.String(), user.Email, req.IP)

	return &dto.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout logs out a user by blacklisting their refresh token
func (s *Service) Logout(ctx context.Context, userID, refreshToken, ipAddress string) error {
	// Validate refresh token first
	tokenUserID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return shared.ErrUnauthorized.WithDetails("message", "invalid refresh token")
	}

	// Ensure the token belongs to the user
	if tokenUserID != userID {
		return shared.ErrUnauthorized.WithDetails("message", "token does not belong to user")
	}

	// Parse userID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("message", "invalid user ID")
	}

	// Calculate token expiration (7 days from now to match refresh token lifetime)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Add token to blacklist
	if err := s.tokenBlacklistRepo.Add(ctx, refreshToken, userUUID, "logout", expiresAt); err != nil {
		return err
	}

	// Get user for logging
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		// Log with just userID if we can't get user
		s.securityLogger.LogLogout(ctx, userID, "", ipAddress)
		return nil
	}

	// Log successful logout
	s.securityLogger.LogLogout(ctx, user.ID.String(), user.Email, ipAddress)

	return nil
}

// RefreshToken generates a new access token from a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	// Check if token is blacklisted first
	isBlacklisted, err := s.tokenBlacklistRepo.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}
	if isBlacklisted {
		return nil, shared.ErrUnauthorized.WithDetails("message", "token has been revoked")
	}

	// Validate refresh token
	userID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, shared.ErrUnauthorized.WithDetails("message", "invalid refresh token")
	}

	// Get user
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, shared.ErrUnauthorized.WithDetails("message", "user not found")
	}

	// Generate new access token
	accessToken, expiresAt, err := s.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return dto.NewTokenResponse(accessToken, expiresAt), nil
}

// AuthenticateGoogle authenticates a user with Google OAuth
func (s *Service) AuthenticateGoogle(ctx context.Context, req dto.GoogleAuthRequest) (*dto.AuthResult, error) {
	// Verify Google token and get user info
	googleUser, err := s.googleOAuthService.VerifyGoogleToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	// Try to find existing user by email
	user, err := s.userService.GetByEmail(ctx, googleUser.Email)
	if err != nil {
		// User doesn't exist, create new one
		if err == shared.ErrUserNotFound {
			newUser := &domain.User{
				Email:         googleUser.Email,
				FullName:      googleUser.Name,
				Role:          domain.UserRoleUser,
				Status:        domain.UserStatusActive,
				EmailVerified: googleUser.VerifiedEmail,
			}

			if googleUser.Picture != "" {
				newUser.AvatarURL = &googleUser.Picture
			}

			// Set a random password (won't be used for Google OAuth users)
			randomPassword, err := s.passwordService.HashPassword(fmt.Sprintf("google_oauth_%s", googleUser.ID))
			if err != nil {
				return nil, shared.ErrInternal.WithError(err)
			}
			newUser.Password = randomPassword

			// Create user
			user, err = s.userService.Create(ctx, newUser)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, err
		}
	}

	// If user exists but email not verified, mark as verified (Google confirmed it)
	if !user.EmailVerified && googleUser.VerifiedEmail {
		if err := s.userService.MarkEmailVerified(ctx, user.ID.String(), time.Now()); err != nil {
			// Log error but don't fail authentication
			s.logger.Warn(
				"Failed to mark email as verified",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email),
				zap.Error(err),
			)
		}
		user.EmailVerified = true
	}

	// Generate tokens
	accessToken, expiresAt, err := s.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	refreshToken, _, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// Update last login
	ip := ""
	_ = s.userService.UpdateLastLogin(ctx, user.ID.String(), time.Now(), &ip)

	// Determine if this is a new user for logging
	isNewUser := user.CreatedAt.After(time.Now().Add(-5 * time.Second))

	// Log Google OAuth login
	s.securityLogger.LogGoogleOAuthLogin(ctx, user.ID.String(), user.Email, ip, isNewUser)

	return &dto.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
