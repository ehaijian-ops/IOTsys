package system

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"iot-platform/internal/connector/tcpserver"
	"iot-platform/internal/sse"
	"iot-platform/pkg/config"
	"iot-platform/pkg/database/mongodb"
	"iot-platform/pkg/database/mysql"
	"iot-platform/pkg/database/redis"
	"iot-platform/pkg/mq/kafka"

	"github.com/gin-gonic/gin"
)

// StartTime 服务启动时间（由 main.go 设置）
var StartTime time.Time

// ServerInfo 服务器基本信息
type ServerInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	StartTime string `json:"start_time"`
	Env       string `json:"env"`
	Port      int    `json:"port"`
}

// ServiceStatus 单个服务状态
type ServiceStatus struct {
	Status  string `json:"status"` // running / stopped / disabled / error
	Details string `json:"details,omitempty"`
}

// ServicesInfo 关键服务状态
type ServicesInfo struct {
	MySQL     ServiceStatus `json:"mysql"`
	Redis     ServiceStatus `json:"redis"`
	MongoDB   ServiceStatus `json:"mongodb"`
	Kafka     ServiceStatus `json:"kafka"`
	TCPServer ServiceStatus `json:"tcp_server"`
	SSEHub    ServiceStatus `json:"sse_hub"`
}

// ResourceInfo 系统资源使用
type ResourceInfo struct {
	Goroutines   int    `json:"goroutines"`
	HeapAllocMB  string `json:"heap_alloc_mb"`
	NumCPU       int    `json:"num_cpu"`
	TCPConns     int64  `json:"tcp_connections"`
	SSEClients   int    `json:"sse_clients"`
}

// SystemStatus 完整系统状态
type SystemStatus struct {
	Server    ServerInfo   `json:"server"`
	Services  ServicesInfo `json:"services"`
	Resources ResourceInfo `json:"resources"`
	CheckedAt string       `json:"checked_at"`
}

// Handler 系统状态处理器
type Handler struct {
	cfg     *config.Config
	tcpSrv  *tcpserver.Server
	sseHub  *sse.Hub
	kafka   kafka.MessagePublisher
}

// NewHandler 创建系统状态处理器
func NewHandler(cfg *config.Config, tcpSrv *tcpserver.Server, sseHub *sse.Hub, producer kafka.MessagePublisher) *Handler {
	StartTime = time.Now()
	return &Handler{
		cfg:    cfg,
		tcpSrv: tcpSrv,
		sseHub: sseHub,
		kafka:  producer,
	}
}

// GetStatus 获取系统运行状态
func (h *Handler) GetStatus(c *gin.Context) {
	status := h.collectStatus()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": status,
	})
}

// collectStatus 收集各服务状态
func (h *Handler) collectStatus() *SystemStatus {
	now := time.Now()

	status := &SystemStatus{
		Server: ServerInfo{
			Name:      h.cfg.Server.Name,
			Version:   "1.0.0",
			Uptime:    formatDuration(now.Sub(StartTime)),
			StartTime: StartTime.Format("2006-01-02 15:04:05"),
			Env:       h.cfg.Server.Env,
			Port:      h.cfg.Server.Port,
		},
		Services: ServicesInfo{
			MySQL:     h.checkMySQL(),
			Redis:     h.checkRedis(),
			MongoDB:   h.checkMongoDB(),
			Kafka:     h.checkKafka(),
			TCPServer: h.checkTCPServer(),
			SSEHub:    h.checkSSEHub(),
		},
		Resources: ResourceInfo{
			Goroutines:  runtime.NumGoroutine(),
			HeapAllocMB: formatMB(runtime.MemStats{}),
			NumCPU:      runtime.NumCPU(),
			TCPConns:    h.tcpSrv.GetOnlineCount(),
			SSEClients:  h.sseHub.ClientCount(),
		},
		CheckedAt: now.Format("2006-01-02 15:04:05.000"),
	}

	// 正确获取内存状态
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	status.Resources.HeapAllocMB = fmt.Sprintf("%.2f", float64(mem.HeapAlloc)/1024/1024)

	return status
}

// checkMySQL 检查 MySQL 连接状态
func (h *Handler) checkMySQL() ServiceStatus {
	if mysql.DB == nil {
		return ServiceStatus{Status: "stopped", Details: "not initialized"}
	}
	sqlDB, err := mysql.DB.DB()
	if err != nil {
		return ServiceStatus{Status: "error", Details: err.Error()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return ServiceStatus{Status: "error", Details: err.Error()}
	}
	stats := sqlDB.Stats()
	return ServiceStatus{
		Status:  "running",
		Details: fmt.Sprintf("%s:%d, db=%s, open=%d idle=%d",
			h.cfg.MySQL.Host, h.cfg.MySQL.Port, h.cfg.MySQL.Database,
			stats.OpenConnections, stats.Idle),
	}
}

// checkRedis 检查 Redis 连接状态
func (h *Handler) checkRedis() ServiceStatus {
	if redis.Client == nil {
		return ServiceStatus{Status: "stopped", Details: "not initialized"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redis.Client.Ping(ctx).Err(); err != nil {
		return ServiceStatus{Status: "error", Details: err.Error()}
	}
	return ServiceStatus{
		Status:  "running",
		Details: fmt.Sprintf("%s, pool=%d", h.cfg.Redis.Addr, h.cfg.Redis.PoolSize),
	}
}

// checkMongoDB 检查 MongoDB 连接状态
func (h *Handler) checkMongoDB() ServiceStatus {
	if h.cfg.MongoDB.URI == "" {
		return ServiceStatus{Status: "disabled", Details: "not configured"}
	}
	if mongodb.Client == nil {
		return ServiceStatus{Status: "disabled", Details: "not initialized (degraded)"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mongodb.Client.Ping(ctx, nil); err != nil {
		return ServiceStatus{Status: "error", Details: err.Error()}
	}
	return ServiceStatus{
		Status:  "running",
		Details: fmt.Sprintf("db=%s", h.cfg.MongoDB.Database),
	}
}

// checkKafka 检查 Kafka 连接状态
func (h *Handler) checkKafka() ServiceStatus {
	if len(h.cfg.Kafka.Brokers) == 0 || h.cfg.Kafka.Brokers[0] == "" {
		return ServiceStatus{Status: "disabled", Details: "no brokers configured (using noop)"}
	}
	if _, ok := h.kafka.(*kafka.Producer); ok {
		return ServiceStatus{
			Status:  "running",
			Details: fmt.Sprintf("brokers=%v", h.cfg.Kafka.Brokers),
		}
	}
	return ServiceStatus{Status: "disabled", Details: "using noop producer"}
}

// checkTCPServer 检查 TCP 服务器状态
func (h *Handler) checkTCPServer() ServiceStatus {
	if !h.cfg.TCP.Enabled {
		return ServiceStatus{Status: "disabled", Details: "tcp.enabled=false"}
	}
	return ServiceStatus{
		Status:  "running",
		Details: fmt.Sprintf("port=%d, conns=%d, max=%d",
			h.cfg.TCP.Port, h.tcpSrv.GetOnlineCount(), h.cfg.TCP.MaxConnections),
	}
}

// checkSSEHub 检查 SSE Hub 状态
func (h *Handler) checkSSEHub() ServiceStatus {
	count := h.sseHub.ClientCount()
	return ServiceStatus{
		Status:  "running",
		Details: fmt.Sprintf("clients=%d", count),
	}
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}

// formatMB 格式化 MB
func formatMB(mem runtime.MemStats) string {
	return fmt.Sprintf("%.2f", float64(mem.HeapAlloc)/1024/1024)
}
