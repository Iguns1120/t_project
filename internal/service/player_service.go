package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/pkg/logger"
)

// PlayerService defines the interface for player information retrieval operations.
type PlayerService interface {
	GetPlayerInfo(ctx context.Context, playerID uint) (*model.PlayerInfoResponse, error)
}

// playerService implements PlayerService.
type playerService struct {
	playerRepo repository.PlayerRepository
}

// NewPlayerService creates a new PlayerService.
func NewPlayerService(playerRepo repository.PlayerRepository) PlayerService {
	return &playerService{playerRepo: playerRepo}
}

// GetPlayerInfo retrieves player information by ID.
func (s *playerService) GetPlayerInfo(ctx context.Context, playerID uint) (*model.PlayerInfoResponse, error) {
	log := logger.FromContext(ctx)

	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Error("Failed to get player info from repository", zap.Error(err), zap.Uint("playerID", playerID))
		return nil, fmt.Errorf("failed to retrieve player information: %w", err)
	}
	if player == nil {
		log.Warn("Player not found", zap.Uint("playerID", playerID))
		return nil, fmt.Errorf("player not found")
	}

	log.Info("Player information retrieved successfully", zap.Uint("playerID", playerID))
	resp := player.ToPlayerInfoResponse()
	return &resp, nil
}
