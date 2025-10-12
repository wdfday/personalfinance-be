package auth

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/identify/auth/handler"
	repository2 "personalfinancedss/internal/module/identify/auth/repository"
	service2 "personalfinancedss/internal/module/identify/auth/service"
	userservice "personalfinancedss/internal/module/identify/user/service"
	notificationservice "personalfinancedss/internal/module/notification/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ProvideJWTService creates a JWT service with configuration
func ProvideJWTService(cfg *config.Config) service2.IJWTService {
	return service2.NewJWTService(cfg.Auth.JWTSecret)
}

// ProvidePasswordService creates a password service
func ProvidePasswordService(
	userService userservice.IUserService,
	tokenRepo repository2.TokenRepository,
	tokenService service2.ITokenService,
	emailService notificationservice.EmailService,
	logger *zap.Logger,
) service2.IPasswordService {
	return service2.NewPasswordService(userService, tokenRepo, tokenService, emailService, logger)
}

// ProvideTokenService creates a token service
func ProvideTokenService() service2.ITokenService {
	return service2.NewTokenService()
}

// Module provides auth module dependencies
var Module = fx.Module("auth",
	fx.Provide(
		// Core services
		ProvideJWTService,
		ProvidePasswordService,
		ProvideTokenService,
		service2.NewGoogleOAuthService,

		// Repositories
		repository2.NewTokenRepository,
		repository2.NewTokenBlacklistRepository,

		// Verification Service - provide as interface
		fx.Annotate(
			service2.NewVerificationService,
			fx.As(new(service2.IVerificationService)),
		),

		// Auth Service - provide as interface
		fx.Annotate(
			service2.NewService,
			fx.As(new(service2.IAuthService)),
		),

		// Handler
		handler.NewHandler,
	),
)
