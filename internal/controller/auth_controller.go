package controller

import (
	"microservice-mvp/internal/model"
	"microservice-mvp/internal/service"
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController 處理認證相關的請求
type AuthController struct {
	authService service.AuthService
}

// NewAuthController 建立一個新的 AuthController
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Login 處理玩家登入請求
// @Summary 玩家登入
// @Description 驗證玩家憑證並返回認證 Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body model.LoginRequest true "登入請求參數"
// @Success 200 {object} response.Response{data=model.LoginResponse} "登入成功"
// @Failure 400 {object} response.HTTPError400 "請求參數錯誤"
// @Failure 401 {object} response.HTTPError401 "認證失敗"
// @Router /api/v1/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("無效的登入請求", zap.Error(err))
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	resp, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		log.Error("認證服務登入失敗", zap.Error(err))
		response.Fail(c, http.StatusUnauthorized, err)
		return
	}

	response.OK(c, resp)
}