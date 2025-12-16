package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/example/clean-architecture/internal/logger"
)

const (
	slowRequestThreshold = 1 * time.Second
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID := GetRequestID(c)

		c.Next()

		latency := time.Since(start)
		log := logger.GetLogger().With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Log slow requests as warning
		if latency > slowRequestThreshold {
			log.Warn("Slow request detected")
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Error("Request error",
					zap.Error(e.Err),
					zap.Int("error_type", int(e.Type)),
				)
			}
		} else {
			// Only log successful requests at info level, errors are logged above
			if c.Writer.Status() < 400 {
				log.Info("Request completed")
			} else {
				log.Warn("Request failed",
					zap.Int("status", c.Writer.Status()),
				)
			}
		}
	}
}
