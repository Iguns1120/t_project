package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"microservice-mvp/pkg/configs"
	pkgLogger "microservice-mvp/pkg/logger" // 別名以避免與 gorm.io/gorm/logger 衝突
	"go.uber.org/zap"
)

// DB 是全域 GORM DB 客戶端
var DB *gorm.DB

// InitTiDB 使用 GORM 初始化 TiDB 連線
func InitTiDB(cfg configs.DatabaseConfig) (*gorm.DB, error) {
	newLogger := logger.New(
		&logWriter{}, // 自定義日誌寫入器以整合 Zap
		logger.Config{
			SlowThreshold: time.Second, // 慢查詢閾值
			LogLevel:      logger.Warn, // 日誌級別
			Colorful:      false,       // 禁用顏色
		},
	)

	dsn := cfg.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("連線到 TiDB 失敗: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("獲取底層 sql.DB 失敗: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)

	DB = db // 設定全域 DB 實例
	pkgLogger.Logger.Info("TiDB 連線初始化成功")
	return db, nil
}

// GetDB 回傳全域 GORM DB 客戶端
func GetDB() *gorm.DB {
	return DB
}

// WithContext 將上下文中的日誌器傳遞給 GORM 的 session
// 這允許 GORM 日誌包含 traceID
func WithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return DB
	}
	// 從上下文中獲取 logger (應包含 traceID)
	zapLogger := pkgLogger.FromContext(ctx)
	return DB.WithContext(ctx).Session(&gorm.Session{
		Logger: logger.New(
			&logWriter{zapLogger: zapLogger},
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Warn,
				Colorful:      false,
			},
		),
	})
}

// logWriter 是自定義寫入器，用於將 GORM 日誌整合到 Zap
type logWriter struct {
	zapLogger *zap.Logger
}

func (l *logWriter) Printf(format string, v ...interface{}) {
	if l.zapLogger == nil {
		l.zapLogger = pkgLogger.Logger // 如果上下文 logger 未設定，則回退到全域 logger
	}
	// GORM 的 Printf 通常在格式字串中包含日誌級別
	// 我們將其解析或僅以預設級別記錄
	// 為簡單起見，我們在 Debug 級別記錄大多數 GORM 訊息，
	// 並依賴 GORM 的 LogLevel 配置進行過濾
	// 慢查詢已由 GORM 在 Warn 級別記錄
	l.zapLogger.Debug(fmt.Sprintf(format, v...))
}