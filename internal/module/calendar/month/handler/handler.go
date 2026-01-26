package handler

import (
	"fmt"
	"net/http"
	"strings"

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
		months.POST("", h.createMonth)            // POST /api/v1/months - create new month
		months.GET("/current", h.getCurrentMonth) // Must be before /:month
		months.GET("/:month", h.getMonthView)
		months.POST("/:month/income", h.receiveIncome)
		months.POST("/:month/close", h.closeMonth)
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

// createMonth creates a new month
// POST /api/v1/months
func (h *Handler) createMonth(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req dto.CreateMonthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	month, err := h.service.CreateMonth(c.Request.Context(), currentUser.ID, req.Month)
	if err != nil {
		// Check if it's a "month already exists" error
		if err.Error() != "" && (err.Error() == fmt.Sprintf("month already exists: %s", req.Month) ||
			strings.Contains(err.Error(), "month already exists")) {
			shared.RespondWithError(c, http.StatusConflict, err.Error())
			return
		}
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Month created successfully", month)
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
