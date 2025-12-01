package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	goRedis "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"microservice-mvp/pkg/configs"
	_ "microservice-mvp/docs" // Import generated Swagger docs
	"microservice-mvp/internal/controller"
	"microservice-mvp/internal/middleware"
	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/internal/service"
	"microservice-mvp/pkg/database"
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/redis"
)

// @title Microservice MVP API (Template)
// @version 1.0
// @description This is a lightweight Microservice Template suitable for quick start. It supports both In-Memory and MySQL modes.
// @contact.name API Support
// @license.name Apache 2.0
// @host localhost:8080
// @BasePath /
func main() {
	// 1. Load Configuration
	cfg, err := configs.LoadConfig("./configs/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	zapLogger, err := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Encoding)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger.Logger = zapLogger
	defer func() {
		_ = logger.Logger.Sync()
	}()

	logger.Logger.Info("Application starting up...", zap.String("persistence_mode", cfg.Persistence.Type))

	// 3. Initialize Persistence (Repository)
	var playerRepo repository.PlayerRepository
	var sqlDB *gorm.DB
	var redisClient *goRedis.Client

	switch cfg.Persistence.Type {
	case "mysql":
		// Init DB
		dbClient, err := database.InitTiDB(cfg.Database)
		if err != nil {
			logger.Logger.Fatal("Failed to initialize TiDB", zap.Error(err))
		}
		sqlDB = dbClient
		sqlDBGeneric, _ := sqlDB.DB()
		defer func() {
			_ = sqlDBGeneric.Close()
			logger.Logger.Info("Database connection closed.")
		}()

		// Auto-migrate
		err = dbClient.AutoMigrate(&model.Player{})
		if err != nil {
			logger.Logger.Fatal("Failed to auto-migrate database", zap.Error(err))
		}

		// Init Redis
		redisClient, err = redis.InitRedis(cfg.Redis)
		if err != nil {
			logger.Logger.Fatal("Failed to initialize Redis", zap.Error(err))
		}
		defer func() {
			_ = redisClient.Close()
			logger.Logger.Info("Redis client closed.")
		}()

		playerRepo = repository.NewPlayerRepositoryMySQL(sqlDB, redisClient)

	case "memory":
		logger.Logger.Info("Using In-Memory storage. Data will be lost on restart.")
		playerRepo = repository.NewPlayerRepositoryMemory()

	default:
		logger.Logger.Fatal("Invalid persistence type defined in config", zap.String("type", cfg.Persistence.Type))
	}

	// 4. Initialize Services
	authService := service.NewAuthService(playerRepo)
	playerService := service.NewPlayerService(playerRepo)

	// 5. Initialize Controllers
	// Note: HealthCheckController logic needs to be aware of enabled components.
	// For simplicity in this template, we keep it simple or would need refactoring to check only what's enabled.
	// Passing nil for now as placeholders if components are disabled.
	healthCheckController := controller.NewHealthCheckController(cfg) 
	authController := controller.NewAuthController(authService)
	playerController := controller.NewPlayerController(playerService)

	// 6. Setup Gin Engine
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()

	// Global Middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.TraceID())
	router.Use(middleware.LoggerMiddleware(cfg.Server))

	// Register Routes
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthCheckController.Check)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/login", authController.Login)
		v1.GET("/players/:id", playerController.GetPlayerInfo)
	}

	// 7. Start Server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Server startup failed", zap.Error(err))
		}
	}()

	logger.Logger.Info(fmt.Sprintf("Server is running on %s", serverAddr))

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server exited.")
}