package middleware

import (
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕獲任何 panic 並寫入 500 錯誤回應
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log := logger.FromContext(c.Request.Context())
				log.Error("捕獲 Panic",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
				response.FailWithMessage(c, http.StatusInternalServerError, "內部伺服器錯誤")
				c.Abort()
			}
		}()
		c.Next()
	}
}