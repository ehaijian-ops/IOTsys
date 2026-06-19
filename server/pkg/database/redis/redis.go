package redis

import (
	"context"
	"fmt"
	"time"

	"iot-platform/pkg/config"
	"iot-platform/pkg/logger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var Client *redis.Client

// Init 初始化 Redis 连接
func Init(cfg config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", cfg.Addr),
		zap.Int("db", cfg.DB),
	)
	return nil
}

// Close 关闭连接
func Close() {
	if Client != nil {
		Client.Close()
	}
}
