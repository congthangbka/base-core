package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	log := GetLogger()
	
	if requestID := GetRequestID(ctx); requestID != "" {
		log = log.With(zap.String("request_id", requestID))
	}
	
	return log
}

