package middleware

import (
	"microservice-mvp/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// HeaderXRequestID 是請求 ID 的標頭名稱
	HeaderXRequestID = "X-Request-ID"
	// HeaderXTraceID 是追蹤 ID 的標頭名稱，對應 logger.TraceIDKey
	HeaderXTraceID = "X-Trace-ID"
)

// TraceID 是一個 Gin 中間件，用於將唯一的追蹤 ID 注入請求上下文和回應標頭中
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(HeaderXTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 在回應標頭中設定 X-Trace-ID
		c.Writer.Header().Set(HeaderXTraceID, traceID)

		// 設定 X-Request-ID (可選，通常與 TraceID 相同或每個 hop 唯一的請求 ID)
		// 為了簡單起見，如果不存在，我們生成一個新的 UUID 作為 X-Request-ID，以區別於可能的分散式 traceID
		requestID := c.GetHeader(HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderXRequestID, requestID)
		
		// 將 traceID 注入上下文以供日誌使用
		ctx := c.Request.Context()
		ctx = logger.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 記錄帶有 trace ID 的初始請求
		logger.FromContext(c.Request.Context()).Info("收到請求",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()
	}
}