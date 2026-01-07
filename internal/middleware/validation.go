package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"llm-aggregator/internal/common"
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
			common.RespondFailWithMessage(c, common.ErrorCodeBadRequest, "Content-Type header is required")
			c.Abort()
			return
		}

		// Check if Content-Type is application/json
		if !strings.HasPrefix(contentType, "application/json") {
			common.RespondFailWithMessage(c, common.ErrorCodeBadRequest, "Content-Type must be application/json")
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
			common.RespondFailWithMessage(c, common.ErrorCodeBadRequest, "Request body exceeds maximum size")
			c.Abort()
			return
		}

		c.Next()
	}
}
