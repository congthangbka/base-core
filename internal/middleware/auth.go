package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BasicAuth returns a basic authentication middleware
// In production, replace with JWT or OAuth2
func BasicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		"admin": "admin", // username:password - should be from config in production
	})
}

// APIKeyAuth validates API key from header
func APIKeyAuth(validKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "API key is required",
				"code":    "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		// Check if API key is valid
		valid := false
		for _, key := range validKeys {
			if apiKey == key {
				valid = true
				break
			}
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid API key",
				"code":    "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BearerTokenAuth validates Bearer token from Authorization header
func BearerTokenAuth(validateToken func(token string) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
				"code":    "MISSING_AUTHORIZATION",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
				"code":    "INVALID_AUTHORIZATION_FORMAT",
			})
			c.Abort()
			return
		}

		token := parts[1]
		valid, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to validate token",
				"code":    "TOKEN_VALIDATION_ERROR",
			})
			c.Abort()
			return
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Store token in context for later use
		c.Set("token", token)
		c.Next()
	}
}
