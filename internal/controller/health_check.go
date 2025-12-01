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

// HealthCheckResponse defines the structure for health check API response.
type HealthCheckResponse struct {
	Status     string                     `json:"status" example:"UP"`
	Uptime     string                     `json:"uptime" example:"1h2m3s"`
	System     SystemMetrics              `json:"system"`
	Components map[string]ComponentStatus `json:"components"`
}

// SystemMetrics holds runtime metrics.
type SystemMetrics struct {
	Goroutines int    `json:"goroutines" example:"12"`
	Memory     string `json:"memory_usage" example:"5 MB"`
	GoVersion  string `json:"go_version" example:"go1.23.0"`
}

// ComponentStatus defines the status of an individual component.
type ComponentStatus struct {
	Status  string `json:"status" example:"UP"`
	Latency string `json:"latency,omitempty" example:"5ms"`
	Details string `json:"details,omitempty"` // Extra info (e.g., item count, db stats)
	Message string `json:"message,omitempty"`
}

// HealthCheckController handles health check requests.
type HealthCheckController struct {
	cfg       *configs.Config
	startTime time.Time
}

// NewHealthCheckController creates a new HealthCheckController.
func NewHealthCheckController(cfg *configs.Config) *HealthCheckController {
	return &HealthCheckController{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// Check handles GET /health requests.
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

	// 1. System Metrics (Always relevant)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	systemMetrics := SystemMetrics{
		Goroutines: runtime.NumGoroutine(),
		Memory:     fmt.Sprintf("%v MB", m.Alloc/1024/1024),
		GoVersion:  runtime.Version(),
	}

	// 2. Persistence Checks (Based on Config)
	if ctrl.cfg.Persistence.Type == "memory" {
		// Memory Mode Check
		components["memory_store"] = ComponentStatus{
			Status:  "UP",
			Details: "In-Memory persistence enabled",
		}
	} else {
		// MySQL Mode Check
		dbStatus := ctrl.checkTiDB(ctx, log)
		components["tidb"] = dbStatus
		if dbStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}

		// Redis Check
		redisStatus := ctrl.checkRedis(ctx, log)
		components["redis"] = redisStatus
		if redisStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}
	}

	// 3. RocketMQ Check (Only if actually initialized)
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
		return ComponentStatus{Status: "DOWN", Message: "Database client not initialized"}
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get TiDB underlying DB for health check", zap.Error(err))
		return ComponentStatus{Status: "DOWN", Message: "Failed to get DB connection pool"}
	}
	err = sqlDB.PingContext(ctx)
	latency := time.Since(start)
	status := "UP"
	message := ""

	if err != nil {
		status = "DOWN"
		message = fmt.Sprintf("Ping failed: %v", err)
		log.Error("TiDB health check failed", zap.Error(err))
	} else if latency > time.Duration(ctrl.cfg.HealthCheck.LatencyThreshold)*time.Millisecond {
		status = "DEGRADED"
		message = fmt.Sprintf("High latency: %s > %dms", latency, ctrl.cfg.HealthCheck.LatencyThreshold)
		log.Warn("TiDB health check detected high latency", zap.Duration("latency", latency), zap.Int("threshold", ctrl.cfg.HealthCheck.LatencyThreshold))
	}

	return ComponentStatus{Status: status, Latency: latency.String(), Message: message}
}

func (ctrl *HealthCheckController) checkRedis(ctx context.Context, log *zap.Logger) ComponentStatus {
	start := time.Now()
	rdb := redis.GetClient()
	if rdb == nil {
		return ComponentStatus{Status: "DOWN", Message: "Redis client not initialized"}
	}

	err := rdb.Ping(ctx).Err()
	latency := time.Since(start)
	status := "UP"
	message := ""

	if err != nil {
		status = "DOWN"
		message = fmt.Sprintf("Ping failed: %v", err)
		log.Error("Redis health check failed", zap.Error(err))
	} else if latency > time.Duration(ctrl.cfg.HealthCheck.LatencyThreshold)*time.Millisecond {
		status = "DEGRADED"
		message = fmt.Sprintf("High latency: %s > %dms", latency, ctrl.cfg.HealthCheck.LatencyThreshold)
		log.Warn("Redis health check detected high latency", zap.Duration("latency", latency), zap.Int("threshold", ctrl.cfg.HealthCheck.LatencyThreshold))
	}

	return ComponentStatus{Status: status, Latency: latency.String(), Message: message}
}