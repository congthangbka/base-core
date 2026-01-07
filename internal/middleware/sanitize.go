package middleware

import (
	"html"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizeInput sanitizes user input to prevent XSS attacks
// This middleware should be used carefully as it may modify legitimate data
// Consider using it only for specific endpoints that need sanitization
func SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			sanitized := make([]string, len(values))
			for i, v := range values {
				sanitized[i] = html.EscapeString(v)
			}
			c.Request.URL.Query()[key] = sanitized
		}

		// Sanitize form values if present
		if c.Request.PostForm != nil {
			for key, values := range c.Request.PostForm {
				sanitized := make([]string, len(values))
				for i, v := range values {
					sanitized[i] = html.EscapeString(v)
				}
				c.Request.PostForm[key] = sanitized
			}
		}

		c.Next()
	}
}

// SanitizeString sanitizes a single string value
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// HTML escape
	s = html.EscapeString(s)
	// Trim whitespace
	s = strings.TrimSpace(s)
	return s
}

// SanitizeStrings sanitizes a slice of strings
func SanitizeStrings(strs []string) []string {
	sanitized := make([]string, len(strs))
	for i, s := range strs {
		sanitized[i] = SanitizeString(s)
	}
	return sanitized
}

