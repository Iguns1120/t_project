package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"microservice-mvp/internal/model"
)

// playerRepositoryMemory 使用記憶體中的 map 實作 PlayerRepository
type playerRepositoryMemory struct {
	mu      sync.RWMutex
	players map[uint]*model.Player
	nextID  uint
}

// NewPlayerRepositoryMemory 建立一個新的 playerRepositoryMemory
func NewPlayerRepositoryMemory() PlayerRepository {
	return &playerRepositoryMemory{
		players: make(map[uint]*model.Player),
		nextID:  1,
	}
}

// CreatePlayer 在記憶體中建立一個新玩家
func (r *playerRepositoryMemory) CreatePlayer(ctx context.Context, player *model.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 檢查使用者名稱是否唯一
	for _, p := range r.players {
		if p.Username == player.Username {
			return fmt.Errorf("使用者名稱已存在") // 簡單模擬唯一性約束
		}
	}

	player.ID = r.nextID
	r.nextID++
	player.CreatedAt = time.Now()
	player.UpdatedAt = time.Now()

	// 儲存副本
	p := *player
	r.players[player.ID] = &p
	
	return nil
}

// GetPlayerByUsername 從記憶體中根據使用者名稱檢索玩家
func (r *playerRepositoryMemory) GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.players {
		if p.Username == username {
			copy := *p
			return &copy, nil
		}
	}
	return nil, nil // 找不到
}

// GetPlayerByID 從記憶體中根據 ID 檢索玩家
func (r *playerRepositoryMemory) GetPlayerByID(ctx context.Context, id uint) (*model.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.players[id]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, nil // 找不到
}