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

	// Wait for service to be ready
	waitForService()

	os.Exit(m.Run())
}

func waitForService() {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Waiting for service at " + baseURL)
	for {
		select {
		case <-timeout:
			fmt.Println("Timeout waiting for service to be ready")
			os.Exit(1) // Exit if service is not ready
		case <-ticker.C:
			resp, err := http.Get(baseURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				fmt.Println("Service is ready!")
				return
			}
			if err != nil {
				fmt.Printf("Waiting for service: %v\n", err)
			} else {
				fmt.Printf("Waiting for service: status code %d\n", resp.StatusCode)
				resp.Body.Close()
			}
		}
	}
}

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err, "Failed to connect to service")
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

// TestLogin attempts to login.
// Note: This test depends on seeded data. 
// For now, we just ensure the API is reachable and returns a valid HTTP response (even 401 is a valid API response compared to 500 or connection refused).
func TestLogin(t *testing.T) {
	reqBody := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err, "Failed to post to login endpoint")
	defer resp.Body.Close()

	// We expect 200 if user exists, or 401 if not. But definitely not 500.
	assert.NotEqual(t, 500, resp.StatusCode, "Internal Server Error is not expected")
	
	if resp.StatusCode == 200 {
		var loginResp LoginResponse
		err := json.NewDecoder(resp.Body).Decode(&loginResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, loginResp.Data.Token, "Token should not be empty on success")
	}
}
