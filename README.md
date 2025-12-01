# Go Microservice Template (MVP)

![CI Status](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)

A lightweight, production-ready Microservice Template designed for quick bootstrapping. 
It follows the **Standard Go Project Layout** and includes essential features like CI/CD, Swagger Documentation, and Auto Health Checks.

## 🚀 Key Features

*   **Dual Persistence Mode**: 
    *   **In-Memory**: Zero dependencies. Run `go run cmd/server/main.go` and it works immediately. Great for prototyping.
    *   **MySQL + Redis**: Production-ready mode with GORM and Redis caching. Configurable via `config.yaml`.
*   **Clean Architecture**: Decoupled layers (Controller -> Service -> Repository).
*   **Observability**: 
    *   Structured JSON Logging (Zap).
    *   TraceID Injection for request tracking.
    *   Health Check endpoint.
*   **API Documentation**: Auto-generated Swagger UI.
*   **DevOps Ready**:
    *   **Docker**: Optimized Multi-stage `Dockerfile`.
    *   **CI/CD**: GitHub Actions workflows for Linting, Testing, and Docker Building.

## 📂 Project Structure

```
.
├── cmd/
│   └── server/         # Entry point (main.go)
├── configs/            # Configuration files (config.yaml)
├── internal/
│   ├── controller/     # HTTP Handlers
│   ├── service/        # Business Logic
│   ├── repository/     # Data Access Layer (Interfaces & Impls)
│   └── model/          # Data Models
├── pkg/                # Public libraries (Logger, Response, etc.)
└── tests/              # Integration tests
```

## 🛠️ Getting Started

### 1. Run Locally (In-Memory Mode)

No Docker or Database required!

```bash
# Clone the repo
git clone https://github.com/your/repo.git
cd repo

# Run the server
go run cmd/server/main.go
```

Access the API:
*   **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
*   **Health Check**: [http://localhost:8080/health](http://localhost:8080/health)

### 2. Run with Database (MySQL + Redis)

1.  Modify `configs/config.yaml`:
    ```yaml
    persistence:
      type: "mysql" # Change from "memory" to "mysql"
    ```
2.  Start dependencies using Docker Compose:
    ```bash
    docker-compose -f deploy/docker-compose.yaml up -d
    ```
3.  Run the server:
    ```bash
    go run cmd/server/main.go
    ```

## 🧪 Testing

Run all unit and integration tests:

```bash
go test -v ./...
```

## 📝 API Documentation

To regenerate Swagger documentation after modifying controllers:

```bash
swag init -g cmd/server/main.go --parseDependency --parseInternal -o docs
```

## 🏗️ Design Decisions

1.  **Dependency Injection**: Services and Repositories are injected manually in `main.go`. This avoids complex DI frameworks and keeps the startup logic transparent.
2.  **Repository Pattern**: The `PlayerRepository` is an interface. This allows us to switch between `memory` and `mysql` implementations easily, facilitating testing and prototyping.
3.  **Configuration**: Using `Viper` for configuration management allows overriding settings via Environment Variables (e.g., `DATABASE_DSN` overrides `database.dsn`), which is crucial for Cloud Native deployments.
