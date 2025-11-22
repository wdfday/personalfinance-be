package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/investment/investment_asset/dto"
	"personalfinancedss/internal/module/investment/investment_asset/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles investment asset-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new investment asset handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all investment asset routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	assets := r.Group("/api/v1/investment/assets")
	assets.Use(authMiddleware.AuthMiddleware())
	{
		assets.GET("", h.listAssets)
		assets.GET("/summary", h.getPortfolioSummary)
		assets.GET("/watchlist", h.getWatchlist)
		assets.GET("/by-type/:type", h.getAssetsByType)
		assets.POST("/prices/bulk", h.bulkUpdatePrices)
		// Sync with broker routes

	}
}

// ListAssets godoc
// @Summary List investment assets
// @Description Get a paginated list of investment assets with optional filters
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param asset_type query string false "Filter by asset type"
// @Param status query string false "Filter by status"
// @Param symbol query string false "Search by symbol"
// @Param name query string false "Search by name"
// @Param sector query string false "Filter by sector"
// @Param industry query string false "Filter by industry"
// @Param exchange query string false "Filter by exchange"
// @Param is_watchlist query boolean false "Filter watchlist items"
// @Param min_value query number false "Minimum current value"
// @Param max_value query number false "Maximum current value"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param sort_by query string false "Sort by field (current_value, unrealized_gain, symbol)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} dto.AssetListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets [get]
func (h *Handler) listAssets(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var query dto.ListAssetsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	response, err := h.service.ListAssets(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Assets retrieved successfully", response)
}

// GetAsset godoc
// @Summary Get asset by ID
// @Description Get detailed information about a specific investment asset
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Success 200 {object} dto.AssetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets/{id} [get]
func (h *Handler) getAsset(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	assetID := c.Param("id")

	asset, err := h.service.GetAsset(c.Request.Context(), user.ID.String(), assetID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	response := dto.ToAssetResponse(asset)
	shared.RespondWithSuccess(c, http.StatusOK, "Asset retrieved successfully", response)
}

// GetPortfolioSummary godoc
// @Summary Get portfolio summary
// @Description Get an overview of the entire investment portfolio
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.PortfolioSummary
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets/summary [get]
func (h *Handler) getPortfolioSummary(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	summary, err := h.service.GetPortfolioSummary(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Portfolio summary retrieved successfully", summary)
}

// GetWatchlist godoc
// @Summary Get watchlist
// @Description Get all assets on the watchlist
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.AssetResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets/watchlist [get]
func (h *Handler) getWatchlist(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	assets, err := h.service.GetWatchlist(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	responses := make([]dto.AssetResponse, 0, len(assets))
	for _, asset := range assets {
		responses = append(responses, dto.ToAssetResponse(asset))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Watchlist retrieved successfully", responses)
}

// GetAssetsByType godoc
// @Summary Get assets by type
// @Description Get all assets of a specific type
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type path string true "Asset type"
// @Success 200 {array} dto.AssetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets/by-type/{type} [get]
func (h *Handler) getAssetsByType(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	assetType := c.Param("type")

	assets, err := h.service.GetAssetsByType(c.Request.Context(), user.ID.String(), assetType)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	responses := make([]dto.AssetResponse, 0, len(assets))
	for _, asset := range assets {
		responses = append(responses, dto.ToAssetResponse(asset))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Assets retrieved successfully", responses)
}

// BulkUpdatePrices godoc
// @Summary Bulk update asset prices
// @Description Update prices for multiple assets at once
// @Tags investment-assets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param prices body dto.BulkUpdatePricesRequest true "Price updates"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/assets/prices/bulk [post]
func (h *Handler) bulkUpdatePrices(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req dto.BulkUpdatePricesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	if err := h.service.BulkUpdatePrices(c.Request.Context(), req); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Prices updated successfully")
}
