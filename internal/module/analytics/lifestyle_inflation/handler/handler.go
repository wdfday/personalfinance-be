package handler

import (
	"net/http"
	"personalfinancedss/internal/module/analytics/lifestyle_inflation/dto"
	"personalfinancedss/internal/module/analytics/lifestyle_inflation/service"
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
	router.POST("/api/v1/analytics/lifestyle-inflation", h.Execute)
}

func (h *Handler) Execute(c *gin.Context) {
	var input dto.Lifestyle_inflationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.service.Execute(c.Request.Context(), &input)
	if err != nil {
		shared.HandleError(c, err)
		return
	}
	shared.RespondWithSuccess(c, http.StatusOK, "Lifestyle inflation analysis executed successfully", output)
}
