package rocketmq

import (
	"context"

	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
	"go.uber.org/zap"
	
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// MQProducer defines the interface for RocketMQ Producer.
type MQProducer interface {
	Start() error
	Shutdown() error
	SendSync(ctx context.Context, msg *primitive.Message) (*primitive.SendResult, error)
	Started() bool // Added for health check
}

// MQConsumer defines the interface for RocketMQ Consumer.
type MQConsumer interface {
	Start() error
	Shutdown() error
}

// ProducerClient is the global RocketMQ Producer client
var ProducerClient MQProducer
// ConsumerClient is the global RocketMQ Consumer client
var ConsumerClient MQConsumer

// Stub implementation
type stubProducer struct {}
func (s *stubProducer) Start() error { return nil }
func (s *stubProducer) Shutdown() error { return nil }
func (s *stubProducer) SendSync(ctx context.Context, msg *primitive.Message) (*primitive.SendResult, error) { return nil, nil }
func (s *stubProducer) Started() bool { return true } // Stub implementation

type stubConsumer struct {}
func (s *stubConsumer) Start() error { return nil }
func (s *stubConsumer) Shutdown() error { return nil }

// InitProducer initializes the RocketMQ Producer client.
func InitProducer(cfg configs.RocketMQConfig) (MQProducer, error) {
	// STUBBED for environment compatibility
	ProducerClient = &stubProducer{}
	logger.Logger.Info("RocketMQ Producer initialized (STUB)", zap.String("namesrv", cfg.NameSrvAddr))
	return ProducerClient, nil
}

// InitConsumer initializes the RocketMQ Consumer client.
func InitConsumer(cfg configs.RocketMQConfig, msgListener interface{}) (MQConsumer, error) {
	// STUBBED for environment compatibility
	ConsumerClient = &stubConsumer{}
	logger.Logger.Info("RocketMQ Consumer initialized (STUB)", zap.String("namesrv", cfg.NameSrvAddr))
	return ConsumerClient, nil
}

// SendMessage sends a RocketMQ message.
func SendMessage(ctx context.Context, topic string, payload []byte, keys []string) (*primitive.SendResult, error) {
	// STUBBED: Just log and return success
	logger.FromContext(ctx).Info("RocketMQ message sent (STUB)",
		zap.String("topic", topic),
		zap.String("payload", string(payload)),
	)
	
	return &primitive.SendResult{
		Status: primitive.SendOK,
		MsgID:  "stub-msg-id",
	}, nil
}

// GracefulShutdown closes the RocketMQ clients.
func GracefulShutdown() {
	logger.Logger.Info("RocketMQ clients shut down (STUB).")
}
