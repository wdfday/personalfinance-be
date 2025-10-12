package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"personalfinancedss/internal/shared"
)

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := GetLogger(c)

		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.GetHeader("User-Agent")),
					zap.Stack("stacktrace"),
				)

				// Check if it's an AppError
				if appErr, ok := err.(*shared.AppError); ok {
					logger.Error("AppError panic",
						zap.String("error_code", appErr.Code),
						zap.String("message", appErr.Message),
						zap.Int("status_code", appErr.StatusCode),
					)
					shared.RespondWithAppError(c, appErr)
					c.Abort()
					return
				}

				// Generic error
				shared.RespondWithError(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()

		// Check for errors in the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("Request error",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)

			if appErr := shared.ToAppError(err); appErr != nil {
				shared.RespondWithAppError(c, appErr)
				c.Abort()
				return
			}

			shared.RespondWithError(c, http.StatusInternalServerError, "Internal server error")
			c.Abort()
		}
	}
}

// RecoveryMiddleware provides panic recovery
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger := GetLogger(c)
		logger.Error("Panic recovered in recovery middleware",
			zap.Any("error", recovered),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
			zap.Stack("stacktrace"),
		)
		shared.RespondWithError(c, http.StatusInternalServerError, "Internal server error")
		c.Abort()
	})
}
