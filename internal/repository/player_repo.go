package repository

import (
	"context"
	"microservice-mvp/internal/model"
)

// PlayerRepository defines the interface for player data operations.
// This interface allows for swapping different storage implementations (e.g., MySQL, In-Memory).
type PlayerRepository interface {
	CreatePlayer(ctx context.Context, player *model.Player) error
	GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error)
	GetPlayerByID(ctx context.Context, id uint) (*model.Player, error)
}
