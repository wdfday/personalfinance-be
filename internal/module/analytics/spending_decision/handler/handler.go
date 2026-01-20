package handler

import (
	"net/http"
	"personalfinancedss/internal/module/analytics/spending_decision/dto"
	"personalfinancedss/internal/module/analytics/spending_decision/service"
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
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/v1/analytics/large-purchase", h.AnalyzeLargePurchase)
}

func (h *Handler) AnalyzeLargePurchase(c *gin.Context) {
	var input dto.LargePurchaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	output, err := h.service.ExecuteLargePurchase(c.Request.Context(), &input)
	if err != nil {
		shared.HandleError(c, err)
		return
	}
	shared.RespondWithSuccess(c, http.StatusOK, "Large purchase analysis completed", output)
}
