package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance
var Logger *zap.Logger

type loggerKey struct{}

// NewLogger initializes a new Zap logger based on the provided log level and encoding.
func NewLogger(level, encoding string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel // Default to Info if parsing fails
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 for time formatting
	encoderConfig.CallerKey = "caller"                   // Add caller info
	encoderConfig.TimeKey = "time"                       // Add time key

	var encoder zapcore.Encoder
	switch encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig) // Default to JSON
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout), // Write to stdout
		zapLevel,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return Logger, nil
}

// FromContext returns a logger with fields from the context, or the global logger if none is found.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}
	if l, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return l
	}
	return Logger
}

// WithContext returns a new context with the provided logger.
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// TraceIDKey is the key used to store and retrieve trace ID in context and log fields.
const TraceIDKey = "traceID"

// WithTraceID adds a traceID field to the logger and returns a new context with this logger.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	l := FromContext(ctx).With(zap.String(TraceIDKey, traceID))
	return WithContext(ctx, l)
}
