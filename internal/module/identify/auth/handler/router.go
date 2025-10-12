package handler

import (
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/identify/auth/service"

	"github.com/gin-gonic/gin"
)

// Handler is the facade that wires individual auth sub-handlers into the router.
type Handler struct {
	auth     *AuthHandler
	password *PasswordHandler
	verify   *VerifyHandler
}

// NewHandler builds the facade handler with its specialized sub-handlers.
func NewHandler(
	authService service.IAuthService,
	passwordService service.IPasswordService,
	verificationService service.IVerificationService,
	cfg *config.Config,
) *Handler {
	return &Handler{
		auth:     NewAuthHandler(authService, cfg),
		password: NewPasswordHandler(passwordService),
		verify:   NewVerifyHandler(verificationService),
	}
}

// RegisterRoutes exposes every auth-related route group.
func (h *Handler) RegisterRoutes(
	r *gin.Engine,
	authMiddleware *middleware.Middleware,
	emailVerificationMiddleware *middleware.EmailVerificationMiddleware,
) {
	h.auth.RegisterRoutes(r, authMiddleware)
	h.password.RegisterRoutes(r, authMiddleware, emailVerificationMiddleware)
	h.verify.RegisterRoutes(r, authMiddleware)
}
