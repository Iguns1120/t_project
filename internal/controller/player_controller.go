package controller

import (
	"microservice-mvp/internal/model"
	"microservice-mvp/internal/service"
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PlayerController handles player-related requests.
type PlayerController struct {
	playerService service.PlayerService
}

// NewPlayerController creates a new PlayerController.
func NewPlayerController(playerService service.PlayerService) *PlayerController {
	return &PlayerController{playerService: playerService}
}

// GetPlayerInfo handles requests to get player information by ID.
// @Summary 取得玩家資料
// @Description 根據玩家 ID 取得玩家的詳細資料，包含餘額。
// @Tags Player
// @Produce json
// @Param id path int true "玩家 ID"
// @Success 200 {object} response.Response{data=model.PlayerInfoResponse} "成功取得玩家資料"
// @Failure 400 {object} response.HTTPError400 "請求參數錯誤"
// @Failure 404 {object} response.HTTPError404 "玩家不存在"
// @Failure 500 {object} response.HTTPError500 "內部伺服器錯誤"
// @Router /api/v1/players/{id} [get]
func (ctrl *PlayerController) GetPlayerInfo(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	playerIDStr := c.Param("id")
	playerID, err := strconv.ParseUint(playerIDStr, 10, 32)
	if err != nil {
		log.Warn("Invalid player ID format", zap.Error(err), zap.String("playerIDStr", playerIDStr))
		response.FailWithMessage(c, http.StatusBadRequest, "Invalid player ID format")
		return
	}

	var resp *model.PlayerInfoResponse
	resp, err = ctrl.playerService.GetPlayerInfo(c.Request.Context(), uint(playerID))
	if err != nil {
		log.Error("Player service get player info failed", zap.Error(err), zap.Uint("playerID", uint(playerID)))
		if err.Error() == "player not found" {
			response.FailWithMessage(c, http.StatusNotFound, err.Error())
		} else {
			response.FailWithMessage(c, http.StatusInternalServerError, "Failed to retrieve player information")
		}
		return
	}

	response.OK(c, resp)
}
