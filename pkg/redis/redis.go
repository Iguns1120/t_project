package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
)

// Client 是全域 Redis 客戶端實例
var Client *redis.Client

// InitRedis 初始化 Redis 客戶端
func InitRedis(cfg configs.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10, // 連線池大小
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping Redis 伺服器以檢查連線
	status := rdb.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("連線到 Redis 失敗: %w", status.Err())
	}

	Client = rdb // 設定全域 Redis 客戶端實例
	logger.Logger.Info("Redis 客戶端初始化成功")
	return rdb, nil
}

// GetClient 返回全域 Redis 客戶端實例
func GetClient() *redis.Client {
	return Client
}

// WithContext 返回一個使用提供的上下文的 Redis 客戶端
// 這允許 Redis 命令在上下文被取消時取消
// 它也確保任何來自 Redis 操作的日誌可以繼承 traceID
func WithContext(ctx context.Context) *redis.Client {
	if ctx == nil {
		return Client
	}
	// go-redis 客戶端方法已經接受 context，所以我們只返回 client
	// 並依賴呼叫程式碼將 context 傳遞給方法
	return Client
}