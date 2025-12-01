package mocks

import (
	"context"
	"microservice-mvp/internal/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockPlayerRepository is a mock implementation of repository.PlayerRepository
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) CreatePlayer(ctx context.Context, player *model.Player) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetPlayerByUsername(ctx context.Context, username string) (*model.Player, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerByID(ctx context.Context, id uint) (*model.Player, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Player), args.Error(1)
}

func (m *MockPlayerRepository) UpdatePlayerBalance(ctx context.Context, tx *gorm.DB, playerID uint, amount float64) error {
	args := m.Called(ctx, tx, playerID, amount)
	return args.Error(0)
}
