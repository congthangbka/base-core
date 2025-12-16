package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second
// burst: maximum burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}

	// Cleanup old limiters periodically
	go rl.cleanup()

	return rl
}

// GetLimiter returns a rate limiter for the given key (IP address)
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double check after acquiring write lock
		limiter, exists = rl.limiters[key]
		if !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[key] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// cleanup removes old limiters periodically to prevent memory leak
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, limiter := range rl.limiters {
			// Remove limiters that haven't been used recently
			// This is a simple approach - in production, you might want
			// to track last access time more precisely
			if limiter.Allow() {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that rate limits requests per IP
// Default: 100 requests per second, burst of 200
func RateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(100, 200)
	return RateLimitWithLimiter(limiter)
}

// RateLimitWithConfig returns a middleware with custom rate limit
func RateLimitWithConfig(rps float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)
	return RateLimitWithLimiter(limiter)
}

// RateLimitWithLimiter returns a middleware using the provided limiter
func RateLimitWithLimiter(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get or create limiter for this IP
		ipLimiter := limiter.GetLimiter(clientIP)

		// Check if request is allowed
		if !ipLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

