package handler

import (
	"net/http"
	"personalfinancedss/internal/module/analytics/emergency_fund/dto"
	"personalfinancedss/internal/module/analytics/emergency_fund/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
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
	router.POST("/api/v1/analytics/emergency-fund", h.Execute)
}

func (h *Handler) Execute(c *gin.Context) {
	var input dto.Emergency_fundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.service.Execute(c.Request.Context(), &input)
	if err != nil {
		shared.HandleError(c, err)
		return
	}
	shared.RespondWithSuccess(c, http.StatusOK, "Emergency fund calculation executed successfully", output)
}
