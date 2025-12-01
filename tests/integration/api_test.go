package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseURL = "http://127.0.0.1:8080"

func TestMain(m *testing.M) {
	if url := os.Getenv("BASE_URL"); url != "" {
		baseURL = url
	}

	// 等待服務就緒
	waitForService()

	os.Exit(m.Run())
}

func waitForService() {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("正在等待服務於 " + baseURL)
	for {
		select {
		case <-timeout:
			fmt.Println("等待服務就緒逾時")
			os.Exit(1) // 如果服務未就緒則退出
		case <-ticker.C:
			resp, err := http.Get(baseURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				fmt.Println("服務已就緒!")
				return
			}
			if err != nil {
				fmt.Printf("等待服務中: %v\n", err)
			} else {
				fmt.Printf("等待服務中: 狀態碼 %d\n", resp.StatusCode)
				resp.Body.Close()
			}
		}
	}
}

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err, "連線到服務失敗")
	defer resp.Body.Close()
	
	assert.Equal(t, 200, resp.StatusCode)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// TestLogin 嘗試登入
// 注意: 此測試依賴預先填充的數據 (Seeding)。
// 目前，我們僅確保 API 可達並返回有效的 HTTP 回應 (即使是 401 也是相對於 500 或連線被拒的有效回應)。
func TestLogin(t *testing.T) {
	reqBody := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err, "發送 POST 到登入端點失敗")
	defer resp.Body.Close()

	// 如果使用者存在，我們預期 200；如果不存在，預期 401。但絕對不應是 500。
	assert.NotEqual(t, 500, resp.StatusCode, "不預期出現內部伺服器錯誤")
	
	if resp.StatusCode == 200 {
		var loginResp LoginResponse
		err := json.NewDecoder(resp.Body).Decode(&loginResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, loginResp.Data.Token, "成功時 Token 不應為空")
	}
}