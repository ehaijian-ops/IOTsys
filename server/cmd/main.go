package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot-platform/internal/connector/tcpserver"
	cmdHandler "iot-platform/internal/command/handler"
	cmdService "iot-platform/internal/command/service"
	deviceHandler "iot-platform/internal/device/handler"
	deviceRepo "iot-platform/internal/device/repository"
	deviceService "iot-platform/internal/device/service"
	"iot-platform/internal/sse"
	sysHandler "iot-platform/internal/system"
	"iot-platform/migrations"
	"iot-platform/pkg/auth"
	"iot-platform/pkg/config"
	"iot-platform/pkg/database/mongodb"
	"iot-platform/pkg/database/mysql"
	"iot-platform/pkg/database/redis"
	"iot-platform/pkg/logger"
	"iot-platform/pkg/middleware"
	"iot-platform/pkg/mq/kafka"

	// 导入协议适配器（触发 init 注册）
	_ "iot-platform/internal/protocol/adapters/ap3000"
	_ "iot-platform/internal/protocol/adapters/tf100"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting IoT Platform...",
		zap.String("env", cfg.Server.Env),
		zap.Int("port", cfg.Server.Port),
	)

	// 3. 初始化数据库
	if err := mysql.Init(cfg.MySQL); err != nil {
		logger.Fatal("Failed to init MySQL", zap.Error(err))
	}
	defer mysql.Close()

	if err := redis.Init(cfg.Redis); err != nil {
		logger.Fatal("Failed to init Redis", zap.Error(err))
	}
	defer redis.Close()

	// MongoDB - 可选，失败时降级运行
	if cfg.MongoDB.URI != "" {
		if err := mongodb.Init(cfg.MongoDB); err != nil {
			logger.Warn("MongoDB init failed, running without MongoDB", zap.Error(err))
		} else {
			defer mongodb.Close()
		}
	} else {
		logger.Info("MongoDB not configured, skipping")
	}

	// 4. 数据库迁移
	if err := migrations.AutoMigrate(mysql.DB); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 5. 初始化消息队列 - Kafka 不可用时使用 NoopProducer 降级
	var producer kafka.MessagePublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		producer = kafka.NewProducer(cfg.Kafka)
		defer producer.Close()
	} else {
		producer = kafka.NewNoopProducer()
		defer producer.Close()
	}

	// 6. 初始化 SSE Hub（实时日志推送）
	sseHub := sse.NewHub()

	// 7. 初始化 TCP 设备接入服务器
	tcpSrv := tcpserver.NewServer(cfg.TCP, producer, sseHub)
	if cfg.TCP.Enabled {
		if err := tcpSrv.Start(); err != nil {
			logger.Fatal("Failed to start TCP server", zap.Error(err))
		}
		defer tcpSrv.Stop()
	}

	// 8. 初始化业务服务
	// 设备管理
	devRepo := deviceRepo.NewDeviceRepository(mysql.DB)
	devSvc := deviceService.NewDeviceService(devRepo)
	devH := deviceHandler.NewDeviceHandler(devSvc)

	// 指令管理
	cmdSvc := cmdService.NewCommandService(mysql.DB, producer, tcpSrv)
	cmdH := cmdHandler.NewCommandHandler(cmdSvc)

	// 9. 初始化系统状态处理器
	sysH := sysHandler.NewHandler(cfg, tcpSrv, sseHub, producer)

	// 10. 初始化认证
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// 10. 创建 HTTP 路由
	router := gin.New()

	// 中间件
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	// 健康检查
	router.GET("/health", auth.HealthCheck)
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"tcp_connections": tcpSrv.GetOnlineCount(),
		})
	})

	// API v1 路由
	v1 := router.Group("/api/v1")
	{
		// 公开接口
		v1.POST("/auth/login", func(c *gin.Context) {
			// TODO: 实现登录逻辑
			c.JSON(http.StatusOK, gin.H{"message": "login endpoint"})
		})

		// 系统状态监控
		v1.GET("/system/status", sysH.GetStatus)

		// 实时日志 SSE 流（公开，前端 EventSource 不支持自定义 Header）
		v1.GET("/devices/logs/stream", sseHub.HandleSSE)

		// 需要认证的接口
		authGroup := v1.Group("")
		authGroup.Use(auth.AuthMiddleware(jwtManager))
		{
			// 设备管理
			devices := authGroup.Group("/devices")
			{
				devices.GET("", devH.ListDevices)
				devices.POST("", auth.RequireRoles("admin"), devH.CreateDevice)
				devices.GET("/:id", devH.GetDevice)
				devices.PUT("/:id", auth.RequireRoles("admin"), devH.UpdateDevice)
				devices.DELETE("/:id", auth.RequireRoles("admin"), devH.DeleteDevice)

				// 设备指令
				devices.POST("/:id/commands", cmdH.CreateCommand)
				devices.GET("/:id/commands", cmdH.ListCommands)
			}

			// 指令详情
			authGroup.GET("/commands/:id", cmdH.GetCommand)
		}
	}

	// 11. 启动 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 优雅关闭
	go func() {
		logger.Info("HTTP server started", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("Shutting down server...", zap.String("signal", sig.String()))

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
