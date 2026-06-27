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
	siteHandler "iot-platform/internal/site/handler"
	"iot-platform/internal/sse"
	sysHandler "iot-platform/internal/system"
	userHandler "iot-platform/internal/user/handler"
	userRepo "iot-platform/internal/user/repository"
	userService "iot-platform/internal/user/service"

	// 新增模块
	adHandler "iot-platform/internal/advertisement/handler"
	adRepo "iot-platform/internal/advertisement/repository"
	adSvc "iot-platform/internal/advertisement/service"
	agentHandler "iot-platform/internal/agent/handler"
	agentRepo "iot-platform/internal/agent/repository"
	agentSvc "iot-platform/internal/agent/service"
	billingRepo "iot-platform/internal/billing/repository"
	billingSvc "iot-platform/internal/billing/service"
	billingHandler "iot-platform/internal/billing/handler"
	cardHandler "iot-platform/internal/card/handler"
	cardRepo "iot-platform/internal/card/repository"
	cardSvc "iot-platform/internal/card/service"
	financeHandler "iot-platform/internal/finance/handler"
	financeRepo "iot-platform/internal/finance/repository"
	financeSvc "iot-platform/internal/finance/service"
	icHandler "iot-platform/internal/interconnect/handler"
	icRepo "iot-platform/internal/interconnect/repository"
	icSvc "iot-platform/internal/interconnect/service"
	mtHandler "iot-platform/internal/maintenance/handler"
	mtRepo "iot-platform/internal/maintenance/repository"
	mtSvc "iot-platform/internal/maintenance/service"
	orderHandler "iot-platform/internal/order/handler"
	orderRepo "iot-platform/internal/order/repository"
	orderSvc "iot-platform/internal/order/service"
	sys2Handler "iot-platform/internal/system/handler"
	sys2Repo "iot-platform/internal/system/repository"
	sys2Svc "iot-platform/internal/system/service"

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
	_ "iot-platform/internal/protocol/adapters/wsd"

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

	// 设备UUID查找回调：协议层 DeviceID (SN) → 数据库 UUID
	// 用于 tcpserver 设置在线状态时同时写入 UUID key
	tcpSrv.SetDeviceUUIDLookup(func(protocolDeviceID string) string {
		dev, err := devRepo.GetBySN(context.Background(), protocolDeviceID)
		if err != nil {
			return ""
		}
		return dev.ID
	})

	// 未注册设备管理
	unregisteredH := deviceHandler.NewUnregisteredDeviceHandler(tcpSrv, devSvc, mysql.DB)

	// 站点管理
	siteH := siteHandler.NewSiteHandler(mysql.DB, devRepo)

	// 指令管理
	cmdSvc := cmdService.NewCommandService(mysql.DB, producer, tcpSrv)
	cmdH := cmdHandler.NewCommandHandler(cmdSvc)

	// 9. 初始化系统状态处理器
	sysH := sysHandler.NewHandler(cfg, tcpSrv, sseHub, producer)

	// 10. 初始化认证
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// 11. 初始化用户服务
	uRepo := userRepo.NewUserRepository(mysql.DB)
	uSvc := userService.NewUserService(uRepo, jwtManager)
	uH := userHandler.NewUserHandler(uSvc)

	// 12. 初始化订单服务
	ordRepo := orderRepo.NewOrderRepository(mysql.DB)
	recRepo := orderRepo.NewChargeRecordRepository(mysql.DB)
	refRepo := orderRepo.NewOrderRefundRepository(mysql.DB)
	ordSvc := orderSvc.NewOrderService(ordRepo, recRepo, refRepo)
	ordH := orderHandler.NewOrderHandler(ordSvc)

	// 13. 初始化收费方案服务
	billRepo := billingRepo.NewBillingRepository(mysql.DB)
	billSvc := billingSvc.NewBillingService(billRepo)
	billH := billingHandler.NewBillingHandler(billSvc)

	// 14. 初始化财务服务
	finRepo := financeRepo.NewFinanceRepository(mysql.DB)
	finSvc := financeSvc.NewFinanceService(finRepo)
	finH := financeHandler.NewFinanceHandler(finSvc)

	// 15. 初始化卡片服务
	cdRepo := cardRepo.NewCardRepository(mysql.DB)
	cdSvc := cardSvc.NewCardService(cdRepo)
	cdH := cardHandler.NewCardHandler(cdSvc)

	// 16. 初始化代理/运营商服务
	agRepo := agentRepo.NewAgentRepository(mysql.DB)
	agSvc := agentSvc.NewAgentService(agRepo)
	agH := agentHandler.NewAgentHandler(agSvc)

	// 17. 初始化运维服务
	mtR := mtRepo.NewMaintenanceRepository(mysql.DB)
	mtS := mtSvc.NewMaintenanceService(mtR)
	mtH := mtHandler.NewMaintenanceHandler(mtS)

	// 18. 初始化互联互通服务
	icR := icRepo.NewInterconnectRepository(mysql.DB)
	icS := icSvc.NewInterconnectService(icR)
	icH := icHandler.NewInterconnectHandler(icS)

	// 19. 初始化系统管理服务
	sys2R := sys2Repo.NewSystemRepository(mysql.DB)
	sys2S := sys2Svc.NewSystemService(sys2R)
	sys2H := sys2Handler.NewSystemHandler(sys2S)

	// 20. 初始化广告/运营服务
	adR := adRepo.NewAdvertisementRepository(mysql.DB)
	adS := adSvc.NewAdvertisementService(adR)
	adH := adHandler.NewAdvertisementHandler(adS)

	// 初始化默认管理员账号
	if err := uSvc.SeedDefaultAdmin(); err != nil {
		logger.Error("Failed to seed default admin", zap.Error(err))
	} else {
		logger.Info("Default admin user ready (username: admin)")
	}

	// 12. 创建 HTTP 路由
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
		v1.POST("/auth/login", uH.Login)

		// 系统状态监控
		v1.GET("/system/status", sysH.GetStatus)

		// 实时日志 SSE 流（公开，前端 EventSource 不支持自定义 Header）
		v1.GET("/devices/logs/stream", sseHub.HandleSSE)

		// 需要认证的接口
		authGroup := v1.Group("")
		authGroup.Use(auth.AuthMiddleware(jwtManager))
		{
			// 用户自己的操作
			authGroup.GET("/auth/userinfo", uH.GetUserInfo)
			authGroup.PUT("/auth/password", uH.ChangePassword)

			// 设备管理
			devices := authGroup.Group("/devices")
			{
				devices.GET("", devH.ListDevices)
				devices.POST("", auth.RequireRoles("admin", "super_admin"), devH.CreateDevice)
				devices.GET("/:id", devH.GetDevice)
				devices.PUT("/:id", auth.RequireRoles("admin", "super_admin"), devH.UpdateDevice)
				devices.DELETE("/:id", auth.RequireRoles("admin", "super_admin"), devH.DeleteDevice)

				// 未注册设备
				devices.GET("/unregistered", unregisteredH.ListUnregisteredDevices)
				devices.POST("/unregistered/add", auth.RequireRoles("admin", "super_admin"), unregisteredH.AddToSite)

				// 设备指令
				devices.POST("/:id/commands", cmdH.CreateCommand)
				devices.GET("/:id/commands", cmdH.ListCommands)
			}

			// 站点列表（简要，供下拉选择）
			authGroup.GET("/sites", unregisteredH.ListSites)

			// 站点管理（CRUD）
			sites := authGroup.Group("/sites/manage")
			{
				sites.GET("", siteH.ListSites)
				sites.GET("/:id", siteH.GetSite)
				sites.POST("", auth.RequireRoles("admin", "super_admin"), siteH.CreateSite)
				sites.PUT("/:id", auth.RequireRoles("admin", "super_admin"), siteH.UpdateSite)
				sites.DELETE("/:id", auth.RequireRoles("admin", "super_admin"), siteH.DeleteSite)
			}

			// 指令详情
			authGroup.GET("/commands/:id", cmdH.GetCommand)

			// ========== 用户管理（需要 admin 或 super_admin 权限） ==========
			users := authGroup.Group("/users")
			users.Use(auth.RequireRoles("admin", "super_admin"))
			{
				users.GET("", uH.ListUsers)
				users.GET("/roles", uH.GetRoles)
				users.POST("", uH.CreateUser)
				users.GET("/:id", uH.GetUser)
				users.PUT("/:id", uH.UpdateUser)
				users.DELETE("/:id", uH.DeleteUser)
				users.PUT("/:id/reset-password", uH.ResetPassword)
			}

			// ========== 财务管理 ==========
			finance := authGroup.Group("/finance")
			finance.Use(auth.RequireRoles("admin", "super_admin"))
			{
				finance.GET("/wallet", finH.GetWallet)
				finance.POST("/recharge", finH.AdminRecharge)
				finance.GET("/recharges", finH.ListRecharges)
				finance.GET("/withdraws", finH.ListWithdraws)
				finance.PUT("/withdraws/:id", finH.ProcessWithdraw)
				finance.GET("/splits", finH.ListSplits)
			}
			// 用户充值（普通认证用户可操作）
			authGroup.POST("/recharge", finH.Recharge)
			authGroup.POST("/withdraw", finH.ApplyWithdraw)

			// ========== 卡片管理 ==========
			cards := authGroup.Group("/cards")
			cards.Use(auth.RequireRoles("admin", "super_admin"))
			{
				// IC卡
				icCards := cards.Group("/ic")
				{
					icCards.GET("", cdH.ListICCards)
					icCards.POST("", cdH.CreateICCard)
					icCards.GET("/:id", cdH.GetICCard)
					icCards.POST("/:id/recharge", cdH.RechargeICCard)
					icCards.POST("/:id/bind", cdH.BindICCard)
					icCards.POST("/:id/lost", cdH.ReportLostICCard)
					icCards.DELETE("/:id", cdH.DeleteICCard)
					icCards.POST("/batch-import", cdH.BatchImportICCards)
				}
				// 流量卡
				trafficCards := cards.Group("/traffic")
				{
					trafficCards.GET("", cdH.ListTrafficCards)
					trafficCards.POST("", cdH.CreateTrafficCard)
					trafficCards.POST("/:id/bind", cdH.BindTrafficCard)
				}
				// 月卡
				monthlyCards := cards.Group("/monthly")
				{
					monthlyCards.GET("", cdH.ListMonthlyCards)
				}
			}

			// ========== 代理管理 ==========
			agents := authGroup.Group("/agents")
			agents.Use(auth.RequireRoles("admin", "super_admin"))
			{
				agents.GET("", agH.ListAgents)
				agents.POST("", agH.CreateAgent)
				agents.GET("/:id", agH.GetAgent)
				agents.PUT("/:id", agH.UpdateAgent)
				agents.DELETE("/:id", agH.DeleteAgent)
			}

			// ========== 运营商管理 ==========
			operators := authGroup.Group("/operators")
			operators.Use(auth.RequireRoles("admin", "super_admin"))
			{
				operators.GET("", agH.ListOperators)
				operators.POST("", agH.CreateOperator)
				operators.GET("/:id", agH.GetOperator)
				operators.PUT("/:id", agH.UpdateOperator)
				operators.DELETE("/:id", agH.DeleteOperator)
			}

			// ========== 运维管理 ==========
			maintenance := authGroup.Group("/maintenance")
			maintenance.Use(auth.RequireRoles("admin", "super_admin"))
			{
				faults := maintenance.Group("/faults")
				{
					faults.GET("", mtH.ListFaults)
					faults.POST("", mtH.CreateFault)
					faults.PUT("/:id", mtH.HandleFault)
				}
				tasks := maintenance.Group("/tasks")
				{
					tasks.GET("", mtH.ListTasks)
					tasks.POST("", mtH.CreateTask)
					tasks.PUT("/:id", mtH.UpdateTask)
					tasks.DELETE("/:id", mtH.DeleteTask)
					tasks.GET("/:id/logs", mtH.GetTaskLogs)
				}
				downloads := maintenance.Group("/downloads")
				{
					downloads.GET("", mtH.ListDownloads)
					downloads.POST("", mtH.CreateDownload)
				}
			}
			// 故障报修（普通用户可操作）
			authGroup.POST("/faults", mtH.CreateFault)

			// ========== 互联互通 ==========
			interconnect := authGroup.Group("/interconnect")
			interconnect.Use(auth.RequireRoles("admin", "super_admin"))
			{
				orgs := interconnect.Group("/orgs")
				{
					orgs.GET("", icH.ListOrgs)
					orgs.POST("", icH.CreateOrg)
					orgs.GET("/:id", icH.GetOrg)
					orgs.PUT("/:id", icH.UpdateOrg)
					orgs.DELETE("/:id", icH.DeleteOrg)
				}
				keys := interconnect.Group("/keys")
				{
					keys.GET("", icH.ListKeys)
					keys.POST("", icH.CreateKey)
					keys.DELETE("/:id", icH.DeleteKey)
				}
			}

			// ========== 系统管理 ==========
			system := authGroup.Group("/system")
			system.Use(auth.RequireRoles("admin", "super_admin"))
			{
				roles := system.Group("/roles")
				{
					roles.GET("", sys2H.ListRoles)
					roles.POST("", sys2H.CreateRole)
					roles.PUT("/:id", sys2H.UpdateRole)
					roles.DELETE("/:id", sys2H.DeleteRole)
				}
				menus := system.Group("/menus")
				{
					menus.GET("/tree", sys2H.GetMenuTree)
					menus.POST("", sys2H.CreateMenu)
					menus.PUT("/:id", sys2H.UpdateMenu)
					menus.DELETE("/:id", sys2H.DeleteMenu)
				}
				system.GET("/login-logs", sys2H.ListLoginLogs)
				system.GET("/operation-logs", sys2H.ListSystemLogs)
			}

			// ========== 广告/运营 ==========
			ads := authGroup.Group("/ads")
			ads.Use(auth.RequireRoles("admin", "super_admin"))
			{
				ads.GET("", adH.ListAds)
				ads.POST("", adH.CreateAd)
				ads.PUT("/:id", adH.UpdateAd)
				ads.DELETE("/:id", adH.DeleteAd)
			}
			franchises := authGroup.Group("/franchises")
			{
				franchises.GET("", auth.RequireRoles("admin", "super_admin"), adH.ListFranchises)
				franchises.POST("", adH.ApplyFranchise)
				franchises.PUT("/:id", auth.RequireRoles("admin", "super_admin"), adH.ProcessFranchise)
			}
			// 微信用户管理
			wechatUsers := authGroup.Group("/wechat-users")
			wechatUsers.Use(auth.RequireRoles("admin", "super_admin"))
			{
				wechatUsers.GET("", adH.ListWechatUsers)
				wechatUsers.PUT("/:id/freeze", adH.FreezeWechatUser)
			}

			// ========== 订单管理 ==========
			orders := authGroup.Group("/orders")
			{
				orders.GET("", ordH.ListOrders)
				orders.POST("", ordH.CreateOrder)
				orders.GET("/:id", ordH.GetOrder)
				orders.PUT("/:id/start", ordH.StartCharging)
				orders.PUT("/:id/end", ordH.EndOrder)
				orders.PUT("/:id/cancel", ordH.CancelOrder)
				orders.DELETE("/:id", auth.RequireRoles("admin", "super_admin"), ordH.DeleteOrder)
				orders.GET("/:id/curve", ordH.GetChargeCurve)
				orders.POST("/:id/refund", auth.RequireRoles("admin", "super_admin"), ordH.RefundOrder)
				orders.PUT("/:id/refund/:refund_id", auth.RequireRoles("admin", "super_admin"), ordH.ProcessRefund)
			}

			// ========== 收费方案管理 ==========
			billing := authGroup.Group("/billing")
			billing.Use(auth.RequireRoles("admin", "super_admin"))
			{
				// 收费方案
				schemes := billing.Group("/schemes")
				{
					schemes.GET("", billH.ListSchemes)
					schemes.POST("", billH.CreateScheme)
					schemes.GET("/:id", billH.GetScheme)
					schemes.PUT("/:id", billH.UpdateScheme)
					schemes.DELETE("/:id", billH.DeleteScheme)
					schemes.PUT("/:id/periods", billH.BatchSetPeriods)
					schemes.GET("/:id/periods", billH.GetPeriods)
				}
				// 月卡方案
				monthly := billing.Group("/monthly")
				{
					monthly.GET("", billH.ListMonthlySchemes)
					monthly.POST("", billH.CreateMonthlyScheme)
					monthly.PUT("/:id", billH.UpdateMonthlyScheme)
					monthly.DELETE("/:id", billH.DeleteMonthlyScheme)
				}
				// 充值方案
				recharges := billing.Group("/recharges")
				{
					recharges.GET("", billH.ListRechargeSchemes)
					recharges.POST("", billH.CreateRechargeScheme)
					recharges.PUT("/:id", billH.UpdateRechargeScheme)
					recharges.DELETE("/:id", billH.DeleteRechargeScheme)
				}
				// 业务配置
				configs := billing.Group("/configs")
				{
					configs.GET("", billH.ListConfigs)
					configs.GET("/:key", billH.GetConfig)
					configs.PUT("/:key", billH.SetConfig)
				}
			}
		}
	}

	// 13. 启动 HTTP 服务器
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
