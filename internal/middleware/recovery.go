package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"llm-aggregator/internal/logger"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)
		log := logger.GetLogger().With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		// Log panic with stack trace
		log.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("stack", string(debug.Stack())),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"isSuccess": false,
			"message":   "Internal server error",
			"requestId": requestID,
		})
		c.Abort()
	})
}
