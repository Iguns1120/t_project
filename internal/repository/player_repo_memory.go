package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"microservice-mvp/internal/model"
)

// playerRepositoryMemory implements PlayerRepository using in-memory map.
type playerRepositoryMemory struct {
	mu      sync.RWMutex
	players map[uint]*model.Player
	nextID  uint
}

// NewPlayerRepositoryMemory creates a new playerRepositoryMemory.
func NewPlayerRepositoryMemory() PlayerRepository {
	return &playerRepositoryMemory{
		players: make(map[uint]*model.Player),
		nextID:  1,
	}
}

// CreatePlayer creates a new player in memory.
func (r *playerRepositoryMemory) CreatePlayer(ctx context.Context, player *model.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for unique username
	for _, p := range r.players {
		if p.Username == player.Username {
			return fmt.Errorf("username already exists") // Simple simulation of unique constraint
		}
	}

	player.ID = r.nextID
	r.nextID++
	player.CreatedAt = time.Now()
	player.UpdatedAt = time.Now()

	// Store a copy
	p := *player
	r.players[player.ID] = &p
	
	return nil
}

// GetPlayerByUsername retrieves a player by username from memory.
func (r *playerRepositoryMemory) GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.players {
		if p.Username == username {
			copy := *p
			return &copy, nil
		}
	}
	return nil, nil // Not found
}

// GetPlayerByID retrieves a player by ID from memory.
func (r *playerRepositoryMemory) GetPlayerByID(ctx context.Context, id uint) (*model.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.players[id]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, nil // Not found
}
