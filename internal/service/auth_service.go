package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/pkg/logger"
)

// AuthService 定義認證操作的介面
type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	// 在 MVP 中，我們保持密碼處理簡單。在真實應用中，應使用 bcrypt。
	// 我們可能也需要 Register 方法，但在 MVP 中，假設玩家已存在。
}

// authService 實作 AuthService
type authService struct {
	playerRepo repository.PlayerRepository
}

// NewAuthService 建立一個新的 authService
func NewAuthService(playerRepo repository.PlayerRepository) AuthService {
	return &authService{playerRepo: playerRepo}
}

// Login 驗證玩家
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	log := logger.FromContext(ctx)

	player, err := s.playerRepo.GetPlayerByUsername(ctx, req.Username)
	if err != nil {
		log.Error("取得玩家進行登入失敗", zap.Error(err), zap.String("username", req.Username))
		return nil, fmt.Errorf("認證失敗: %w", err)
	}
	if player == nil {
		log.Warn("嘗試使用不存在的使用者名稱登入", zap.String("username", req.Username))
		return nil, fmt.Errorf("憑證無效")
	}

	// 在真實應用程式中，應使用 bcrypt 比較雜湊密碼。
	// 對於 MVP，僅進行簡單字串比較作為演示。
	if player.Password != req.Password {
		log.Warn("嘗試使用錯誤密碼登入", zap.String("username", req.Username))
		return nil, fmt.Errorf("憑證無效")
	}

	// 對於 MVP，回傳一個虛擬 Token。在真實應用中，應生成 JWT。
	token := fmt.Sprintf("mock-jwt-token-for-player-%d", player.ID)

	log.Info("玩家登入成功", zap.Uint("playerID", player.ID), zap.String("username", player.Username))
	return &model.LoginResponse{Token: token}, nil
}