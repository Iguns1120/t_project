# Go 微服務範本 (Microservice Template)

![CI Status](https://github.com/Iguns1120/t_project/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)

這是一個輕量級、可立即投入生產的微服務範本，專為快速啟動新專案而設計。
它遵循 **Standard Go Project Layout**，並內建了 CI/CD、Swagger 文件自動生成以及詳細的系統健康檢查功能。
所有代碼註解與 API 文件皆已**繁體中文化**，大幅降低維護與交接門檻。

## 🚀 核心功能

*   **雙重持久化模式 (Dual Persistence Mode)**:
    *   **In-Memory 模式**: 零外部依賴。執行 `go run cmd/server/main.go` 即可立即啟動，適合快速原型開發。
    *   **MySQL + Redis 模式**: 生產級配置，整合 GORM 與 Redis 快取。可透過 `configs/config.yaml` 輕鬆切換。
*   **清晰架構 (Clean Architecture)**: 解耦的層級設計 (Controller -> Service -> Repository)，易於維護與擴充。
*   **可觀測性 (Observability)**:
    *   結構化 JSON 日誌 (Zap)。
    *   TraceID 全鏈路追蹤。
    *   增強型健康檢查 (包含系統資源監控、Goroutine 數量、記憶體使用量)。
*   **API 文件**: 自動生成的 Swagger UI，包含完整的中文描述與正確的錯誤碼範例。
*   **DevOps 就緒**:
    *   **Docker**: 優化的 Multi-stage `Dockerfile` (基於 Alpine, Go 1.24)。
    *   **CI/CD**: 完整的 GitHub Actions 流程，包含 Lint 檢查、單元/集成測試與 Docker 映像檔自動構建推送。

## 📂 專案結構

```
.
├── cmd/
│   └── server/         # 程式進入點 (main.go)
├── configs/            # 設定檔 (config.yaml) - 含詳細中文註解
├── internal/
│   ├── controller/     # HTTP 路由處理
│   ├── service/        # 核心業務邏輯
│   ├── repository/     # 資料存取層 (介面定義與實作)
│   └── model/          # 資料模型
├── pkg/                # 公用函式庫 (Logger, Response 等)
├── tests/              # 集成測試
├── start_service.bat   # [Windows] 一鍵編譯並啟動服務
└── stop_service.bat    # [Windows] 一鍵停止服務
```

## 🛠️ 快速開始

### 1. 本地執行 (In-Memory 模式)

無需安裝 Docker 或資料庫，直接運行！

**Windows 使用者 (推薦):**
雙擊 `start_service.bat` 即可自動編譯並在背景啟動服務。

**手動執行:**
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

## 💡 設計思路與演進 (Design Philosophy)

本專案經歷了從「特定業務 MVP」到「通用微服務範本」的重構過程，以下是核心的設計考量與演進紀錄：

### 1. 從 MVP 到通用範本 (Evolution to Template)
最初，本專案包含特定的遊戲下注邏輯（Game/Bet Service）。為了使其成為一個可複用的通用範本（Starter Kit），我們執行了以下決策：
*   **業務抽離**: 移除了特定領域邏輯，僅保留通用的「用戶認證 (Auth)」與「玩家資訊 (Player)」作為 CRUD 範例。
*   **零依賴優先**: 引入 `In-Memory` 模式，確保開發者在 `clone` 專案後，無需安裝 Docker 或資料庫即可直接運行 (`go run`)，大幅降低上手門檻。

### 2. 雙模式架構 (Dual Mode Architecture)
為了同時滿足「快速原型開發」與「生產環境部署」，我們設計了可切換的持久化層：
*   **介面驅動 (Interface-Driven)**: 透過定義 `PlayerRepository` 介面，將業務邏輯與底層儲存解耦。
*   **動態注入**: 在 `main.go` 啟動時，根據 `config.yaml` 中的 `persistence.type` 動態決定注入 `MemoryRepo` 或 `MySQLRepo`。
*   **健康檢查適配**: `/health` 接口會感知當前模式，在 Memory 模式下自動隱藏 DB/Redis 的檢查項，避免誤報錯誤，僅顯示系統核心指標（Uptime, Memory, Goroutines）。

### 3. 完整的中文化 (Localization)
為了提升團隊協作效率與可維護性，我們執行了全面的中文化工程：
*   **代碼註解**: 所有 Go 檔案的註解皆翻譯為清晰的繁體中文，幫助開發者快速理解邏輯。
*   **API 文件**: Swagger 註解同步中文化，並修正了錯誤碼範例（不再讓 400/500 錯誤顯示 200 範例），使前後端對接更順暢。

### 4. 持續整合與部署 (CI/CD)
雖然是範本，但我們堅持「開箱即用 DevOps」：
*   **多階段構建**: Dockerfile 採用 Multi-stage build，產出極小的 Alpine 映像檔。
*   **自動化流程**: GitHub Actions 配置了完整的 Lint -> Test -> Build 流程，並在 Tag 推送時自動發布 Docker Image。
*   **版本鎖定**: 嚴格鎖定 Go 版本 (1.24) 於 `go.mod`、`Dockerfile` 與 CI 配置中，避免環境不一致導致的構建失敗。