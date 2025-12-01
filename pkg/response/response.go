package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the unified API response structure.
type Response struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// OK responds with a success message and optional data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Fail responds with an error message and status code.
func Fail(c *gin.Context, httpCode int, err error) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: err.Error(),
	})
}

// FailWithMessage responds with a custom error message and status code.
func FailWithMessage(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// NewError creates a new error response.
func NewError(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// --- Swagger Documentation Helpers ---

// HTTPError400 represents a 400 Bad Request response for Swagger.
type HTTPError400 struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Bad Request"`
}

// HTTPError401 represents a 401 Unauthorized response for Swagger.
type HTTPError401 struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"Unauthorized"`
}

// HTTPError404 represents a 404 Not Found response for Swagger.
type HTTPError404 struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Not Found"`
}

// HTTPError500 represents a 500 Internal Server Error response for Swagger.
type HTTPError500 struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal Server Error"`
}
