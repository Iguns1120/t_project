package controller

import (
	"context"
	"fmt"
	"net/http"
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
	Status     string                  `json:"status" example:"UP"`
	Components map[string]ComponentStatus `json:"components"`
}

// ComponentStatus defines the status of an individual component.
type ComponentStatus struct {
	Status  string `json:"status" example:"UP"`
	Latency string `json:"latency,omitempty" example:"5ms"`
	Message string `json:"message,omitempty"`
}

// HealthCheckController handles health check requests.
type HealthCheckController struct {
	cfg *configs.Config
}

// NewHealthCheckController creates a new HealthCheckController.
func NewHealthCheckController(cfg *configs.Config) *HealthCheckController {
	return &HealthCheckController{cfg: cfg}
}

// Check handles GET /health requests.
// @Summary 健康檢查
// @Description 檢查服務及其依賴組件的健康狀態
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

	// Check TiDB
	if ctrl.cfg.Persistence.Type == "memory" {
		components["tidb"] = ComponentStatus{Status: "DISABLED", Message: "Running in memory mode"}
	} else {
		dbStatus := ctrl.checkTiDB(ctx, log)
		components["tidb"] = dbStatus
		if dbStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}
	}

	// Check Redis
	if ctrl.cfg.Persistence.Type == "memory" {
		components["redis"] = ComponentStatus{Status: "DISABLED", Message: "Running in memory mode"}
	} else {
		redisStatus := ctrl.checkRedis(ctx, log)
		components["redis"] = redisStatus
		if redisStatus.Status != "UP" {
			overallStatus = "DEGRADED"
		}
	}

	// Check RocketMQ Producer
	mqStatus := ctrl.checkRocketMQProducer(ctx, log)
	components["rocketmq"] = mqStatus
	// Only affect overall status if it's DOWN (not DISABLED)
	if mqStatus.Status == "DOWN" {
		overallStatus = "DEGRADED"
	}

	httpStatus := http.StatusOK
	if overallStatus != "UP" {
		httpStatus = http.StatusServiceUnavailable // 503 Service Unavailable
	}

	c.JSON(httpStatus, HealthCheckResponse{
		Status:     overallStatus,
		Components: components,
	})
}

func (ctrl *HealthCheckController) checkTiDB(ctx context.Context, log *zap.Logger) ComponentStatus {
	start := time.Now()
	// Guard against nil DB
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

func (ctrl *HealthCheckController) checkRocketMQProducer(ctx context.Context, log *zap.Logger) ComponentStatus {
	// For template, RocketMQ is optional.
	if rocketmq.ProducerClient == nil {
		return ComponentStatus{Status: "DISABLED", Message: "RocketMQ not initialized"}
	}

	if !rocketmq.ProducerClient.Started() {
		return ComponentStatus{Status: "DOWN", Message: "RocketMQ Producer stopped"}
	}
	
	return ComponentStatus{Status: "UP"}
}
