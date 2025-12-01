package middleware

import (
	"microservice-mvp/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// HeaderXRequestID is the header name for request ID.
	HeaderXRequestID = "X-Request-ID"
	// HeaderXTraceID is the header name for trace ID, which maps to logger.TraceIDKey.
	HeaderXTraceID = "X-Trace-ID"
)

// TraceID is a Gin middleware that injects a unique trace ID into the request context and response headers.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(HeaderXTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set X-Trace-ID in response header
		c.Writer.Header().Set(HeaderXTraceID, traceID)

		// Set X-Request-ID (optional, often same as TraceID or a unique request ID per hop)
		// For simplicity, we can use TraceID for X-Request-ID or generate a new one.
		// Here, we'll use a new UUID for X-Request-ID if not present, to distinguish from potential distributed traceID.
		requestID := c.GetHeader(HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderXRequestID, requestID)
		
		// Inject traceID into context for logging
		ctx := c.Request.Context()
		ctx = logger.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		// Log initial request with trace ID
		logger.FromContext(c.Request.Context()).Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()
	}
}
