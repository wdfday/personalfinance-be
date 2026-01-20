package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/calendar/month/dto"
	"personalfinancedss/internal/module/calendar/month/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the month module
type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// New Handler creates a new month handler
func NewHandler(svc service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.Named("calendar.month.handler"),
	}
}

// RegisterRoutes registers the month API routes
// Note: Routes changed from nested /budgets/:budgetId/months to /months
// to avoid route conflict with budget module's /:id parameter
func (h *Handler) RegisterRoutes(router *gin.Engine, auth *middleware.Middleware) {
	months := router.Group("/api/v1/months")
	months.Use(auth.AuthMiddleware(middleware.WithEmailVerified()))
	{
		months.GET("", h.listMonths)
		months.GET("/current", h.getCurrentMonth) // Must be before /:month
		months.GET("/:month", h.getMonthView)
		months.POST("/:month/assign", h.assignCategory)
		months.POST("/:month/move", h.moveMoney)
		months.POST("/:month/income", h.receiveIncome)
		months.POST("/:month/close", h.closeMonth)

		// Planning iterations
		months.POST("/:month/recalculate", h.recalculatePlanning)
	}

	// Sequential DSS Workflow routes
	h.RegisterDSSWorkflowRoutes(months, auth)
}

// getMonthView retrieves the full budget grid for a month
// GET /api/v1/months/:month
func (h *Handler) getMonthView(c *gin.Context) {

	month := c.Param("month")
	if month == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "Month parameter is required")
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	view, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, month)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Month view retrieved successfully", view)
}

// listMonths lists all months for the user
// GET /api/v1/months
func (h *Handler) listMonths(c *gin.Context) {

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	months, err := h.service.ListMonths(c.Request.Context(), currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Months retrieved successfully", gin.H{
		"months": months,
		"count":  len(months),
	})
}

// assignCategory assigns money to a category
// POST /api/v1/months/:month/assign
func (h *Handler) assignCategory(c *gin.Context) {

	monthStr := c.Param("month")

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// First get the month ID
	monthView, err := h.service.GetMonth(c.Request.Context(), currentUser.ID, monthStr)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req dto.AssignCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.AssignCategory(c.Request.Context(), req, &currentUser.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Category assigned successfully", gin.H{})
}

// moveMoney moves money between categories
// POST /api/v1/months/:month/move
func (h *Handler) moveMoney(c *gin.Context) {

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

	var req dto.MoveMoneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.MoveMoney(c.Request.Context(), req, &currentUser.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Money moved successfully", gin.H{})
}

// receiveIncome adds income to TBB
// POST /api/v1/months/:month/income
func (h *Handler) receiveIncome(c *gin.Context) {

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

	var req dto.IncomeReceivedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	if err := h.service.ReceiveIncome(c.Request.Context(), req, &currentUser.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Income received successfully", gin.H{})
}

// closeMonth closes a month for reporting
// POST /api/v1/months/:month/close
func (h *Handler) closeMonth(c *gin.Context) {

	monthStr := c.Param("month")
	if monthStr == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "Month parameter is required")
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	if err := h.service.CloseMonth(c.Request.Context(), currentUser.ID, monthStr); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess[any](c, http.StatusOK, "Month closed successfully", nil)
}

// getCurrentMonth gets or creates the current month
// GET /api/v1/months/current
func (h *Handler) getCurrentMonth(c *gin.Context) {

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	view, err := h.service.GetOrCreateCurrentMonth(c.Request.Context(), currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Current month retrieved successfully", view)
}

// recalculatePlanning creates a new planning iteration
// POST /api/v1/months/:month/recalculate
func (h *Handler) recalculatePlanning(c *gin.Context) {

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

	var req dto.RecalculatePlanningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	req.MonthID = monthView.MonthID

	result, err := h.service.RecalculatePlanning(c.Request.Context(), req, &currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Planning iteration created successfully", result)
}
