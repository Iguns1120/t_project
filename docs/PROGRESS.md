# 微服務 MVP 開發進度報告 (Project Progress Report)

**日期**: 2025-12-01
**狀態**: 核心功能開發完成，已通過單元測試，等待集成測試。

---

## 1. 已完成項目 (Completed Features)

### 1.1 基礎架構 (Infrastructure)
- [x] **目錄結構**: 建立符合 Standard Go Project Layout 的目錄結構。
- [x] **依賴管理**: 初始化 `go.mod` 並安裝所有依賴 (Gin, GORM, Zap, Viper, RocketMQ, etc.)。
- [x] **配置管理**: 實作 `pkg/configs`，使用 Viper 載入 `configs/config.yaml`。
- [x] **日誌系統**: 實作 `pkg/logger` (Zap)，支援 TraceID 注入與 Context 傳遞。
- [x] **依賴服務封裝**:
    - `pkg/database`: GORM + TiDB (MySQL 協議) 連線池與日誌整合。
    - `pkg/redis`: go-redis 客戶端封裝。
    - `pkg/rocketmq`: Producer/Consumer 封裝 (目前為 Stub 模式以通過編譯，環境備妥後可切換回真實模式)。

### 1.2 核心業務模組 (Core Modules)
- [x] **Web 框架**: Gin 路由與 Middleware 設定。
- [x] **Middleware**:
    - `TraceID`: 全鏈路追蹤 ID 生成與傳遞。
    - `Logger`: 請求日誌、耗時監控與慢查詢預警。
    - `Recovery`: Panic 捕獲與恢復。
- [x] **API 實作**:
    - `POST /api/v1/login`: 玩家登入 (AuthService)。
    - `GET /api/v1/players/:id`: 取得玩家資料 (PlayerService)，含 Redis 緩存策略。
    - `POST /api/v1/game/bet`: 玩家下注 (GameService)，含 DB 事務與 RocketMQ 事件發送。
    - `GET /health`: 系統健康檢查，監控 TiDB/Redis 延遲與狀態。
- [x] **API 文件**: Swagger 註解已添加，文件生成腳本 `scripts/gen_swagger.bat` 已建立。

### 1.3 測試與部署 (Testing & Deployment)
- [x] **單元測試**: 完成 `internal/service` 層的單元測試，使用 `testify` 與 Mock Repository。
- [x] **Docker Compose**: 建立 `deploy/docker-compose.yaml` 用於一鍵啟動 TiDB (MySQL), Redis, RocketMQ。

---

## 2. 快速開始 (Quick Start)

### 2.1 啟動依賴環境
在專案根目錄執行：
```powershell
docker-compose -f deploy/docker-compose.yaml up -d
```
這將啟動 MySQL (port 4000), Redis (port 6379), RocketMQ (port 9876)。

### 2.2 啟動應用程式
```powershell
go run cmd/server/main.go
```
應用程式將監聽 `8080` 埠。

### 2.3 驗證
- **Health Check**: `http://localhost:8080/health`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`

---

## 3. 測試指南 (Testing Guide)

執行所有單元測試 (Mock 模式，無需真實 DB)：
```powershell
go test ./...
```

---

## 4. 下一步計畫 (Next Steps)

1.  **集成測試**: 手動或自動化驗證真實 DB/MQ 環境下的完整流程（需將 `pkg/rocketmq` 中的 Stub 代碼切換回真實代碼）。
2.  **資料填充**: 在 DB 中插入測試玩家數據，以便測試登入與下注 API。
3.  **RocketMQ Consumer**: 實作 Consumer 邏輯來處理下注事件 (例如：數據分析、日誌歸檔)。
4.  **鑑權**: 將 `AuthService` 中的 Token 替換為真實的 JWT 實作。
