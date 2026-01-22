package handler

import (
	"errors"
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/calendar/month/dto"
	"personalfinancedss/internal/module/calendar/month/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// handleDSSError handles DSS-specific errors with proper HTTP status codes
func handleDSSError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrDSSNotInitialized) {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	shared.HandleError(c, err)
}

// RegisterDSSWorkflowRoutes registers the new sequential 5-step DSS workflow routes
func (h *Handler) RegisterDSSWorkflowRoutes(months *gin.RouterGroup, auth *middleware.Middleware) {
	// Initialize DSS (MUST be called first)
	months.POST("/:month/dss/initialize", h.initializeDSS)

	// Step 0: Auto-Scoring Preview (optional)
	months.POST("/:month/auto-scoring/preview", h.previewAutoScoring)

	// Step 1: Goal Prioritization
	months.POST("/:month/goal-prioritization/preview", h.previewGoalPrioritization)
	months.POST("/:month/goal-prioritization/apply", h.applyGoalPrioritization)

	// Step 2: Debt Strategy
	months.POST("/:month/debt-strategy/preview", h.previewDebtStrategy)
	months.POST("/:month/debt-strategy/apply", h.applyDebtStrategy)
	// Alias routes (đúng theo swagger /dss/...) để FE không bị 404
	months.POST("/:month/dss/debt-strategy/preview", h.previewDebtStrategy)
	months.POST("/:month/dss/debt-strategy/apply", h.applyDebtStrategy)

	// Step 3: Goal-Debt Trade-off
	months.POST("/:month/goal-debt-tradeoff/preview", h.previewGoalDebtTradeoff)
	months.POST("/:month/goal-debt-tradeoff/apply", h.applyGoalDebtTradeoff)

	// Step 4: Budget Allocation
	months.POST("/:month/budget-allocation/preview", h.previewBudgetAllocation)
	months.POST("/:month/budget-allocation/apply", h.applyBudgetAllocation)

	// Workflow management
	months.GET("/:month/workflow/status", h.getDSSWorkflowStatus)
	months.POST("/:month/workflow/reset", h.resetDSSWorkflow)

	// DSS Finalization (Approach 2: Apply All at Once)
	months.POST("/:month/dss/finalize", h.finalizeDSS)
}

// ==================== Initialize DSS Workflow ====================

