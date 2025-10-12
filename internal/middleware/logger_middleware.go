package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const LoggerKey = "logger"

// LoggerMiddleware injects the zap logger into the Gin context
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(LoggerKey, logger)
		c.Next()
	}
}

// GetLogger retrieves the zap logger from the Gin context
// Returns a no-op logger if not found in context
func GetLogger(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get(LoggerKey); exists {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			return zapLogger
		}
	}
	// Return a no-op logger if not found
	return zap.NewNop()
}
