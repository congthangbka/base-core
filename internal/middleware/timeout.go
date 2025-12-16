package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns a middleware that sets a timeout for the request context
// Default timeout: 30 seconds
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal completion
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either completion or timeout
		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(504, gin.H{
					"error":   "Gateway Timeout",
					"message": "Request timeout. The server did not receive a timely response.",
					"code":    "REQUEST_TIMEOUT",
				})
				c.Abort()
			}
		}
	}
}

