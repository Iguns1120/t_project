package model

import (
	"time"

	"gorm.io/gorm"
)

// Player represents a player in the system.
type Player struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"type:varchar(100);uniqueIndex" json:"username" binding:"required"`
	Password  string         `gorm:"type:varchar(255)" json:"-" binding:"required"` // Store hashed password
	Balance   float64        `gorm:"type:decimal(10,2);default:0.00" json:"balance"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LoginRequest represents the request body for player login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response body for player login.
type LoginResponse struct {
	Token string `json:"token"` // Placeholder for JWT token
}

// PlayerInfoResponse represents the response body for getting player information.
type PlayerInfoResponse struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	Balance   float64 `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// ToPlayerInfoResponse converts a Player model to PlayerInfoResponse.
func (p *Player) ToPlayerInfoResponse() PlayerInfoResponse {
	return PlayerInfoResponse{
		ID:        p.ID,
		Username:  p.Username,
		Balance:   p.Balance,
		CreatedAt: p.CreatedAt,
	}
}