package handler

import (
	"net/http"
	dto2 "personalfinancedss/internal/module/identify/broker/dto"
	"personalfinancedss/internal/module/identify/broker/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BrokerConnectionHandler handles HTTP requests for broker connections
type BrokerConnectionHandler struct {
	service service.BrokerConnectionService
}

// NewBrokerConnectionHandler creates a new broker connection handler
func NewBrokerConnectionHandler(service service.BrokerConnectionService) *BrokerConnectionHandler {
	return &BrokerConnectionHandler{
		service: service,
	}
}

// RegisterRoutes registers all broker connection routes
func (h *BrokerConnectionHandler) RegisterRoutes(router *gin.Engine) {
	brokerGroup := router.Group("/api/v1/broker-connections")
	{
		brokerGroup.GET("/providers", h.ListProviders)
		brokerGroup.POST("", h.Create)
		brokerGroup.GET("", h.List)
		brokerGroup.GET("/:id", h.GetByID)
		brokerGroup.PUT("/:id", h.Update)
		brokerGroup.DELETE("/:id", h.Delete)
		brokerGroup.POST("/:id/activate", h.Activate)
		brokerGroup.POST("/:id/deactivate", h.Deactivate)
		brokerGroup.POST("/:id/refresh-token", h.RefreshToken)
		brokerGroup.POST("/:id/test", h.TestConnection)
		brokerGroup.POST("/:id/sync", h.SyncNow)
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
	response := dto2.GetBrokerProviders()
	c.JSON(http.StatusOK, response)
}

// Create creates a new broker connection
// @Summary Create a new broker connection
// @Description Create a new broker connection and authenticate with the broker
// @Tags Broker Connections
// @Accept json
// @Produce json
// @Param request body dto.CreateBrokerConnectionRequest true "Broker connection details"
// @Success 201 {object} dto.BrokerConnectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/broker-connections [post]
func (h *BrokerConnectionHandler) Create(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req dto2.CreateBrokerConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create broker connection
	connection, err := h.service.Create(c.Request.Context(), req.ToServiceRequest(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto2.ToBrokerConnectionResponse(connection))
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

	var query dto2.ListBrokerConnectionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connections, err := h.service.List(c.Request.Context(), userID, query.ToServiceFilters())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto2.ToBrokerConnectionListResponse(connections))
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto2.ToBrokerConnectionResponse(connection))
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

	var req dto2.UpdateBrokerConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connection, err := h.service.Update(c.Request.Context(), connectionID, userID, req.ToServiceRequest())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto2.ToBrokerConnectionResponse(connection))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "broker connection activated"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "broker connection deactivated"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto2.ToBrokerConnectionResponse(connection))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "connection test successful"})
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

	c.JSON(http.StatusOK, dto2.ToSyncResultResponse(result))
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
