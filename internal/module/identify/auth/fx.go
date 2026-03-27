package auth

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/identify/auth/handler"
	"personalfinancedss/internal/module/identify/auth/repository"
	"personalfinancedss/internal/module/identify/auth/service"
	userservice "personalfinancedss/internal/module/identify/user/service"
	notificationservice "personalfinancedss/internal/module/notification/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ProvideJWTService creates a JWT service with configuration
func ProvideJWTService(cfg *config.Config) service.IJWTService {
	return service.NewJWTService(cfg.Auth.JWTSecret)
}

// ProvidePasswordService creates a password service
func ProvidePasswordService(
	userService userservice.IUserService,
	tokenRepo repository.TokenRepository,
	tokenService service.ITokenService,
	emailService notificationservice.EmailService,
	logger *zap.Logger,
) service.IPasswordService {
	return service.NewPasswordService(userService, tokenRepo, tokenService, emailService, logger)
}

// ProvideTokenService creates a token service
func ProvideTokenService() service.ITokenService {
	return service.NewTokenService()
}

// Module provides auth module dependencies
var Module = fx.Module("auth",
	fx.Provide(
		// Core services
		ProvideJWTService,
		ProvidePasswordService,
		ProvideTokenService,
		service.NewGoogleOAuthService,

		// Repositories
		repository.NewTokenRepository,
		repository.NewTokenBlacklistRepository,

		// Verification Service - provide as interface
		fx.Annotate(
			service.NewVerificationService,
			fx.As(new(service.IVerificationService)),
		),

		// Auth Service - provide as interface
		fx.Annotate(
			service.NewService,
			fx.As(new(service.IAuthService)),
		),

		// Handler
		handler.NewHandler,
	),
	fx.Invoke(registerAuthRoutes),
)

func registerAuthRoutes(
	router *gin.Engine,
	h *handler.Handler,
	authMiddleware *middleware.Middleware,
	emailVerificationMiddleware *middleware.EmailVerificationMiddleware,
) {
	h.RegisterRoutes(router, authMiddleware, emailVerificationMiddleware)
}
