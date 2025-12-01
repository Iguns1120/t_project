package controller

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/database"
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/redis"
	"microservice-mvp/pkg/rocketmq"
)

// HealthCheckResponse 定義健康檢查 API 的回應結構
type HealthCheckResponse struct {
	Status     string                     `json:"status" example:"UP"`
	Uptime     string                     `json:"uptime" example:"1h2m3s"`
	System     SystemMetrics              `json:"system"`
	Components map[string]ComponentStatus `json:"components"`
}

// SystemMetrics 保存運行時指標
type SystemMetrics struct {
	Goroutines int    `json:"goroutines" example:"12"`
	Memory     string `json:"memory_usage" example:"5 MB"`
	GoVersion  string `json:"go_version" example:"go1.23.0"`
}

// ComponentStatus 定義個別組件的狀態
type ComponentStatus struct {
	Status  string `json:"status" example:"UP"`
	Latency string `json:"latency,omitempty" example:"5ms"`
	Details string `json:"details,omitempty"` // 額外資訊 (例如: 項目數量, DB 統計)
	Message string `json:"message,omitempty"`
}

// HealthCheckController 處理健康檢查請求
type HealthCheckController struct {
	cfg       *configs.Config
	startTime time.Time
}

// NewHealthCheckController 建立一個新的 HealthCheckController
func NewHealthCheckController(cfg *configs.Config) *HealthCheckController {
	return &HealthCheckController{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// Check 處理 GET /health 請求
// @Summary 健康檢查
// @Description 檢查服務運行狀態、系統指標及依賴組件健康狀況
// @Tags System
// @Produce json
// @Success 200 {object} HealthCheckResponse "服務正常"
// @Failure 503 {object} HealthCheckResponse "服務降級或異常"
// @Router /health [get]
func (ctrl *HealthCheckController) Check(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	overallStatus := "UP"
	components := make(map[string]ComponentStatus)

	// 1. 系統指標 (始終相關)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	systemMetrics := SystemMetrics{
		Goroutines: runtime.NumGoroutine(),
		Memory:     fmt.Sprintf("%v MB", m.Alloc/1024/1024),
		GoVersion:  runtime.Version(),
	}

	// 2. 持久化檢查 (基於配置)
	if ctrl.cfg.Persistence.Type == "memory" {
		// Memory 模式檢查
		components["memory_store"] = ComponentStatus{
			Status:  "UP",
			Details: "In-Memory 持久化已啟用",
		}
	} else {
		// MySQL 模式檢查
		dbStatus := ctrl.checkTiDB(ctx, log)
		components["tidb"] = dbStatus
		if dbStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}

		// Redis 檢查
		redisStatus := ctrl.checkRedis(ctx, log)
		components["redis"] = redisStatus
		if redisStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}
	}

	// 3. RocketMQ 檢查 (僅當實際初始化時)
	if rocketmq.ProducerClient != nil && rocketmq.ProducerClient.Started() {
		mqStatus := ComponentStatus{Status: "UP"}
		components["rocketmq"] = mqStatus
	}

	httpStatus := http.StatusOK
	if overallStatus != "UP" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, HealthCheckResponse{
		Status:     overallStatus,
		Uptime:     time.Since(ctrl.startTime).String(),
		System:     systemMetrics,
		Components: components,
	})
}

func (ctrl *HealthCheckController) checkTiDB(ctx context.Context, log *zap.Logger) ComponentStatus {
	start := time.Now()
	gormDB := database.GetDB()
	if gormDB == nil {
		return ComponentStatus{Status: "DOWN", Message: "資料庫客戶端未初始化"}
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("無法獲取 TiDB 底層 DB 連線以進行健康檢查", zap.Error(err))
		return ComponentStatus{Status: "DOWN", Message: "無法獲取 DB 連線池"}
	}
	err = sqlDB.PingContext(ctx)
	latency := time.Since(start)
	status := "UP"
	message := ""

	if err != nil {
		status = "DOWN"
		message = fmt.Sprintf("Ping 失敗: %v", err)
		log.Error("TiDB 健康檢查失敗", zap.Error(err))
	} else if latency > time.Duration(ctrl.cfg.HealthCheck.LatencyThreshold)*time.Millisecond {
		status = "DEGRADED"
		message = fmt.Sprintf("高延遲: %s > %dms", latency, ctrl.cfg.HealthCheck.LatencyThreshold)
		log.Warn("TiDB 健康檢查偵測到高延遲", zap.Duration("latency", latency), zap.Int("threshold", ctrl.cfg.HealthCheck.LatencyThreshold))
	}

	return ComponentStatus{Status: status, Latency: latency.String(), Message: message}
}

func (ctrl *HealthCheckController) checkRedis(ctx context.Context, log *zap.Logger) ComponentStatus {
	start := time.Now()
	rdb := redis.GetClient()
	if rdb == nil {
		return ComponentStatus{Status: "DOWN", Message: "Redis 客戶端未初始化"}
	}

	err := rdb.Ping(ctx).Err()
	latency := time.Since(start)
	status := "UP"
	message := ""

	if err != nil {
		status = "DOWN"
		message = fmt.Sprintf("Ping 失敗: %v", err)
		log.Error("Redis 健康檢查失敗", zap.Error(err))
	} else if latency > time.Duration(ctrl.cfg.HealthCheck.LatencyThreshold)*time.Millisecond {
		status = "DEGRADED"
		message = fmt.Sprintf("高延遲: %s > %dms", latency, ctrl.cfg.HealthCheck.LatencyThreshold)
		log.Warn("Redis 健康檢查偵測到高延遲", zap.Duration("latency", latency), zap.Int("threshold", ctrl.cfg.HealthCheck.LatencyThreshold))
	}

	return ComponentStatus{Status: status, Latency: latency.String(), Message: message}
}
