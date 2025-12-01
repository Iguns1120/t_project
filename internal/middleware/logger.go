package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
)

// LoggerMiddleware is a Gin middleware for logging HTTP requests and handling latency warnings.
func LoggerMiddleware(cfg configs.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get logger with trace ID from context
		log := logger.FromContext(c.Request.Context())

		// Log details after request is processed
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		fields := []zap.Field{
			zap.Int("statusCode", statusCode),
			zap.String("latency", latency.String()),
			zap.String("clientIP", clientIP),
			zap.String("method", method),
			zap.String("path", path),
		}

		// Log request context error (e.g., client disconnected, deadline exceeded)
		select {
		case <-c.Request.Context().Done():
			err := c.Request.Context().Err()
			log.Error("Request context finished unexpectedly", zap.Error(err), zap.Any("context_error", err.Error()))
			fields = append(fields, zap.Error(err))
		default:
			// No context error
		}

		if errorMessage != "" {
			log.Error(errorMessage, fields...)
		} else if latency > time.Duration(cfg.SlowThreshold)*time.Millisecond {
			log.Warn("Slow request", fields...)
		} else {
			log.Info("Request completed", fields...)
		}
	}
}
