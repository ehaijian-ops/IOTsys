package kafka

import (
	"context"
	"iot-platform/pkg/logger"
	"go.uber.org/zap"
)

// NoopProducer 空操作生产者，用于 Kafka 不可用时的降级方案
type NoopProducer struct{}

// NewNoopProducer 创建空操作生产者
func NewNoopProducer() *NoopProducer {
	logger.Warn("Kafka not available, using noop producer - messages will not be sent")
	return &NoopProducer{}
}

// Publish 空发布（仅记录日志）
func (np *NoopProducer) Publish(ctx context.Context, topic string, key string, msg interface{}) error {
	logger.Debug("NoopProducer: message dropped",
		zap.String("topic", topic),
		zap.String("key", key),
	)
	return nil
}

// Close 关闭（空操作）
func (np *NoopProducer) Close() {}

// MessagePublisher 消息发布接口
type MessagePublisher interface {
	Publish(ctx context.Context, topic string, key string, msg interface{}) error
	Close()
}

// 确保 Producer 和 NoopProducer 都实现了 MessagePublisher
var _ MessagePublisher = (*Producer)(nil)
var _ MessagePublisher = (*NoopProducer)(nil)
