package tcpserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"iot-platform/internal/protocol/engine"
	"iot-platform/internal/protocol/model"
	"iot-platform/internal/sse"
	"iot-platform/pkg/config"
	"iot-platform/pkg/logger"
	redisClient "iot-platform/pkg/database/redis"
	"iot-platform/pkg/mq/kafka"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Server TCP 设备接入服务器
type Server struct {
	cfg       config.TCPConfig
	listener  net.Listener
	sessions  sync.Map            // deviceID -> *Session
	connCount atomic.Int64
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	producer  kafka.MessagePublisher
	sseHub    *sse.Hub
}

// Session 设备会话
type Session struct {
	ID          string
	DeviceID    string
	Protocol    string
	Conn        net.Conn
	ConnectedAt time.Time
	LastActive  time.Time
	RemoteAddr  string
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewServer 创建 TCP 服务器
func NewServer(cfg config.TCPConfig, producer kafka.MessagePublisher, sseHub *sse.Hub) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		producer: producer,
		sseHub:   sseHub,
	}
}

// Start 启动 TCP 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP server on %s: %w", addr, err)
	}
	s.listener = listener

	logger.Info("TCP server started",
		zap.String("addr", addr),
		zap.Int("max_connections", s.cfg.MaxConnections),
	)

	// 启动心跳检测协程
	s.wg.Add(1)
	go s.heartbeatChecker()

	// 接受连接循环
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					select {
					case <-s.ctx.Done():
						return
					default:
						logger.Error("Failed to accept connection", zap.Error(err))
						continue
					}
				}

				// 检查连接数
				if s.connCount.Load() >= int64(s.cfg.MaxConnections) {
					logger.Warn("Max connections reached, rejecting",
						zap.String("remote_addr", conn.RemoteAddr().String()),
					)
					conn.Close()
					continue
				}

				s.connCount.Add(1)
				s.wg.Add(1)
				go s.handleConnection(conn)
			}
		}
	}()

	return nil
}

// handleConnection 处理单个连接
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.connCount.Add(-1)
		s.wg.Done()
	}()

	sessionID := uuid.New().String()
	session := &Session{
		ID:          sessionID,
		Conn:        conn,
		ConnectedAt: time.Now(),
		LastActive:  time.Now(),
		RemoteAddr:  conn.RemoteAddr().String(),
	}
	session.ctx, session.cancel = context.WithCancel(s.ctx)

	logger.Info("New device connection",
		zap.String("session_id", sessionID),
		zap.String("remote_addr", session.RemoteAddr),
	)

	// 读取循环
	buf := make([]byte, 4096)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// 设置读超时
		conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))

		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			logger.Debug("Connection closed",
				zap.String("session_id", sessionID),
				zap.String("remote_addr", session.RemoteAddr),
				zap.Error(err),
			)
			break
		}

		if n > 0 {
			session.LastActive = time.Now()
			raw := make([]byte, n)
			copy(raw, buf[:n])

			// 处理数据
			s.processData(session, raw)
		}
	}

	// 设备断线处理
	s.onDisconnect(session)
}

// processData 处理设备上报数据
func (s *Server) processData(session *Session, raw []byte) {
	// 如果已识别协议，直接用已知协议解析
	if session.Protocol != "" {
		stdData, err := engine.Decode(session.Protocol, raw)
		if err != nil {
			logger.Error("Failed to decode device data",
				zap.String("protocol", session.Protocol),
				zap.String("device_id", session.DeviceID),
				zap.Error(err),
			)
			return
		}
		s.publishDeviceData(session, stdData)
		return
	}

	// 首次通信，自动检测协议
	adapter, err := engine.DetectProtocol(raw)
	if err != nil {
		logger.Error("Failed to detect protocol",
			zap.String("remote_addr", session.RemoteAddr),
			zap.Error(err),
		)
		return
	}

	session.Protocol = adapter.Name()
	logger.Info("Protocol detected",
		zap.String("session_id", session.ID),
		zap.String("protocol", adapter.Name()),
	)

	stdData, err := adapter.Decode(raw)
	if err != nil {
		logger.Error("Failed to decode data", zap.Error(err))
		return
	}

	session.DeviceID = stdData.DeviceID
	s.sessions.Store(stdData.DeviceID, session)

	// 缓存设备在线状态
	s.updateDeviceOnline(stdData.DeviceID, session)

	s.publishDeviceData(session, stdData)
}

