package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"personalfinancedss/internal/module/analytics/budget_allocation/dto"
	"personalfinancedss/internal/module/analytics/budget_allocation/service"
	"personalfinancedss/internal/shared"
)

// Handler handles budget allocation HTTP requests
type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// NewHandler creates a new budget allocation handler
func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers budget allocation routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	analytics := router.Group("/api/v1/analytics")
	{
		analytics.POST("/budget-allocation", h.ExecuteBudgetAllocation)
		analytics.POST("/budget-allocation/generate", h.GenerateAllocations)
	}
}

// ExecuteBudgetAllocation godoc
// @Summary Execute Budget Allocation
// @Description Run Goal Programming budget allocation model to generate allocation scenarios
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.BudgetAllocationModelInput true "Budget Allocation Input"
// @Success 200 {object} dto.BudgetAllocationModelOutput
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/budget-allocation [post]
func (h *Handler) ExecuteBudgetAllocation(c *gin.Context) {
	var input dto.BudgetAllocationModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind budget allocation input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context if available and not provided
	if input.UserID == uuid.Nil {
		if userID, exists := c.Get("user_id"); exists {
			input.UserID = userID.(uuid.UUID)
		}
	}

	// Execute budget allocation
	output, err := h.service.ExecuteBudgetAllocation(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to execute budget allocation", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget allocation executed successfully", output)
}

// GenerateAllocations godoc
// @Summary Generate Budget Allocation Scenarios
// @Description Generate safe and balanced budget allocation scenarios
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.GenerateAllocationRequest true "Generate Allocation Request"
// @Success 200 {object} dto.GenerateAllocationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/budget-allocation/generate [post]
func (h *Handler) GenerateAllocations(c *gin.Context) {
	var req dto.GenerateAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind generate allocation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to model input (always return all scenarios for this endpoint)
	input := &dto.BudgetAllocationModelInput{
		UserID:          req.UserID,
		Year:            req.Year,
		Month:           req.Month,
		OverrideIncome:  req.OverrideIncome,
		UseAllScenarios: true,
	}

	// Execute budget allocation
	output, err := h.service.ExecuteBudgetAllocation(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to generate allocations", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	// Convert to response format
	response := &dto.GenerateAllocationResponse{
		UserID:         output.UserID,
		Period:         output.Period,
		TotalIncome:    output.TotalIncome,
		Scenarios:      output.Scenarios,
		IsFeasible:     output.IsFeasible,
		GlobalWarnings: output.GlobalWarnings,
		Metadata:       output.Metadata,
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Allocations generated successfully", response)
}
