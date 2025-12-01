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

// playerRepositoryMySQL implements PlayerRepository using GORM and Redis.
type playerRepositoryMySQL struct {
	db  *gorm.DB
	rdb *goRedis.Client
}

// NewPlayerRepositoryMySQL creates a new playerRepositoryMySQL.
func NewPlayerRepositoryMySQL(db *gorm.DB, rdb *goRedis.Client) PlayerRepository {
	return &playerRepositoryMySQL{db: db, rdb: rdb}
}

// CreatePlayer creates a new player in the database.
func (r *playerRepositoryMySQL) CreatePlayer(ctx context.Context, player *model.Player) error {
	log := logger.FromContext(ctx)
	if err := database.WithContext(ctx).Create(player).Error; err != nil {
		log.Error("Failed to create player", zap.Error(err), zap.String("username", player.Username))
		return fmt.Errorf("failed to create player: %w", err)
	}
	log.Info("Player created successfully", zap.Uint("playerID", player.ID), zap.String("username", player.Username))
	return nil
}

// GetPlayerByUsername retrieves a player by username.
func (r *playerRepositoryMySQL) GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error) {
	log := logger.FromContext(ctx)
	var player model.Player
	if err := database.WithContext(ctx).Where("username = ?", username).First(&player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Player not found
		}
		log.Error("Failed to get player by username", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to get player by username: %w", err)
	}
	return &player, nil
}

// GetPlayerByID retrieves a player by ID, trying Redis cache first.
func (r *playerRepositoryMySQL) GetPlayerByID(ctx context.Context, id uint) (*model.Player, error) {
	log := logger.FromContext(ctx)
	cacheKey := fmt.Sprintf("player:%d", id)
	var player model.Player

	// Try to get from Redis cache
	if r.rdb != nil {
		val, err := r.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			if err := json.Unmarshal(val, &player); err == nil {
				log.Debug("Player data retrieved from Redis cache", zap.Uint("playerID", id))
				return &player, nil
			}
		} else if err != goRedis.Nil {
			log.Warn("Failed to get player from Redis cache", zap.Error(err), zap.Uint("playerID", id))
		}
	}

	// Get from database if not in cache or unmarshal failed
	if err := database.WithContext(ctx).First(&player, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Player not found
		}
		log.Error("Failed to get player by ID from DB", zap.Error(err), zap.Uint("playerID", id))
		return nil, fmt.Errorf("failed to get player by ID: %w", err)
	}

	// Store in Redis cache
	if r.rdb != nil {
		playerBytes, err := json.Marshal(player)
		if err == nil {
			r.rdb.Set(ctx, cacheKey, playerBytes, 5*time.Minute)
		}
	}

	return &player, nil
}
