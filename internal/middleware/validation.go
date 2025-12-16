package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContentTypeValidation validates Content-Type header
func ContentTypeValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET, DELETE, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip validation for requests without body
		if c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		contentType := c.Request.Header.Get("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Content-Type header is required",
				"code":    "MISSING_CONTENT_TYPE",
			})
			c.Abort()
			return
		}

		// Check if Content-Type is application/json
		if !strings.HasPrefix(contentType, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error":   "Unsupported Media Type",
				"message": "Content-Type must be application/json",
				"code":    "INVALID_CONTENT_TYPE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSizeValidation validates request body size
func RequestSizeValidation(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request Entity Too Large",
				"message": "Request body exceeds maximum size",
				"code":    "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

