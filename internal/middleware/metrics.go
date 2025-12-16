package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/clean-architecture/internal/metrics"
)

// Metrics middleware to collect Prometheus metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record request size
		if c.Request.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(
				c.Request.Method,
				path,
			).Observe(float64(c.Request.ContentLength))
		}

		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Observe(duration)

		// Record response size
		if c.Writer.Size() > 0 {
			metrics.HTTPResponseSize.WithLabelValues(
				c.Request.Method,
				path,
			).Observe(float64(c.Writer.Size()))
		}
	}
}

