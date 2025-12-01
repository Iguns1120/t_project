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

func TestAuthService_Login(t *testing.T) {
	// Setup logger for testing
	_, _ = logger.NewLogger("info", "console")

	tests := []struct {
		name          string
		loginRequest  *model.LoginRequest
		mockBehavior  func(m *mocks.MockPlayerRepository)
		expectedToken string
		expectedError string
	}{
		{
			name: "Success",
			loginRequest: &model.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByUsername", mock.Anything, "testuser").Return(&model.Player{
					ID:       1,
					Username: "testuser",
					Password: "password123",
				}, nil)
			},
			expectedToken: "mock-jwt-token-for-player-1",
			expectedError: "",
		},
		{
			name: "UserNotFound",
			loginRequest: &model.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByUsername", mock.Anything, "nonexistent").Return(nil, nil)
			},
			expectedToken: "",
			expectedError: "invalid credentials",
		},
		{
			name: "WrongPassword",
			loginRequest: &model.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByUsername", mock.Anything, "testuser").Return(&model.Player{
					ID:       1,
					Username: "testuser",
					Password: "password123",
				}, nil)
			},
			expectedToken: "",
			expectedError: "invalid credentials",
		},
		{
			name: "RepositoryError",
			loginRequest: &model.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockBehavior: func(m *mocks.MockPlayerRepository) {
				m.On("GetPlayerByUsername", mock.Anything, "testuser").Return(nil, errors.New("db error"))
			},
			expectedToken: "",
			expectedError: "authentication failed: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockPlayerRepository)
			tt.mockBehavior(mockRepo)

			authService := service.NewAuthService(mockRepo)
			resp, err := authService.Login(context.Background(), tt.loginRequest)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedToken, resp.Token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
