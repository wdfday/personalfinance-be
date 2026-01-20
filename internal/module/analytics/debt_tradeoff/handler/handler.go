package handler

import (
	"net/http"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/dto"
	"personalfinancedss/internal/module/analytics/debt_tradeoff/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service service.Service
	logger  *zap.Logger
}

func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	analytics := router.Group("/api/v1/analytics")
	{
		analytics.POST("/savings-debt-tradeoff", h.ExecuteTradeoff)
	}
}

func (h *Handler) ExecuteTradeoff(c *gin.Context) {
	var input dto.TradeoffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	output, err := h.service.ExecuteTradeoff(c.Request.Context(), &input)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Tradeoff analysis executed successfully", output)
}
