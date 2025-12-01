package rocketmq

import (
	"context"

	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
	"go.uber.org/zap"
	
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// MQProducer 定義 RocketMQ Producer 的介面
type MQProducer interface {
	Start() error
	Shutdown() error
	SendSync(ctx context.Context, msg *primitive.Message) (*primitive.SendResult, error)
	Started() bool // 為健康檢查添加
}

// MQConsumer 定義 RocketMQ Consumer 的介面
type MQConsumer interface {
	Start() error
	Shutdown() error
}

// ProducerClient 是全域 RocketMQ Producer 客戶端
var ProducerClient MQProducer
// ConsumerClient 是全域 RocketMQ Consumer 客戶端
var ConsumerClient MQConsumer

// Stub 實作
type stubProducer struct {}
func (s *stubProducer) Start() error { return nil }
func (s *stubProducer) Shutdown() error { return nil }
func (s *stubProducer) SendSync(ctx context.Context, msg *primitive.Message) (*primitive.SendResult, error) { return nil, nil }
func (s *stubProducer) Started() bool { return true } // Stub 實作

type stubConsumer struct {}
func (s *stubConsumer) Start() error { return nil }
func (s *stubConsumer) Shutdown() error { return nil }

// InitProducer 初始化 RocketMQ Producer 客戶端
func InitProducer(cfg configs.RocketMQConfig) (MQProducer, error) {
	// STUBBED: 為了環境兼容性使用 Stub
	ProducerClient = &stubProducer{}
	logger.Logger.Info("RocketMQ Producer 初始化完成 (STUB)", zap.String("namesrv", cfg.NameSrvAddr))
	return ProducerClient, nil
}

// InitConsumer 初始化 RocketMQ Consumer 客戶端
func InitConsumer(cfg configs.RocketMQConfig, msgListener interface{}) (MQConsumer, error) {
	// STUBBED: 為了環境兼容性使用 Stub
	ConsumerClient = &stubConsumer{}
	logger.Logger.Info("RocketMQ Consumer 初始化完成 (STUB)", zap.String("namesrv", cfg.NameSrvAddr))
	return ConsumerClient, nil
}

// SendMessage 發送 RocketMQ 訊息
func SendMessage(ctx context.Context, topic string, payload []byte, keys []string) (*primitive.SendResult, error) {
	// STUBBED: 僅記錄日誌並返回成功
	logger.FromContext(ctx).Info("RocketMQ 訊息已發送 (STUB)",
		zap.String("topic", topic),
		zap.String("payload", string(payload)),
	)
	
	return &primitive.SendResult{
		Status: primitive.SendOK,
		MsgID:  "stub-msg-id",
	}, nil
}

// GracefulShutdown 關閉 RocketMQ 客戶端
func GracefulShutdown() {
	logger.Logger.Info("RocketMQ 客戶端已關閉 (STUB)")
}