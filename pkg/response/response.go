package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 代表統一的 API 回應結構
type Response struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 回應成功訊息和可選資料
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Fail 回應錯誤訊息和狀態碼
func Fail(c *gin.Context, httpCode int, err error) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: err.Error(),
	})
}

// FailWithMessage 回應自定義錯誤訊息和狀態碼
func FailWithMessage(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// NewError 建立一個新的錯誤回應
func NewError(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// --- Swagger 文件輔助結構 ---

// HTTPError400 代表 Swagger 的 400 Bad Request 回應
type HTTPError400 struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"請求參數錯誤"`
}

// HTTPError401 代表 Swagger 的 401 Unauthorized 回應
type HTTPError401 struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"未經授權"`
}

// HTTPError404 代表 Swagger 的 404 Not Found 回應
type HTTPError404 struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"找不到資源"`
}

// HTTPError500 代表 Swagger 的 500 Internal Server Error 回應
type HTTPError500 struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"內部伺服器錯誤"`
}