package model

import (
	"time"

	"gorm.io/gorm"
)

// Player 代表系統中的玩家
type Player struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"type:varchar(100);uniqueIndex" json:"username" binding:"required"`
	Password  string         `gorm:"type:varchar(255)" json:"-" binding:"required"` // 儲存雜湊後的密碼
	Balance   float64        `gorm:"type:decimal(10,2);default:0.00" json:"balance"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LoginRequest 代表玩家登入的請求主體
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 代表玩家登入的回應主體
type LoginResponse struct {
	Token string `json:"token"` // JWT Token 的佔位符
}

// PlayerInfoResponse 代表取得玩家資訊的回應主體
type PlayerInfoResponse struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	Balance   float64 `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// ToPlayerInfoResponse 將 Player 模型轉換為 PlayerInfoResponse
func (p *Player) ToPlayerInfoResponse() PlayerInfoResponse {
	return PlayerInfoResponse{
		ID:        p.ID,
		Username:  p.Username,
		Balance:   p.Balance,
		CreatedAt: p.CreatedAt,
	}
}
