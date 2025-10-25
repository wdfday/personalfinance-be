package handler

import (
	"net/http"

	"personalfinancedss/internal/broker/service"
	"personalfinancedss/internal/middleware"
	accountdto "personalfinancedss/internal/module/cashflow/account/dto"
	accountservice "personalfinancedss/internal/module/cashflow/account/service"
	encryptionService "personalfinancedss/internal/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler manages account endpoints.
type Handler struct {
	service           accountservice.Service
	encryptionService *encryptionService.EncryptionService
	syncService       *service.SyncService
	logger            *zap.Logger
}

// NewHandler constructs an account handler.
func NewHandler(
	service accountservice.Service,
	encryptionService *encryptionService.EncryptionService,
	syncService *service.SyncService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		service:           service,
		encryptionService: encryptionService,
		syncService:       syncService,
		logger:            logger.Named("account.handler"),
	}
}

// RegisterRoutes wires account routes under /api/v1/accounts.
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	accounts := r.Group("/api/v1/accounts")
	accounts.Use(authMiddleware.AuthMiddleware())
	{
		accounts.POST("", h.createAccount)
		accounts.POST("/broker", h.createWithBroker)
		accounts.GET("", h.getMyAccounts)
		accounts.GET("/:id", h.getAccount)
		accounts.PUT("/:id", h.updateAccount)
		accounts.DELETE("/:id", h.deleteAccount)
	}
}

// createAccount godoc
// @Summary Create a new account
// @Description Create a new financial account for the authenticated user
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param account body accountdto.CreateAccountRequest true "Account data"
// @Success 201 {object} accountdto.AccountResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/accounts [post]
func (h *Handler) createAccount(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req accountdto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data")
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), currentUser.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Account created successfully", accountdto.ToResponse(*account))
}

// createWithBroker godoc
// @Summary Create investment account with broker integration
// @Description Create a new investment/crypto account with broker (SSI or OKX) integration. The server will validate credentials by authenticating with broker API, fetch account info, and create account with real data from broker.
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body accountdto.CreateAccountWithBrokerRequest true "Account and broker credentials"
// @Success 201 {object} accountdto.AccountResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/accounts/broker [post]
func (h *Handler) createWithBroker(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		h.logger.Warn("User not found in context")
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	h.logger.Info("Received create account with broker request",
		zap.String("user_id", currentUser.ID.String()),
		zap.String("ip", c.ClientIP()),
	)

	var req accountdto.CreateAccountWithBrokerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body",
			zap.Error(err),
			zap.String("user_id", currentUser.ID.String()),
		)
		shared.RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.Debug("Request payload validated",
		zap.String("broker_type", string(req.BrokerType)),
		zap.String("account_name", req.AccountName),
	)

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed",
			zap.Error(err),
			zap.String("broker_type", string(req.BrokerType)),
		)
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create account with broker integration
	account, err := h.service.CreateAccountWithBroker(c.Request.Context(), currentUser.ID.String(), req)
	if err != nil {
		h.logger.Error("Failed to create account with broker",
			zap.Error(err),
			zap.String("user_id", currentUser.ID.String()),
			zap.String("broker_type", string(req.BrokerType)),
		)
		shared.HandleError(c, err)
		return
	}

	h.logger.Info("Account with broker created successfully",
		zap.String("user_id", currentUser.ID.String()),
		zap.String("account_id", account.ID.String()),
		zap.String("broker_type", string(req.BrokerType)),
	)

	shared.RespondWithSuccess(c, http.StatusCreated, "Account created with broker successfully", accountdto.ToResponse(*account))
}
