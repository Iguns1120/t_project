package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	goRedis "github.com/redis/go-redis/v9"
	"microservice-mvp/internal/model"
	"microservice-mvp/pkg/database"
	"microservice-mvp/pkg/logger"
)

// playerRepositoryMySQL 使用 GORM 和 Redis 實作 PlayerRepository
type playerRepositoryMySQL struct {
	db  *gorm.DB
	rdb *goRedis.Client
}

// NewPlayerRepositoryMySQL 建立一個新的 playerRepositoryMySQL
func NewPlayerRepositoryMySQL(db *gorm.DB, rdb *goRedis.Client) PlayerRepository {
	return &playerRepositoryMySQL{db: db, rdb: rdb}
}

// CreatePlayer 在資料庫中建立一個新玩家
func (r *playerRepositoryMySQL) CreatePlayer(ctx context.Context, player *model.Player) error {
	log := logger.FromContext(ctx)
	if err := database.WithContext(ctx).Create(player).Error; err != nil {
		log.Error("建立玩家失敗", zap.Error(err), zap.String("username", player.Username))
		return fmt.Errorf("建立玩家失敗: %w", err)
	}
	log.Info("玩家建立成功", zap.Uint("playerID", player.ID), zap.String("username", player.Username))
	return nil
}

// GetPlayerByUsername 根據使用者名稱檢索玩家
func (r *playerRepositoryMySQL) GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error) {
	log := logger.FromContext(ctx)
	var player model.Player
	if err := database.WithContext(ctx).Where("username = ?", username).First(&player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 找不到玩家
		}
		log.Error("根據使用者名稱獲取玩家失敗", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("根據使用者名稱獲取玩家失敗: %w", err)
	}
	return &player, nil
}

// GetPlayerByID 根據 ID 檢索玩家，優先嘗試讀取 Redis 快取
func (r *playerRepositoryMySQL) GetPlayerByID(ctx context.Context, id uint) (*model.Player, error) {
	log := logger.FromContext(ctx)
	cacheKey := fmt.Sprintf("player:%d", id)
	var player model.Player

	// 嘗試從 Redis 快取獲取
	if r.rdb != nil {
		val, err := r.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			if err := json.Unmarshal(val, &player); err == nil {
				log.Debug("從 Redis 快取獲取玩家資料", zap.Uint("playerID", id))
				return &player, nil
			}
		} else if err != goRedis.Nil {
			log.Warn("從 Redis 快取獲取玩家失敗", zap.Error(err), zap.Uint("playerID", id))
		}
	}

	// 如果快取中沒有或反序列化失敗，則從資料庫獲取
	if err := database.WithContext(ctx).First(&player, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 找不到玩家
		}
		log.Error("從 DB 獲取玩家 ID 失敗", zap.Error(err), zap.Uint("playerID", id))
		return nil, fmt.Errorf("根據 ID 獲取玩家失敗: %w", err)
	}

	// 儲存到 Redis 快取
	if r.rdb != nil {
		playerBytes, err := json.Marshal(player)
		if err == nil {
			r.rdb.Set(ctx, cacheKey, playerBytes, 5*time.Minute)
		}
	}

	return &player, nil
}