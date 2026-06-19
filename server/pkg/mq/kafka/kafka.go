package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"iot-platform/pkg/config"
	"iot-platform/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer 生产者
type Producer struct {
	writer *kafka.Writer
}

// NewProducer 创建生产者
func NewProducer(cfg config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}

	return &Producer{writer: writer}
}

// Publish 发布消息
func (p *Producer) Publish(ctx context.Context, topic string, key string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	})
	if err != nil {
		logger.Error("Failed to publish kafka message",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() {
	if p.writer != nil {
		p.writer.Close()
	}
}

// Consumer 消费者
type Consumer struct {
	reader *kafka.Reader
}

// MessageHandler 消息处理器
type MessageHandler func(key, value []byte) error

// NewConsumer 创建消费者
func NewConsumer(cfg config.KafkaConfig, topic string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.ConsumerGroup,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{reader: reader}
}

// Consume 消费消息
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("Failed to read kafka message", zap.Error(err))
				continue
			}

			if err := handler(msg.Key, msg.Value); err != nil {
				logger.Error("Failed to handle message",
					zap.String("topic", msg.Topic),
					zap.ByteString("key", msg.Key),
					zap.Error(err),
				)
				// 继续处理下一条，不中断消费
			}
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() {
	if c.reader != nil {
		c.reader.Close()
	}
}
