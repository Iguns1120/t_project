package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/pkg/logger"
)

// PlayerService 定義玩家資訊檢索操作的介面
type PlayerService interface {
	GetPlayerInfo(ctx context.Context, playerID uint) (*model.PlayerInfoResponse, error)
}

// playerService 實作 PlayerService
type playerService struct {
	playerRepo repository.PlayerRepository
}

// NewPlayerService 建立一個新的 PlayerService
func NewPlayerService(playerRepo repository.PlayerRepository) PlayerService {
	return &playerService{playerRepo: playerRepo}
}

// GetPlayerInfo 根據 ID 檢索玩家資訊
func (s *playerService) GetPlayerInfo(ctx context.Context, playerID uint) (*model.PlayerInfoResponse, error) {
	log := logger.FromContext(ctx)

	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Error("從 Repository 取得玩家資訊失敗", zap.Error(err), zap.Uint("playerID", playerID))
		return nil, fmt.Errorf("檢索玩家資訊失敗: %w", err)
	}
	if player == nil {
		log.Warn("找不到玩家", zap.Uint("playerID", playerID))
		return nil, fmt.Errorf("玩家不存在")
	}

	log.Info("成功取得玩家資訊", zap.Uint("playerID", playerID))
	resp := player.ToPlayerInfoResponse()
	return &resp, nil
}