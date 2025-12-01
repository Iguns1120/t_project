package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/pkg/logger"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	// For MVP, we'll keep password handling simple. In a real app, use bcrypt.
	// We might also need a Register method, but for MVP, assume players exist.
}

// authService implements AuthService.
type authService struct {
	playerRepo repository.PlayerRepository
}

// NewAuthService creates a new authService.
func NewAuthService(playerRepo repository.PlayerRepository) AuthService {
	return &authService{playerRepo: playerRepo}
}

// Login authenticates a player.
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	log := logger.FromContext(ctx)

	player, err := s.playerRepo.GetPlayerByUsername(ctx, req.Username)
	if err != nil {
		log.Error("Failed to get player for login", zap.Error(err), zap.String("username", req.Username))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if player == nil {
		log.Warn("Login attempt with non-existent username", zap.String("username", req.Username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// In a real application, compare hashed password using bcrypt.
	// For MVP, simple string comparison for demonstration.
	if player.Password != req.Password {
		log.Warn("Login attempt with incorrect password", zap.String("username", req.Username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// For MVP, return a dummy token. In real app, generate JWT.
	token := fmt.Sprintf("mock-jwt-token-for-player-%d", player.ID)

	log.Info("Player logged in successfully", zap.Uint("playerID", player.ID), zap.String("username", player.Username))
	return &model.LoginResponse{Token: token}, nil
}
