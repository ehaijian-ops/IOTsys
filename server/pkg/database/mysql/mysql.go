package mysql

import (
	"fmt"
	"time"

	"iot-platform/pkg/config"
	"iot-platform/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化 MySQL 连接
func Init(cfg config.MySQLConfig) error {
	logLevel := gormlogger.Info
	switch cfg.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 连接池测试
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	DB = db
	logger.Info("MySQL connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)
	return nil
}

// Close 关闭连接
func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
