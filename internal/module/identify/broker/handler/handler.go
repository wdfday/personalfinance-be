package handler

import (
	"net/http"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/dto"
	"personalfinancedss/internal/module/identify/broker/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BrokerConnectionHandler handles HTTP requests for broker connections
type BrokerConnectionHandler struct {
	service service.BrokerConnectionService
	logger  *zap.Logger
}

// NewBrokerConnectionHandler creates a new broker connection handler
func NewBrokerConnectionHandler(service service.BrokerConnectionService, logger *zap.Logger) *BrokerConnectionHandler {
	return &BrokerConnectionHandler{
		service: service,
		logger:  logger.Named("broker.handler"),
	}
}

// RegisterRoutes registers all broker connection routes
func (h *BrokerConnectionHandler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.Middleware) {
	brokerGroup := router.Group("/api/v1/broker-connections")
	{
		// Public endpoint - no auth required
		brokerGroup.GET("/providers", h.ListProviders)
	}

	// Protected endpoints - require authentication
	protected := router.Group("/api/v1/broker-connections")
	protected.Use(authMiddleware.AuthMiddleware())
	{
		// Broker-specific create endpoints (recommended)
		protected.POST("/ssi", h.CreateSSI)
		protected.POST("/okx", h.CreateOKX)
		protected.POST("/sepay", h.CreateSepay)

		// Connection management
		protected.GET("", h.List)
		protected.GET("/:id", h.GetByID)
		protected.PUT("/:id", h.Update)
		protected.DELETE("/:id", h.Delete)
		protected.POST("/:id/activate", h.Activate)
		protected.POST("/:id/deactivate", h.Deactivate)
		protected.POST("/:id/refresh-token", h.RefreshToken)
		protected.POST("/:id/test", h.TestConnection)
		protected.POST("/:id/sync", h.SyncNow)
	}
}

// ListProviders lists available broker providers
// @Summary List available broker providers
// @Description Get information about all supported broker providers
// @Tags Broker Connections
// @Produce json
// @Success 200 {object} dto.ListBrokerProvidersResponse
// @Router /api/v1/broker-connections/providers [get]
func (h *BrokerConnectionHandler) ListProviders(c *gin.Context) {
	response := dto.GetBrokerProviders()
	shared.RespondWithSuccess(c, http.StatusOK, "Broker providers retrieved successfully", response)
}

// CreateSSI creates a new SSI broker connection
// @Summary Create SSI broker connection
// @Description Create and validate a new SSI Securities broker connection
// @Tags Broker Connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSSIConnectionRequest true "SSI connection details"
// @Success 201 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/ssi [post]
func (h *BrokerConnectionHandler) CreateSSI(c *gin.Context) {
	h.logger.Info("📥 Received SSI broker connection request")

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.logger.Error("❌ Failed to get user ID from context", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ User authenticated", zap.String("user_id", userID.String()))

	var req dto.CreateSSIConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("❌ Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("📝 Parsed SSI connection request",
		zap.String("broker_name", req.BrokerName),
		zap.Bool("auto_sync", req.AutoSync != nil && *req.AutoSync),
	)

	h.logger.Info("🔄 Creating SSI broker connection...")
	connection, err := h.service.Create(c.Request.Context(), req.ToServiceRequest(userID))
	if err != nil {
		h.logger.Error("❌ Failed to create SSI connection",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ SSI broker connection created successfully",
		zap.String("connection_id", connection.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("broker_name", connection.BrokerName),
		zap.String("status", string(connection.Status)),
	)

	shared.RespondWithSuccess(c, http.StatusCreated, "SSI broker connection created successfully", dto.ToBrokerConnectionResponse(connection))
}

// CreateOKX creates a new OKX broker connection
// @Summary Create OKX broker connection
// @Description Create and validate a new OKX exchange connection
// @Tags Broker Connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOKXConnectionRequest true "OKX connection details"
// @Success 201 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/okx [post]
func (h *BrokerConnectionHandler) CreateOKX(c *gin.Context) {
	h.logger.Info("📥 Received OKX broker connection request")

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.logger.Error("❌ Failed to get user ID from context", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ User authenticated", zap.String("user_id", userID.String()))

	var req dto.CreateOKXConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("❌ Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("📝 Parsed OKX connection request",
		zap.String("broker_name", req.BrokerName),
		zap.Bool("auto_sync", req.AutoSync != nil && *req.AutoSync),
	)

	h.logger.Info("🔄 Creating OKX broker connection...")
	connection, err := h.service.Create(c.Request.Context(), req.ToServiceRequest(userID))
	if err != nil {
		h.logger.Error("❌ Failed to create OKX connection",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ OKX broker connection created successfully",
		zap.String("connection_id", connection.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("broker_name", connection.BrokerName),
		zap.String("status", string(connection.Status)),
	)

	shared.RespondWithSuccess(c, http.StatusCreated, "OKX broker connection created successfully", dto.ToBrokerConnectionResponse(connection))
}

// CreateSepay creates a new SePay broker connection
// @Summary Create SePay broker connection
// @Description Create and validate a new SePay payment integration
// @Tags Broker Connections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSepayConnectionRequest true "SePay connection details"
// @Success 201 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/sepay [post]
func (h *BrokerConnectionHandler) CreateSepay(c *gin.Context) {
	h.logger.Info("📥 Received SePay broker connection request")

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.logger.Error("❌ Failed to get user ID from context", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ User authenticated", zap.String("user_id", userID.String()))

	var req dto.CreateSepayConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("❌ Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("📝 Parsed SePay connection request",
		zap.String("broker_name", req.BrokerName),
		zap.Bool("auto_sync", req.AutoSync != nil && *req.AutoSync),
	)

	h.logger.Info("🔄 Creating SePay broker connection...")
	connection, err := h.service.Create(c.Request.Context(), req.ToServiceRequest(userID))
	if err != nil {
		h.logger.Error("❌ Failed to create SePay connection",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("✅ SePay broker connection created successfully",
		zap.String("connection_id", connection.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("broker_name", connection.BrokerName),
		zap.String("status", string(connection.Status)),
	)

	shared.RespondWithSuccess(c, http.StatusCreated, "SePay broker connection created successfully", dto.ToBrokerConnectionResponse(connection))
}

// List retrieves all broker connections for the authenticated user
// @Summary List broker connections
// @Description Get all broker connections for the authenticated user
// @Tags Broker Connections
// @Produce json
// @Param broker_type query string false "Filter by broker type"
// @Param status query string false "Filter by status"
// @Param auto_sync_only query bool false "Only return auto-sync enabled connections"
// @Param active_only query bool false "Only return active connections"
// @Param needing_sync_only query bool false "Only return connections needing sync"
// @Success 200 {object} dto.BrokerConnectionListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections [get]
func (h *BrokerConnectionHandler) List(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var query dto.ListBrokerConnectionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert query to service filters
	filters := &service.ListFilters{
		AutoSyncOnly:    query.AutoSyncOnly,
		ActiveOnly:      query.ActiveOnly,
		NeedingSyncOnly: query.NeedingSyncOnly,
	}
	if query.BrokerType != nil {
		brokerType := domain.BrokerType(*query.BrokerType)
		filters.BrokerType = &brokerType
	}
	if query.Status != nil {
		status := domain.BrokerConnectionStatus(*query.Status)
		filters.Status = &status
	}

	connections, err := h.service.List(c.Request.Context(), userID, filters)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Broker connections retrieved successfully", dto.ToBrokerConnectionListResponse(connections))
}

// GetByID retrieves a broker connection by ID
// @Summary Get broker connection by ID
// @Description Get a specific broker connection by ID
// @Tags Broker Connections
// @Produce json
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id} [get]
func (h *BrokerConnectionHandler) GetByID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	connection, err := h.service.GetByID(c.Request.Context(), connectionID, userID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Broker connection retrieved successfully", dto.ToBrokerConnectionResponse(connection))
}

// Update updates a broker connection
// @Summary Update broker connection
// @Description Update an existing broker connection
// @Tags Broker Connections
// @Accept json
// @Produce json
// @Param id path string true "Broker Connection ID"
// @Param request body dto.UpdateBrokerConnectionRequest true "Update details"
// @Success 200 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id} [put]
func (h *BrokerConnectionHandler) Update(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	var req dto.UpdateBrokerConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert DTO to service request
	serviceReq := &service.UpdateBrokerConnectionRequest{
		BrokerName:       req.BrokerName,
		APIKey:           req.APIKey,
		APISecret:        req.APISecret,
		Passphrase:       req.Passphrase,
		ConsumerID:       req.ConsumerID,
		ConsumerSecret:   req.ConsumerSecret,
		OTPMethod:        req.OTPMethod,
		AutoSync:         req.AutoSync,
		SyncFrequency:    req.SyncFrequency,
		SyncAssets:       req.SyncAssets,
		SyncTransactions: req.SyncTransactions,
		SyncPrices:       req.SyncPrices,
		SyncBalance:      req.SyncBalance,
		Notes:            req.Notes,
	}

	connection, err := h.service.Update(c.Request.Context(), connectionID, userID, serviceReq)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Broker connection updated successfully", dto.ToBrokerConnectionResponse(connection))
}

// Delete soft-deletes a broker connection
// @Summary Delete broker connection
// @Description Soft-delete a broker connection
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id} [delete]
func (h *BrokerConnectionHandler) Delete(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), connectionID, userID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// Activate activates a broker connection
// @Summary Activate broker connection
// @Description Activate a broker connection to enable syncing
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id}/activate [post]
func (h *BrokerConnectionHandler) Activate(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.Activate(c.Request.Context(), connectionID, userID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Broker connection activated")
}

// Deactivate deactivates a broker connection
// @Summary Deactivate broker connection
// @Description Deactivate a broker connection to disable syncing
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id}/deactivate [post]
func (h *BrokerConnectionHandler) Deactivate(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), connectionID, userID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Broker connection deactivated")
}

// RefreshToken refreshes the access token for a broker connection
// @Summary Refresh access token
// @Description Refresh the access token for a broker connection
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id}/refresh-token [post]
func (h *BrokerConnectionHandler) RefreshToken(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	connection, err := h.service.RefreshToken(c.Request.Context(), connectionID, userID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Token refreshed successfully", dto.ToBrokerConnectionResponse(connection))
}

// TestConnection tests a broker connection
// @Summary Test broker connection
// @Description Test if a broker connection is working properly
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id}/test [post]
func (h *BrokerConnectionHandler) TestConnection(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	if err := h.service.TestConnection(c.Request.Context(), connectionID, userID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Connection test successful")
}

// SyncNow manually triggers a sync for a broker connection
// @Summary Sync broker connection
// @Description Manually trigger a sync for a broker connection
// @Tags Broker Connections
// @Param id path string true "Broker Connection ID"
// @Success 200 {object} dto.SyncResultResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections/{id}/sync [post]
func (h *BrokerConnectionHandler) SyncNow(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection ID"})
		return
	}

	result, err := h.service.SyncNow(c.Request.Context(), connectionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert result to response
	response := &dto.SyncResultResponse{
		Success:            result.Success,
		SyncedAt:           result.SyncedAt,
		AssetsCount:        result.AssetsCount,
		TransactionsCount:  result.TransactionsCount,
		UpdatedPricesCount: result.UpdatedPricesCount,
		BalanceUpdated:     result.BalanceUpdated,
		Error:              result.Error,
		Details:            result.Details,
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Sync triggered successfully", response)
}

// Helper function to get user ID from context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, ErrUserIDNotFound
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try string conversion
		userIDStr, ok := userIDValue.(string)
		if !ok {
			return uuid.Nil, ErrInvalidUserID
		}
		return uuid.Parse(userIDStr)
	}

	return userID, nil
}
