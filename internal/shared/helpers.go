package shared

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// getLogger retrieves the zap logger from the Gin context
// Returns a no-op logger if not found in context
func getLogger(c *gin.Context) *zap.Logger {
	const loggerKey = "logger"
	if logger, exists := c.Get(loggerKey); exists {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			return zapLogger
		}
	}
	// Return a no-op logger if not found
	return zap.NewNop()
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, statusCode int, message string) {
	logger := getLogger(c)
	logger.Error("HTTP Error Response",
		zap.Int("status_code", statusCode),
		zap.String("error", http.StatusText(statusCode)),
		zap.String("message", message),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// RespondWithAppError sends an AppError response
func RespondWithAppError(c *gin.Context, err *AppError) {
	logger := getLogger(c)
	logger.Error("HTTP AppError Response",
		zap.Int("status_code", err.StatusCode),
		zap.String("error_code", err.Code),
		zap.String("message", err.Message),
		zap.Any("details", err.Details),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(err.StatusCode, err.ToResponse())
}

// RespondWithErrorDetails sends an error response with details
func RespondWithErrorDetails(c *gin.Context, statusCode int, message string, details map[string]interface{}) {
	logger := getLogger(c)
	logger.Error("HTTP Error Response with Details",
		zap.Int("status_code", statusCode),
		zap.String("error", http.StatusText(statusCode)),
		zap.String("message", message),
		zap.Any("details", details),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Details: details,
	})
}

// RespondWithSuccess sends a success response with data
func RespondWithSuccess[T any](c *gin.Context, statusCode int, message string, data T) {
	if message == "" {
		message = http.StatusText(statusCode)
	}

	logger := getLogger(c)
	logger.Info("HTTP Success Response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, SuccessResponse[T]{
		Status:  statusCode,
		Message: message,
		Data:    data,
	})
}

// RespondWithSuccessNoData sends a success response without data
func RespondWithSuccessNoData(c *gin.Context, statusCode int, message string) {
	if message == "" {
		message = http.StatusText(statusCode)
	}

	logger := getLogger(c)
	logger.Info("HTTP Success Response (No Data)",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, Success{
		Status:  statusCode,
		Message: message,
	})
}

// RespondWithNoContent sends an empty success response with HTTP 204 status
func RespondWithNoContent(c *gin.Context) {
	logger := getLogger(c)
	logger.Info("HTTP No Content Response",
		zap.Int("status_code", http.StatusNoContent),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.Status(http.StatusNoContent)
}

// RespondWithPagination sends a paginated response
func RespondWithPagination[T any](c *gin.Context, statusCode int, data []T, pagination Page[T]) {
	pagination.Data = data

	logger := getLogger(c)
	logger.Info("HTTP Paginated Response",
		zap.Int("status_code", statusCode),
		zap.Int64("total_items", pagination.TotalItems),
		zap.Int("total_pages", pagination.TotalPages),
		zap.Int("current_page", pagination.CurrentPage),
		zap.Int("items_per_page", pagination.ItemsPerPage),
		zap.Int("returned_items", len(data)),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, pagination)
}

// RespondWithPaginationTimeCursor sends a time-based cursor paginated response
func RespondWithPaginationTimeCursor[T any](c *gin.Context, statusCode int, data *T, pagination PaginationTimeCursor[T]) {
	pagination.Data = data

	logger := getLogger(c)
	logger.Info("HTTP Time-Cursor Paginated Response",
		zap.Int("status_code", statusCode),
		zap.String("time_cursor", pagination.TimeCursor),
		zap.Bool("has_more", pagination.HasMore),
		zap.Int("items_per_page", pagination.ItemsPerPage),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(statusCode, pagination)
}

// NewPagination creates a new pagination object
func NewPagination[T any](totalItems int64, currentPage, itemsPerPage int) Page[T] {
	if itemsPerPage <= 0 {
		itemsPerPage = 10
	}

	totalPages := int(totalItems) / itemsPerPage
	if int(totalItems)%itemsPerPage > 0 {
		totalPages++
	}

	if totalPages == 0 && totalItems > 0 {
		totalPages = 1
	}

	return Page[T]{
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		ItemsPerPage: itemsPerPage,
	}
}

// NewPaginationTimeCursor creates a new time-based cursor pagination object
func NewPaginationTimeCursor[T any](timeCursor string, hasMore bool, itemsPerPage int) PaginationTimeCursor[T] {
	if itemsPerPage <= 0 {
		itemsPerPage = 10
	}

	return PaginationTimeCursor[T]{
		TimeCursor:   timeCursor,
		HasMore:      hasMore,
		ItemsPerPage: itemsPerPage,
	}
}

// HandleError handles errors and sends appropriate response
func HandleError(c *gin.Context, err error) {
	if appErr := ToAppError(err); appErr != nil {
		RespondWithAppError(c, appErr)
		return
	}

	// Fallback to internal server error
	RespondWithError(c, http.StatusInternalServerError, "Internal server error")
}

// ParseID parses and validates an ID parameter from the URL
func ParseID(c *gin.Context, paramName string) (int64, *AppError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, ErrBadRequest.WithDetails("param", paramName).WithDetails("reason", "missing parameter")
	}

	var id int64
	var err error
	if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
		return 0, ErrBadRequest.WithDetails("param", paramName).WithDetails("reason", "invalid format")
	}

	if id <= 0 {
		return 0, ErrBadRequest.WithDetails("param", paramName).WithDetails("reason", "must be positive")
	}

	return id, nil
}
