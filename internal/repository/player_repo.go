package repository

import (
	"context"
	"microservice-mvp/internal/model"
)

// PlayerRepository 定義玩家資料操作的介面
// 此介面允許切換不同的儲存實作（例如 MySQL, In-Memory）
type PlayerRepository interface {
	CreatePlayer(ctx context.Context, player *model.Player) error
	GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error)
	GetPlayerByID(ctx context.Context, id uint) (*model.Player, error)
}