// initializeDSS godoc
// @Summary Initialize DSS workflow
// @Description Create a new DSS session with input snapshot cached in Redis (3h TTL)
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.InitializeDSSRequest true "Initialize request with input snapshot"
// @Success 201 {object} dto.InitializeDSSResponse
// @Router /api/v1/months/{month}/dss/initialize [post]
func (h *Handler) initializeDSS(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.InitializeDSSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.InitializeDSS(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// 201 Created vì đây là session DSS mới trong Redis
	shared.RespondWithSuccess(c, http.StatusCreated, "DSS workflow initialized", result)
}

// ==================== Step 0: Auto-Scoring Preview ====================

// previewAutoScoring godoc
// @Summary Preview auto-scoring
// @Description Run goal auto-scoring and return scored goals (preview only, not saved)
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.PreviewAutoScoringRequest true "Preview request"
// @Success 200 {object} dto.PreviewAutoScoringResponse
// @Router /api/v1/months/{month}/dss/auto-scoring/preview [post]
func (h *Handler) previewAutoScoring(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.PreviewAutoScoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.PreviewAutoScoring(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Auto-scoring preview generated", result)
}

// ==================== Step 1: Goal Prioritization ====================

// previewGoalPrioritization godoc
// @Summary Preview goal prioritization
// @Description Run AHP goal prioritization and return preview without saving
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.PreviewGoalPrioritizationRequest true "Preview request"
// @Success 200 {object} dto.PreviewGoalPrioritizationResponse
// @Router /api/v1/months/{month}/dss/goal-prioritization/preview [post]
func (h *Handler) previewGoalPrioritization(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get month ID
	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.PreviewGoalPrioritizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.PreviewGoalPrioritization(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal prioritization preview generated", result)
}

// applyGoalPrioritization godoc
// @Summary Apply goal prioritization
// @Description Save user's accepted goal ranking to month state
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.ApplyGoalPrioritizationRequest true "Apply request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/months/{month}/dss/goal-prioritization/apply [post]
func (h *Handler) applyGoalPrioritization(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.ApplyGoalPrioritizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.ApplyGoalPrioritization(c.Request.Context(), req, &currentUser.ID); err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal prioritization applied (Step 1 complete)", gin.H{})
}

// ==================== Step 2: Debt Strategy ====================

// previewDebtStrategy godoc
// @Summary Preview debt strategy
// @Description Run debt repayment simulations and return scenarios
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.PreviewDebtStrategyRequest true "Preview request"
// @Success 200 {object} dto.PreviewDebtStrategyResponse
// @Router /api/v1/months/{month}/dss/debt-strategy/preview [post]
func (h *Handler) previewDebtStrategy(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.PreviewDebtStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.PreviewDebtStrategy(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt strategy preview generated", result)
}

// applyDebtStrategy godoc
// @Summary Apply debt strategy
// @Description Save user's selected debt strategy to month state
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.ApplyDebtStrategyRequest true "Apply request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/months/{month}/dss/debt-strategy/apply [post]
func (h *Handler) applyDebtStrategy(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.ApplyDebtStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.ApplyDebtStrategy(c.Request.Context(), req, &currentUser.ID); err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt strategy applied (Step 2 complete)", gin.H{})
}

// ==================== Step 3: Goal-Debt Trade-off ====================

// previewGoalDebtTradeoff godoc
// @Summary Preview goal-debt tradeoff
// @Description Run Monte Carlo trade-off analysis
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.PreviewGoalDebtTradeoffRequest true "Preview request"
// @Success 200 {object} dto.PreviewGoalDebtTradeoffResponse
// @Router /api/v1/months/{month}/dss/goal-debt-tradeoff/preview [post]
func (h *Handler) previewGoalDebtTradeoff(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.PreviewGoalDebtTradeoffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.PreviewGoalDebtTradeoff(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal-debt trade-off preview generated", result)
}

// applyGoalDebtTradeoff godoc
// @Summary Apply goal-debt tradeoff
// @Description Save user's allocation decision to month state
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.ApplyGoalDebtTradeoffRequest true "Apply request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/months/{month}/dss/goal-debt-tradeoff/apply [post]
func (h *Handler) applyGoalDebtTradeoff(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.ApplyGoalDebtTradeoffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.ApplyGoalDebtTradeoff(c.Request.Context(), req, &currentUser.ID); err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal-debt trade-off applied (Step 3 complete)", gin.H{})
}

// ==================== Step 4: Budget Allocation ====================

// previewBudgetAllocation godoc
// @Summary Preview budget allocation
// @Description Run Goal Programming allocation with results from steps 1-3
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.PreviewBudgetAllocationRequest true "Preview request"
// @Success 200 {object} dto.PreviewBudgetAllocationResponse
// @Router /api/v1/months/{month}/dss/budget-allocation/preview [post]
func (h *Handler) previewBudgetAllocation(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.PreviewBudgetAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.PreviewBudgetAllocation(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget allocation preview generated", result)
}

// applyBudgetAllocation godoc
// @Summary Apply budget allocation
// @Description Apply selected allocation scenario to month state categories
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.ApplyBudgetAllocationRequest true "Apply request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/months/{month}/dss/budget-allocation/apply [post]
func (h *Handler) applyBudgetAllocation(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.ApplyBudgetAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.ApplyBudgetAllocation(c.Request.Context(), req, &currentUser.ID); err != nil {
		handleDSSError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget allocation applied - DSS workflow complete! (Step 4 complete)", gin.H{})
}

// ==================== Workflow Management ====================

// getDSSWorkflowStatus godoc
// @Summary Get DSS workflow status
// @Description Get the current state of the DSS workflow
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Success 200 {object} dto.DSSWorkflowStatusResponse
// @Router /api/v1/months/{month}/dss/workflow/status [get]
func (h *Handler) getDSSWorkflowStatus(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	monthID := monthView.MonthID

	status, err := h.service.GetDSSWorkflowStatus(c.Request.Context(), monthID, &currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Workflow status retrieved", status)
}

// resetDSSWorkflow godoc
// @Summary Reset DSS workflow
// @Description Clear all DSS workflow results and start over
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/months/{month}/dss/workflow/reset [post]
func (h *Handler) resetDSSWorkflow(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	req := dto.ResetDSSWorkflowRequest{
		MonthID: monthView.MonthID,
	}

	if err := h.service.ResetDSSWorkflow(c.Request.Context(), req, &currentUser.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "DSS workflow reset successfully", gin.H{})
}

// ==================== DSS Finalization (Apply All at Once) ====================

// finalizeDSS godoc
// @Summary Finalize DSS workflow
// @Description Apply all DSS results at once and create a new MonthState version (Approach 2)
// @Tags month-dss
// @Accept json
// @Produce json
// @Param month path string true "Month (YYYY-MM)"
// @Param request body dto.FinalizeDSSRequest true "Finalize request with all DSS selections"
// @Success 200 {object} dto.FinalizeDSSResponse
// @Router /api/v1/months/{month}/dss/finalize [post]
func (h *Handler) finalizeDSS(c *gin.Context) {
	monthStr := c.Param("month")
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Load month to get monthID
	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Bind request
	var req dto.FinalizeDSSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.HandleError(c, shared.ErrBadRequest.WithError(err))
		return
	}

	// Call service
	response, err := h.service.FinalizeDSS(c.Request.Context(), req, monthView.MonthID, &currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, response.Message, response)
}