// publishDeviceData 发布设备数据
func (s *Server) publishDeviceData(session *Session, stdData *model.StandardData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 写入 Redis 实时数据
	s.cacheRealtimeData(ctx, stdData)

	// 发送到 Kafka
	err := s.producer.Publish(ctx, "device.data.report", stdData.DeviceID, stdData)
	if err != nil {
		logger.Error("Failed to publish device data to kafka", zap.Error(err))
	}

	// 推送到 SSE 实时日志
	if s.sseHub != nil {
		entry := sse.FromStandardData(stdData, session.RemoteAddr)
		s.sseHub.Broadcast(entry)
	}

	logger.Debug("Device data processed",
		zap.String("device_id", stdData.DeviceID),
		zap.String("protocol", stdData.Protocol),
		zap.String("status", stdData.ChargingStatus),
	)
}

// cacheRealtimeData 缓存实时数据到 Redis
func (s *Server) cacheRealtimeData(ctx context.Context, data *model.StandardData) {
	key := fmt.Sprintf("device:realtime:%s", data.DeviceID)
	fields := map[string]interface{}{
		"voltage":           data.Voltage,
		"current":           data.Current,
		"power":             data.Power,
		"energy_total":      data.EnergyTotal,
		"energy_today":      data.EnergyToday,
		"temperature":       data.Temperature,
		"charging_status":   data.ChargingStatus,
		"updated_at":        data.Timestamp.Format(time.RFC3339),
	}

	pipe := redisClient.Client.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, 5*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Failed to cache realtime data", zap.String("device_id", data.DeviceID), zap.Error(err))
	}
}

// updateDeviceOnline 更新设备在线状态
func (s *Server) updateDeviceOnline(deviceID string, session *Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("device:online:%s", deviceID)
	redisClient.Client.Set(ctx, key, "1", 2*time.Minute)

	sessionKey := fmt.Sprintf("device:session:%s", deviceID)
	redisClient.Client.HSet(ctx, sessionKey, map[string]interface{}{
		"connector_id": session.ID,
		"connected_at": session.ConnectedAt.Format(time.RFC3339),
		"ip":           session.RemoteAddr,
		"protocol":     session.Protocol,
	})

	// 添加到在线设备集合
	redisClient.Client.SAdd(ctx, "devices:online", deviceID)

	// 按协议分组
	redisClient.Client.SAdd(ctx, fmt.Sprintf("protocol:online:%s", session.Protocol), deviceID)
}

// onDisconnect 设备断线处理
func (s *Server) onDisconnect(session *Session) {
	if session.DeviceID == "" {
		return
	}

	logger.Info("Device disconnected",
		zap.String("device_id", session.DeviceID),
		zap.String("protocol", session.Protocol),
	)

	s.sessions.Delete(session.DeviceID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 清除在线状态
	redisClient.Client.Del(ctx, fmt.Sprintf("device:online:%s", session.DeviceID))
	redisClient.Client.SRem(ctx, "devices:online", session.DeviceID)
	if session.Protocol != "" {
		redisClient.Client.SRem(ctx, fmt.Sprintf("protocol:online:%s", session.Protocol), session.DeviceID)
	}

	// 发布离线事件
	s.producer.Publish(ctx, "device.event.offline", session.DeviceID, map[string]interface{}{
		"device_id":      session.DeviceID,
		"protocol":       session.Protocol,
		"disconnected_at": time.Now(),
	})
}

// heartbeatChecker 心跳检测协程
func (s *Server) heartbeatChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			timeout := s.cfg.HeartbeatInterval * 3

			s.sessions.Range(func(key, value interface{}) bool {
				session := value.(*Session)
				if now.Sub(session.LastActive) > timeout {
					logger.Warn("Device heartbeat timeout",
						zap.String("device_id", session.DeviceID),
						zap.Duration("timeout", now.Sub(session.LastActive)),
					)
					session.cancel()
					session.Conn.Close()
				}
				return true
			})
		}
	}
}

// SendCommand 向指定设备发送指令
func (s *Server) SendCommand(deviceID string, data []byte) error {
	val, ok := s.sessions.Load(deviceID)
	if !ok {
		return fmt.Errorf("device not connected: %s", deviceID)
	}

	session := val.(*Session)
	session.Conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))

	_, err := session.Conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command to device %s: %w", deviceID, err)
	}

	logger.Info("Command sent to device",
		zap.String("device_id", deviceID),
		zap.Int("data_len", len(data)),
	)
	return nil
}

// Stop 停止 TCP 服务器
func (s *Server) Stop() {
	logger.Info("Stopping TCP server...")
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有会话
	s.sessions.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		session.Conn.Close()
		return true
	})

	s.wg.Wait()
	logger.Info("TCP server stopped")
}

// GetOnlineCount 获取在线设备数
func (s *Server) GetOnlineCount() int64 {
	return s.connCount.Load()
}

// GetSession 获取设备会话
func (s *Server) GetSession(deviceID string) (*Session, bool) {
	val, ok := s.sessions.Load(deviceID)
	if !ok {
		return nil, false
	}
	return val.(*Session), true
}
