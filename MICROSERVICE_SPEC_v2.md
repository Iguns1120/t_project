# 微服務基礎框架設計規格書 (Microservice Framework Design Specification) v2

## 1. 專案概述 (Overview)
本文件旨在定義基於 Golang 的微服務基礎框架規格。此框架將作為標準模板，用於快速構建高可用、可擴展的微服務應用。

### 1.1 核心目標
- **標準化**: 統一專案結構、依賴管理與開發規範。
- **高可用**: 整合 TiDB、Redis、RocketMQ 等分散式組件。
- **可觀測性**: 內建 Swagger、TraceID 追蹤、延遲預警與增強型健康檢查。
- **自動化**: 整合 CI/CD 流程與 Docker 容器化部署。

---

## 2. 技術堆疊 (Technology Stack)

| 類別 | 技術選型 | 說明 |
| :--- | :--- | :--- |
| **Language** | Golang (1.21+) | 高併發、強型別編譯語言 |
| **Web Framework** | **Gin** | 輕量級、高效能的 HTTP Web 框架 |
| **Database** | **TiDB** | 分散式 SQL 資料庫 (兼容 MySQL 協議) |
| **ORM / DAO** | GORM v2 | 資料庫操作封裝 |
| **Cache** | **Redis** (go-redis/v9) | 快取與分散式鎖 |
| **Message Queue** | **RocketMQ** | 非同步解耦與流量削峰 |
| **Config** | Viper | 支援 YAML/JSON 與環境變數覆蓋 |
| **Logging** | Zap (Uber) | 結構化日誌 (整合 TraceID) |
| **Documentation** | Swaggo | 自動生成 OpenAPI/Swagger 文件 |
| **Tracing** | Google UUID | 請求唯一識別碼 (TraceID) |

---

## 3. 系統架構設計 (System Architecture)

### 3.1 專案目錄結構
採用 Go 標準目錄結構 (Standard Go Layout)：

```text
.
├── cmd/server/             # 應用程式入口 (main.go)
├── configs/                # 設定檔 (config.yaml: 定義閾值、連線資訊)
├── docs/                   # Swagger 生成文件
├── internal/
│   ├── controller/         # HTTP Handlers (需處理 Context)
│   ├── service/            # 業務邏輯 (首個參數必為 context.Context)
│   ├── repository/         # 資料存取 (需接收 Context 以傳遞 TraceID)
│   ├── model/              # DTO, Entity
│   ├── middleware/         # TraceID, Logger, Recovery, Timeout
│   └── mq/                 # RocketMQ 封裝
├── pkg/
│   ├── database/           # TiDB 初始化
│   ├── redis/              # Redis 初始化
│   ├── logger/             # Zap 封裝 (支援 Context Field)
│   └── response/           # 統一 API 回應
├── deploy/                 # Dockerfile
├── go.mod
└── README.md
```

---

## 4. 核心功能模組規格 (Core Modules Spec)

### 4.1 Web 伺服器與中間件 (Gin Middleware)

#### 4.1.1 TraceID Middleware (核心追蹤)
- **功能**: 為每個請求生成唯一 UUID (`X-Trace-ID`)。
- **實作**:
    1. 檢查 Request Header 是否已有 `X-Trace-ID` (由上游傳入)。若無則生成新的。
    2. 將 TraceID 寫入 `gin.Context`。
    3. 將 TraceID 寫入 Response Header。
    4. **Context 傳遞**: 要求後續所有層級 (Service/Repo) 呼叫必須傳遞此 Context。

#### 4.1.2 Access Logger & Latency Warning (監控預警)
- **功能**: 記錄請求資訊、耗時，並執行超時/未完成預警。
- **邏輯**:
    1. 記錄 `StartTime`。
    2. 執行 `c.Next()` 處理請求。
    3. 請求結束後，計算 `Latency = Now - StartTime`。
    4. **慢查詢預警**: 若 `Latency > Config.SlowThreshold` (如 2秒)，使用 Zap 記錄 `WARN` 層級日誌，並附帶 TraceID 與相關參數。
    5. **未完成/中斷預警**: 檢查 `c.Request.Context().Err()`。若 Context 已取消 (Canceled) 或超時 (DeadlineExceeded)，記錄 `ERROR` 層級警報，提示流程未完整執行。
    6. **標準日誌**: 記錄 HTTP Status, Method, Path, IP, Latency, TraceID。

#### 4.1.3 Recovery
- 捕獲 Panic，防止服務崩潰，並記錄 Stack Trace (附帶 TraceID) 至錯誤日誌。

### 4.2 資料庫 (TiDB) 與 緩存 (Redis)
- **Context 整合**: GORM 與 go-redis 呼叫時必須使用 `WithContext(ctx)`，確保當上游 Context 取消時，資料庫查詢能被中斷，且慢查詢日誌能關聯 TraceID。

### 4.3 健康檢查 (Advanced Health Check)
- **路由**: `GET /health`
- **邏輯**: 平行檢查依賴組件。
- **指標**:
    1. **狀態 (Status)**: UP / DOWN / DEGRADED (降級)
    2. **延遲 (Latency)**: 記錄 Ping 耗時。
- **預警機制**:
    - 若 TiDB/Redis 連線耗時超過 `Config.HealthCheck.LatencyThreshold` (如 100ms)，狀態標記為 **DEGRADED**，並在回傳的 JSON 中包含警告訊息。
- **回應範例**:
    ```json
    {
      "status": "DEGRADED",
      "components": {
        "tidb": { "status": "UP", "latency": "5ms" },
        "redis": { "status": "DEGRADED", "latency": "150ms", "msg": "High latency detected" },
        "rocketmq": { "status": "UP", "latency": "10ms" }
      }
    }
    ```

---

## 5. DevOps 與 CI/CD

- **Dockerfile**: 多階段建置，優化儲存空間。
- **CI/CD**: 包含 Lint, Test, Build 流程。

---

## 6. 設定檔規劃 (configs/config.yaml)

```yaml
server:
  port: 8080
  mode: debug
  # 慢請求閾值 (毫秒)
  slow_threshold: 500 

logger:
  level: info
  encoding: json

database: # TiDB
  dsn: "user:pass@tcp(host:4000)/db?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "localhost:6379"
  
health_check:
  # 依賴服務延遲警報閾值 (毫秒)
  latency_threshold: 100
```
