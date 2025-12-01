# 微服務基礎框架設計規格書 (Microservice Framework Design Specification) v3

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
| **Documentation** | **Swaggo** | 自動生成 OpenAPI/Swagger 文件 |
| **Tracing** | Google UUID | 請求唯一識別碼 (TraceID) |

---

## 3. 系統架構設計 (System Architecture)

### 3.1 專案目錄結構
採用 Go 標準目錄結構 (Standard Go Layout)：

```text
.
├── cmd/server/             # 應用程式入口 (main.go)
├── configs/                # 設定檔 (config.yaml)
├── docs/                   # Swagger 生成文件 (docs.go, swagger.json)
├── internal/
│   ├── controller/         # HTTP Handlers (需包含 Swagger 註解)
│   ├── service/            # 業務邏輯
│   ├── repository/         # 資料存取
│   ├── model/              # DTO, Entity
│   ├── middleware/         # TraceID, Logger, Recovery, Timeout
│   └── mq/                 # RocketMQ 封裝
├── pkg/
│   ├── database/           # TiDB 初始化
│   ├── redis/              # Redis 初始化
│   ├── logger/             # Zap 封裝
│   └── response/           # 統一 API 回應
├── scripts/                # 自動化腳本
│   └── gen_swagger.bat     # Swagger 生成腳本
├── deploy/                 # Dockerfile
├── go.mod
└── README.md
```

---

## 4. 核心功能模組規格 (Core Modules Spec)

### 4.1 API 文件 (Swagger/OpenAPI)

#### 4.1.1 自動化生成
- **工具**: 使用 `github.com/swaggo/swag/cmd/swag`。
- **腳本**: 專案根目錄/scripts 下提供 `gen_swagger.bat`。
- **指令邏輯**:
  ```bat
  swag init -g cmd/server/main.go -o docs
  ```
  這將掃描 `main.go` (API 通用資訊) 與所有 Controller 的註解，更新 `docs/` 目錄。

#### 4.1.2 註解規範 (Annotation Standard)
所有 `internal/controller` 下的 Handler 必須包含以下註解：
```go
// GetUser 取得使用者資訊
// @Summary 取得使用者資訊
// @Description 根據 ID 取得詳細資訊
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Router /api/v1/users/{id} [get]
func GetUser(c *gin.Context) { ... }
```

#### 4.1.3 路由掛載
- 在 `main.go` 或 `router` 初始化中，必須引用 `gin-swagger` 並掛載：
  ```go
  import "github.com/swaggo/gin-swagger"
  import "github.com/swaggo/files"
  
  r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
  ```

### 4.2 Web 伺服器與中間件

#### 4.2.1 TraceID Middleware
- 生成 `X-Trace-ID` 並注入 Context 與 Response Header。

#### 4.2.2 Access Logger & Latency Warning
- 記錄 Latency，超過 `slow_threshold` (如 500ms) 則 Log WARN。
- 監控 Context 狀態，若 Canceled/DeadlineExceeded 則 Log ERROR。

### 4.3 資料庫與緩存
- 強制使用 `WithContext(ctx)` 傳遞 TraceID。

### 4.4 健康檢查 (Advanced)
- `GET /health` 回傳依賴組件狀態與延遲。
- 延遲過高則標記為 `DEGRADED`。

---

## 5. DevOps 與 CI/CD

- **Dockerfile**: Multi-stage build。
- **CI/CD**: Lint, Test, Build。

---

## 6. 設定檔 (configs/config.yaml)

```yaml
server:
  port: 8080
  mode: debug
  slow_threshold: 500 

logger:
  level: info
  encoding: json

database:
  dsn: "user:pass@tcp(host:4000)/db?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "localhost:6379"
  
health_check:
  latency_threshold: 100
```
