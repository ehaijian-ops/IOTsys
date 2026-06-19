package mongodb

import (
	"context"
	"fmt"
	"time"

	"iot-platform/pkg/config"
	"iot-platform/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var Client *mongo.Client
var DB *mongo.Database

// Init 初始化 MongoDB 连接
func Init(cfg config.MongoDBConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect mongodb: %w", err)
	}

	// 验证连接
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping mongodb: %w", err)
	}

	Client = client
	DB = client.Database(cfg.Database)

	logger.Info("MongoDB connected successfully",
		zap.String("database", cfg.Database),
	)
	return nil
}

// Close 关闭连接
func Close() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		Client.Disconnect(ctx)
	}
}
