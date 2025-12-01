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
	_ "microservice-mvp/docs" // 匯入生成的 Swagger 文件
	"microservice-mvp/internal/controller"
	"microservice-mvp/internal/middleware"
	"microservice-mvp/internal/model"
	"microservice-mvp/internal/repository"
	"microservice-mvp/internal/service"
	"microservice-mvp/pkg/database"
	"microservice-mvp/pkg/logger"
	"microservice-mvp/pkg/redis"
)

// @title Microservice MVP API (範本)
// @version 1.0
// @description 這是一個輕量級的微服務範本，適合快速啟動專案。它支援 In-Memory 和 MySQL 兩種模式。
// @contact.name API Support
// @license.name Apache 2.0
// @host localhost:8080
// @BasePath /
func main() {
	// 1. 載入配置
	cfg, err := configs.LoadConfig("./configs/config.yaml")
	if err != nil {
		fmt.Printf("載入配置失敗: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日誌
	zapLogger, err := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Encoding)
	if err != nil {
		fmt.Printf("初始化日誌失敗: %v\n", err)
		os.Exit(1)
	}
	logger.Logger = zapLogger
	defer func() {
		_ = logger.Logger.Sync()
	}()

	logger.Logger.Info("應用程式啟動中...", zap.String("persistence_mode", cfg.Persistence.Type))

	// 3. 初始化持久化層 (Repository)
	var playerRepo repository.PlayerRepository
	var sqlDB *gorm.DB
	var redisClient *goRedis.Client

	switch cfg.Persistence.Type {
	case "mysql":
		// 初始化資料庫
		dbClient, err := database.InitTiDB(cfg.Database)
		if err != nil {
			logger.Logger.Fatal("初始化 TiDB 失敗", zap.Error(err))
		}
		sqlDB = dbClient
		sqlDBGeneric, _ := sqlDB.DB()
		defer func() {
			_ = sqlDBGeneric.Close()
			logger.Logger.Info("資料庫連線已關閉")
		}()

		// 自動遷移 (Auto-migrate)
		err = dbClient.AutoMigrate(&model.Player{})
		if err != nil {
			logger.Logger.Fatal("資料庫自動遷移失敗", zap.Error(err))
		}

		// 初始化 Redis
		redisClient, err = redis.InitRedis(cfg.Redis)
		if err != nil {
			logger.Logger.Fatal("初始化 Redis 失敗", zap.Error(err))
		}
		defer func() {
			_ = redisClient.Close()
			logger.Logger.Info("Redis 客戶端已關閉")
		}()

		playerRepo = repository.NewPlayerRepositoryMySQL(sqlDB, redisClient)

	case "memory":
		logger.Logger.Info("使用 In-Memory 儲存模式。重啟後資料將會遺失。 সন")
		playerRepo = repository.NewPlayerRepositoryMemory()

	default:
		logger.Logger.Fatal("配置中定義了無效的持久化類型", zap.String("type", cfg.Persistence.Type))
	}

	// 4. 初始化服務層 (Services)
	authService := service.NewAuthService(playerRepo)
	playerService := service.NewPlayerService(playerRepo)

	// 5. 初始化控制器 (Controllers)
	// 注意: HealthCheckController 的邏輯需要感知已啟用的組件。
	// 在此範本中，我們保持簡單，若組件未啟用則傳遞 nil。
	healthCheckController := controller.NewHealthCheckController(cfg) 
	authController := controller.NewAuthController(authService)
	playerController := controller.NewPlayerController(playerService)

	// 6. 設定 Gin 引擎與路由
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()

	// 全域中間件 (Middleware)
	router.Use(middleware.Recovery())
	router.Use(middleware.TraceID())
	router.Use(middleware.LoggerMiddleware(cfg.Server))

	// 註冊路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthCheckController.Check)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/login", authController.Login)
		v1.GET("/players/:id", playerController.GetPlayerInfo)
	}

	// 7. 啟動伺服器
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("伺服器啟動失敗", zap.Error(err))
		}
	}()

	logger.Logger.Info(fmt.Sprintf("伺服器運行於 %s", serverAddr))

	// 8. 優雅關閉 (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("正在關閉伺服器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("伺服器強制關閉", zap.Error(err))
	}

	logger.Logger.Info("伺服器已退出")
}
