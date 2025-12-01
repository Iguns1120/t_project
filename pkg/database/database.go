package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"microservice-mvp/pkg/configs"
	pkgLogger "microservice-mvp/pkg/logger" // Alias to avoid conflict with gorm.io/gorm/logger
	"go.uber.org/zap"
)

// DB is the global GORM DB client
var DB *gorm.DB

// InitTiDB initializes the TiDB connection using GORM.
func InitTiDB(cfg configs.DatabaseConfig) (*gorm.DB, error) {
	newLogger := logger.New(
		&logWriter{}, // Custom log writer to integrate with Zap
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Warn, // Log level
			Colorful:      false,       // Disable color
		},
	)

	dsn := cfg.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TiDB: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)

	DB = db // Set global DB instance
	pkgLogger.Logger.Info("TiDB connection initialized successfully")
	return db, nil
}

// GetDB returns the global GORM DB client.
func GetDB() *gorm.DB {
	return DB
}

// WithContext passes the logger from the context to GORM's session.
// This allows GORM logs to include the traceID.
func WithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return DB
	}
	// Retrieve logger from context, which should contain traceID
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

// logWriter is a custom writer to integrate GORM logs with Zap.
type logWriter struct {
	zapLogger *zap.Logger
}

func (l *logWriter) Printf(format string, v ...interface{}) {
	if l.zapLogger == nil {
		l.zapLogger = pkgLogger.Logger // Fallback to global logger if context logger is not set
	}
	// GORM's Printf often includes the log level in the format itself.
	// We'll parse it or just log it at a default level.
	// For simplicity, we'll log most GORM messages at Debug level,
	// and rely on GORM's LogLevel config to filter.
	// Slow queries will already be logged by GORM at Warn level.
	l.zapLogger.Debug(fmt.Sprintf(format, v...))
}
