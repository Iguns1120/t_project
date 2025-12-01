package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository/mocks"
	"microservice-mvp/internal/service"
	"microservice-mvp/pkg/logger"
)

func TestPlayerService_GetPlayerInfo(t *testing.T) {
	// Setup logger
	_, _ = logger.NewLogger("info", "console")

	tests := []struct {
		name           string
		playerID       uint
		mockBehavior   func(m *mocks.MockPlayerRepository)
		expectedPlayer *model.PlayerInfoResponse
		expectedError  string
	}{
		{
			name:     "Success",
			playerID: 1,
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByID", mock.Anything, uint(1)).Return(&model.Player{
					ID:       1,
					Username: "testuser",
					Balance:  1000.0,
				}, nil)
			},
			expectedPlayer: &model.PlayerInfoResponse{
				ID:       1,
				Username: "testuser",
				Balance:  1000.0,
			},
			expectedError: "",
		},
		{
			name:     "PlayerNotFound",
			playerID: 2,
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByID", mock.Anything, uint(2)).Return(nil, nil)
			},
			expectedPlayer: nil,
			expectedError:  "玩家不存在",
		},
		{
			name:     "RepositoryError",
			playerID: 3,
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByID", mock.Anything, uint(3)).Return(nil, errors.New("db error"))
			},
			expectedPlayer: nil,
			expectedError:  "檢索玩家資訊失敗: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockPlayerRepository)
			tt.mockBehavior(mockRepo)

			playerService := service.NewPlayerService(mockRepo)
			resp, err := playerService.GetPlayerInfo(context.Background(), tt.playerID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedPlayer.ID, resp.ID)
				assert.Equal(t, tt.expectedPlayer.Username, resp.Username)
				assert.Equal(t, tt.expectedPlayer.Balance, resp.Balance)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
