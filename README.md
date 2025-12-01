# Go 微服務範本 (Microservice Template)

![CI Status](https://github.com/Iguns1120/t_project/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)

這是一個輕量級、可立即投入生產的微服務範本，專為快速啟動新專案而設計。
它遵循 **Standard Go Project Layout**，並內建了 CI/CD、Swagger 文件自動生成以及詳細的系統健康檢查功能。

## 🚀 核心功能

*   **雙重持久化模式 (Dual Persistence Mode)**:
    *   **In-Memory 模式**: 零外部依賴。執行 `go run cmd/server/main.go` 即可立即啟動，適合快速原型開發。
    *   **MySQL + Redis 模式**: 生產級配置，整合 GORM 與 Redis 快取。可透過 `configs/config.yaml` 輕鬆切換。
*   **清晰架構 (Clean Architecture)**: 解耦的層級設計 (Controller -> Service -> Repository)，易於維護與擴充。
*   **可觀測性 (Observability)**:
    *   結構化 JSON 日誌 (Zap)。
    *   TraceID 全鏈路追蹤。
    *   增強型健康檢查 (包含系統資源監控、Goroutine 數量、記憶體使用量)。
*   **API 文件**: 自動生成的 Swagger UI，並包含正確的錯誤碼範例。
*   **DevOps 就緒**:
    *   **Docker**: 優化的 Multi-stage `Dockerfile` (基於 Alpine, Go 1.24)。
    *   **CI/CD**: 完整的 GitHub Actions 流程，包含 Lint 檢查、單元/集成測試與 Docker 映像檔自動構建推送。

## 📂 專案結構

```
.
├── cmd/
│   └── server/         # 程式進入點 (main.go)
├── configs/            # 設定檔 (config.yaml)
├── internal/
│   ├── controller/     # HTTP 路由處理
│   ├── service/        # 核心業務邏輯
│   ├── repository/     # 資料存取層 (介面定義與實作)
│   └── model/          # 資料模型
├── pkg/                # 公用函式庫 (Logger, Response 等)
└── tests/              # 集成測試
```

## 🛠️ 快速開始

### 1. 本地執行 (In-Memory 模式)

無需安裝 Docker 或資料庫，直接運行！

```bash
# 下載專案
git clone https://github.com/Iguns1120/t_project.git
cd t_project

# 啟動伺服器
go run cmd/server/main.go
```

存取 API:
*   **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
*   **健康檢查**: [http://localhost:8080/health](http://localhost:8080/health)

### 2. 使用資料庫 (MySQL + Redis)

1.  修改 `configs/config.yaml`：
    ```yaml
    persistence:
      type: "mysql" # 將 "memory" 改為 "mysql"
    ```
2.  使用 Docker Compose 啟動依賴服務：
    ```bash
    docker-compose -f deploy/docker-compose.yaml up -d
    ```
3.  啟動伺服器：
    ```bash
    go run cmd/server/main.go
    ```

## 🧪 測試

執行所有單元測試與集成測試：

```bash
go test -v ./...
```

## 📝 API 文件更新

當您修改了 Controller 的註解後，請執行以下指令來重新生成 Swagger 文件：

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal -o docs
```

## 🏗️ 設計決策

1.  **依賴注入 (Dependency Injection)**: 採用手動注入方式 (在 `main.go` 中組裝)，避免過度依賴複雜的 DI 框架，保持啟動邏輯透明易懂。
2.  **Repository 模式**: `PlayerRepository` 定義為介面 (Interface)。這讓我們能輕鬆切換 `memory` 和 `mysql` 實作，極大地方便了單元測試與原型驗證。
3.  **設定管理**: 使用 `Viper` 管理配置，支援透過環境變數覆蓋設定 (例如 `DATABASE_DSN` 可覆蓋 `database.dsn`)，符合 Cloud Native 部署需求。
4.  **健康檢查**: `/health` 接口會根據當前運行的模式 (Memory/MySQL) 動態調整檢查項目，並提供即時的系統資源數據，讓監控更有意義。