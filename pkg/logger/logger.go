package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是全域 logger 實例
var Logger *zap.Logger

type loggerKey struct{}

// NewLogger 根據提供的日誌級別和編碼初始化新的 Zap logger
func NewLogger(level, encoding string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel // 如果解析失敗，預設為 Info
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 使用 ISO8601 格式化時間
	encoderConfig.CallerKey = "caller"                   // 添加呼叫者資訊
	encoderConfig.TimeKey = "time"                       // 添加時間鍵

	var encoder zapcore.Encoder
	switch encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig) // 預設為 JSON
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout), // 寫入到標準輸出
		zapLevel,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return Logger, nil
}

// FromContext 從上下文中返回帶有欄位的 logger，如果未找到則返回全域 logger
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}
	if l, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return l
	}
	return Logger
}

// WithContext 返回一個帶有提供的 logger 的新上下文
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// TraceIDKey 是用於在上下文和日誌欄位中存儲和檢索 trace ID 的鍵
const TraceIDKey = "traceID"

// WithTraceID 將 traceID 欄位添加到 logger 並返回帶有此 logger 的新上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	l := FromContext(ctx).With(zap.String(TraceIDKey, traceID))
	return WithContext(ctx, l)
}