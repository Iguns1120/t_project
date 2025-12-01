package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
)

// LoggerMiddleware 是 Gin 的中間件，用於記錄 HTTP 請求並處理延遲警告
func LoggerMiddleware(cfg configs.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 處理請求
		c.Next()

		// 從上下文中獲取帶有 TraceID 的 Logger
		log := logger.FromContext(c.Request.Context())

		// 請求處理完成後記錄詳細資訊
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

		// 記錄請求上下文錯誤（例如：客戶端斷線、逾時）
		select {
		case <-c.Request.Context().Done():
			err := c.Request.Context().Err()
			log.Error("請求上下文意外結束", zap.Error(err), zap.Any("context_error", err.Error()))
			fields = append(fields, zap.Error(err))
		default:
			// 無上下文錯誤
		}

		if errorMessage != "" {
			log.Error(errorMessage, fields...)
		} else if latency > time.Duration(cfg.SlowThreshold)*time.Millisecond {
			log.Warn("慢請求", fields...)
		} else {
			log.Info("請求完成", fields...)
		}
	}
}