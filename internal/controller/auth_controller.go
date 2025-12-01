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

// AuthController handles authentication-related requests.
type AuthController struct {
	authService service.AuthService
}

// NewAuthController creates a new AuthController.
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Login handles player login requests.
// @Summary 玩家登入
// @Description 驗證玩家憑證並返回認證 token
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body model.LoginRequest true "登入請求"
// @Success 200 {object} response.Response{data=model.LoginResponse} "登入成功"
// @Failure 400 {object} response.HTTPError400 "請求參數錯誤"
// @Failure 401 {object} response.HTTPError401 "認證失敗"
// @Router /api/v1/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid login request", zap.Error(err))
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	resp, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		log.Error("Auth service login failed", zap.Error(err))
		response.Fail(c, http.StatusUnauthorized, err)
		return
	}

	response.OK(c, resp)
}
