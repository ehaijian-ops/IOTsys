package tcpserver

import (
	"context"
	"encoding/hex"
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
	cfg            config.TCPConfig
	listener       net.Listener
	sessions       sync.Map // deviceID -> *Session
	connCount      atomic.Int64
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	producer       kafka.MessagePublisher
	sseHub         *sse.Hub
	deviceUUIDLookup func(protocolDeviceID string) string // 协议层DeviceID → DB UUID 映射
}

// Session 设备会话
type Session struct {
	ID            string
	DeviceID      string
	Protocol      string
	SimCardNumber string    // 通信模块SIM卡号(20字节ASCII)，AP3000连接首报文
	PortCount     int       // 设备上报的端口/枪数量（AP3000=PortCount, TF100=GunCount）
	Conn          net.Conn
	ConnectedAt   time.Time
	LastActive    time.Time
	RemoteAddr    string
	WsdSessionID  []byte    // WSD协议登录时分配的SESSION_ID，心跳ACK回显使用
	ctx           context.Context
	cancel        context.CancelFunc
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

// SetDeviceUUIDLookup 设置协议层DeviceID → DB UUID 的查找回调
// 调用者（main）注入数据库查询，tcpserver 设置在线状态时同时写入两个 Redis key
func (s *Server) SetDeviceUUIDLookup(lookup func(protocolDeviceID string) string) {
	s.deviceUUIDLookup = lookup
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

	// 禁用 Nagle 算法，确保小包（如 10 字节心跳 ACK）立即发送而不被延迟合并
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

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

// AutoReplyer 协议适配器可选接口：自动回复设备查询类指令
type AutoReplyer interface {
	AutoReply(raw []byte, std *model.StandardData) []byte
}

// processData 处理设备上报数据
func (s *Server) processData(session *Session, raw []byte) {
	// 检测通信模块SIM卡号首报文（20字节, 0x38 0x39 0x38 0x36 = ASCII "8986"）
	// 模块每次连上socket第一时间发送，服务器无需应答
	if isSimCardNumber(raw) {
		simStr := string(raw)
		session.SimCardNumber = simStr
		logger.Info("SIM card number received",
			zap.String("session_id", session.ID),
			zap.String("remote_addr", session.RemoteAddr),
			zap.String("sim_card", simStr),
			zap.String("hex", hex.EncodeToString(raw)),
		)
		// 广播到 SSE 实时日志
		if s.sseHub != nil {
			entry := &sse.LogEntry{
				DeviceID:   simStr,
				Protocol:   "SIM_CARD",
				Timestamp:  time.Now().Format("2006-01-02 15:04:05.000"),
				RemoteAddr: session.RemoteAddr,
				RawHex:     hex.EncodeToString(raw),
				MsgType:    "sim_card",
				Type:       "sim_card",
				Direction:  "rx",
				Status:     simStr,
			}
			s.sseHub.Broadcast(entry)
		}
		return
	}

	// 如果已识别协议，直接用已知协议解析
	if session.Protocol != "" {
		stdData, err := engine.Decode(session.Protocol, raw)
		if err != nil {
			logger.Error("Failed to decode device data",
				zap.String("protocol", session.Protocol),
				zap.String("device_id", session.DeviceID),
				zap.Error(err),
			)
			// 解码失败时仍推送原始数据到 SSE
			s.broadcastRaw(session, raw, "decode_error")
			return
		}

		// 持续刷新在线状态 Redis TTL（首次 detection 后每次数据上报也需刷新）
		s.updateDeviceOnline(session.DeviceID, session)

		s.publishDeviceData(session, stdData, raw)

		// 从上报数据中更新端口/枪数量
		s.updatePortCount(session, stdData)

		// 尝试自动回复
		adapter, _ := engine.GetAdapter(session.Protocol)
		s.tryAutoReply(session, adapter, raw, stdData)
		return
	}

	// 首次通信，自动检测协议
	adapter, err := engine.DetectProtocol(raw)
	if err != nil {
		logger.Error("Failed to detect protocol",
			zap.String("remote_addr", session.RemoteAddr),
			zap.Int("data_len", len(raw)),
			zap.String("hex", hex.EncodeToString(raw[:minInt(len(raw), 40)])),
			zap.Error(err),
		)
		// 协议检测失败时推送原始数据到实时日志
		s.broadcastRaw(session, raw, "unknown_protocol")
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
		// 解码失败也推送原始数据
		s.broadcastRaw(session, raw, "decode_error")
		return
	}

	session.DeviceID = stdData.DeviceID
	s.sessions.Store(stdData.DeviceID, session)

	// 缓存设备在线状态
	s.updateDeviceOnline(stdData.DeviceID, session)

	s.publishDeviceData(session, stdData, raw)

	// 从上报数据中更新端口/枪数量
	s.updatePortCount(session, stdData)

	// 尝试自动回复
	s.tryAutoReply(session, adapter, raw, stdData)
}

// updatePortCount 从标准数据中提取端口/枪数量并更新到会话
func (s *Server) updatePortCount(session *Session, stdData *model.StandardData) {
	if stdData.PortCount > 0 {
		session.PortCount = stdData.PortCount
	} else if stdData.GunCount > 0 {
		session.PortCount = stdData.GunCount
	}
}

// tryAutoReply 如果协议适配器支持自动回复，则构建并发送回复
func (s *Server) tryAutoReply(session *Session, adapter engine.ProtocolAdapter, raw []byte, stdData *model.StandardData) {
	if adapter == nil {
		return
	}
	ar, ok := adapter.(AutoReplyer)
	if !ok {
		return
	}

	cmdStr, _ := stdData.Extra["cmd"].(string)
	// 提取原始帧的 SESSION_ID 用于诊断
	var sidHex string
	if len(raw) >= 9 {
		sidHex = hex.EncodeToString(raw[3:9])
	}

	// 传递登录时分配的 WSD SESSION_ID 给 AutoReply（心跳ACK需要回显此ID）
	if session.WsdSessionID != nil && len(session.WsdSessionID) == 6 {
		stdData.Extra["wsd_session_id"] = session.WsdSessionID
	}

	logger.Info("AutoReply triggered",
		zap.String("device_id", stdData.DeviceID),
		zap.String("protocol", adapter.Name()),
		zap.String("cmd", cmdStr),
		zap.Int("raw_len", len(raw)),
		zap.String("raw_hex", hex.EncodeToString(raw)),
		zap.String("raw_sid", sidHex),
		zap.String("remote_addr", session.RemoteAddr),
		zap.String("local_addr", session.Conn.LocalAddr().String()),
	)

	reply := ar.AutoReply(raw, stdData)
	if len(reply) == 0 {
		logger.Info("AutoReply returned empty",
			zap.String("device_id", stdData.DeviceID),
			zap.String("cmd", cmdStr),
		)
		return
	}

	logger.Info("AutoReply built response",
		zap.String("device_id", stdData.DeviceID),
		zap.String("cmd", cmdStr),
		zap.Int("reply_len", len(reply)),
		zap.String("reply_hex", hex.EncodeToString(reply)),
	)

	// 确保回复写入前连接状态日志
	session.Conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	n, writeErr := session.Conn.Write(reply)
	if writeErr != nil {
		logger.Error("Failed to send auto-reply",
			zap.String("device_id", stdData.DeviceID),
			zap.String("cmd", cmdStr),
			zap.String("remote_addr", session.RemoteAddr),
			zap.Error(writeErr),
		)
		return
	}

	// 登录应答后，提取分配给设备的 SESSION_ID 并存储，供后续心跳ACK回显使用
	if cmdStr == "login" && len(reply) >= 9 {
		sid := make([]byte, 6)
		copy(sid, reply[3:9])
		session.WsdSessionID = sid
		logger.Info("WSD session ID stored",
			zap.String("device_id", stdData.DeviceID),
			zap.String("session_id_hex", hex.EncodeToString(sid)),
		)
	}

	logger.Info("Auto-reply sent",
		zap.String("device_id", stdData.DeviceID),
		zap.String("cmd", cmdStr),
		zap.Int("len", len(reply)),
		zap.Int("written", n),
		zap.String("remote_addr", session.RemoteAddr),
	)

	// 将服务器回复广播到 SSE 实时日志供前端展示
	if s.sseHub != nil {
		deviceID := stdData.DeviceID
		if deviceID == "" {
			deviceID = session.DeviceID // 心跳等指令解析出的 DeviceID 可能为空，用 session 的
		}
		replyEntry := &sse.LogEntry{
			DeviceID:   deviceID,
			Protocol:   adapter.Name(),
			Timestamp:  time.Now().Format("2006-01-02 15:04:05.000"),
			RemoteAddr: session.RemoteAddr,
			RawHex:     hex.EncodeToString(reply),
			Type:       "reply",
			Direction:  "tx",
			Status:     cmdStr, // 标注回复类型: login / heartbeat / get_time
		}
		s.sseHub.Broadcast(replyEntry)
	}
}

// broadcastRaw 将未识别的原始数据广播到 SSE 实时日志
func (s *Server) broadcastRaw(session *Session, raw []byte, reason string) {
	if s.sseHub == nil {
		return
	}
	entry := &sse.LogEntry{
		DeviceID:   session.DeviceID,
		Protocol:   session.Protocol,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05.000"),
		RemoteAddr: session.RemoteAddr,
		RawHex:     hex.EncodeToString(raw),
		Type:       "raw",
		Status:     reason,
		Direction:  "rx",
	}
	s.sseHub.Broadcast(entry)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// publishDeviceData 发布设备数据
func (s *Server) publishDeviceData(session *Session, stdData *model.StandardData, raw []byte) {
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
		// 提取协议层报文类型
		msgType, _ := stdData.Extra["cmd"].(string)
		rawHex := hex.EncodeToString(raw)
		entry := sse.FromStandardData(stdData, session.RemoteAddr, rawHex, msgType)
		s.sseHub.Broadcast(entry)
	}

	logger.Debug("Device data processed",
		zap.String("device_id", stdData.DeviceID),
		zap.String("protocol", stdData.Protocol),
		zap.String("status", stdData.ChargingStatus),
	)
}

// cacheRealtimeData 缓存实时数据到 Redis（同时设置协议层ID和UUID两个key）
func (s *Server) cacheRealtimeData(ctx context.Context, data *model.StandardData) {
	// 格式化为字符串，避免 float64 精度问题（如 234.70000000000002）
	fields := map[string]interface{}{
		"voltage":           fmt.Sprintf("%.1f", data.Voltage),
		"current":           fmt.Sprintf("%.2f", data.Current),
		"power":             fmt.Sprintf("%.2f", data.Power),
		"energy_total":      fmt.Sprintf("%.2f", data.EnergyTotal),
		"energy_today":      fmt.Sprintf("%.2f", data.EnergyToday),
		"temperature":       fmt.Sprintf("%.1f", data.Temperature),
		"charging_status":   data.ChargingStatus,
		"updated_at":        data.Timestamp.Format(time.RFC3339),
	}

	// 以协议层 DeviceID 缓存
	protocolKey := fmt.Sprintf("device:realtime:%s", data.DeviceID)
	pipe := redisClient.Client.Pipeline()
	pipe.HSet(ctx, protocolKey, fields)
	pipe.Expire(ctx, protocolKey, 5*time.Minute)

	// 同步以 DB UUID 缓存（如果查询回调可用）
	if s.deviceUUIDLookup != nil {
		if dbUUID := s.deviceUUIDLookup(data.DeviceID); dbUUID != "" {
			uuidKey := fmt.Sprintf("device:realtime:%s", dbUUID)
			pipe.HSet(ctx, uuidKey, fields)
			pipe.Expire(ctx, uuidKey, 5*time.Minute)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Failed to cache realtime data", zap.String("device_id", data.DeviceID), zap.Error(err))
	}
}

// updateDeviceOnline 更新设备在线状态
// 同时设置两个 Redis key：
//   1. device:online:{protocolDeviceID}  （协议层ID）
//   2. device:online:{dbUUID}             （数据库UUID，如果查找回调可用）
func (s *Server) updateDeviceOnline(deviceID string, session *Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 设置协议层 DeviceID 的在线标记
	key := fmt.Sprintf("device:online:%s", deviceID)
	redisClient.Client.Set(ctx, key, "1", 2*time.Minute)

	// 查找 DB UUID 并同步设置（解决协议层ID与数据库UUID不一致的问题）
	if s.deviceUUIDLookup != nil {
		if dbUUID := s.deviceUUIDLookup(deviceID); dbUUID != "" {
			uuidKey := fmt.Sprintf("device:online:%s", dbUUID)
			redisClient.Client.Set(ctx, uuidKey, "1", 2*time.Minute)
			redisClient.Client.SAdd(ctx, "devices:online", dbUUID)
		}
	}

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

	// 清除在线状态（协议层 DeviceID）
	redisClient.Client.Del(ctx, fmt.Sprintf("device:online:%s", session.DeviceID))
	redisClient.Client.SRem(ctx, "devices:online", session.DeviceID)

	// 同步清除 DB UUID 对应的在线标记
	if s.deviceUUIDLookup != nil {
		if dbUUID := s.deviceUUIDLookup(session.DeviceID); dbUUID != "" {
			redisClient.Client.Del(ctx, fmt.Sprintf("device:online:%s", dbUUID))
			redisClient.Client.SRem(ctx, "devices:online", dbUUID)
		}
	}

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

// SessionBrief 会话摘要（供外部API使用，不暴露net.Conn等敏感字段）
type SessionBrief struct {
	DeviceID      string    `json:"device_id"`
	Protocol      string    `json:"protocol"`
	SimCardNumber string    `json:"sim_card_number,omitempty"`
	PortCount     int       `json:"port_count"`
	RemoteAddr    string    `json:"remote_addr"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastActive    time.Time `json:"last_active"`
}

// GetOnlineSessionBriefs 获取所有在线设备的会话摘要
func (s *Server) GetOnlineSessionBriefs() []SessionBrief {
	var briefs []SessionBrief
	s.sessions.Range(func(key, value interface{}) bool {
		sess := value.(*Session)
		if sess.DeviceID != "" {
			briefs = append(briefs, SessionBrief{
				DeviceID:      sess.DeviceID,
				Protocol:      sess.Protocol,
				SimCardNumber: sess.SimCardNumber,
				PortCount:     sess.PortCount,
				RemoteAddr:    sess.RemoteAddr,
				ConnectedAt:   sess.ConnectedAt,
				LastActive:    sess.LastActive,
			})
		}
		return true
	})
	return briefs
}

// isSimCardNumber 检测是否为通信模块SIM卡号报文
// 格式: 20字节ASCII字符串，固定以0x38 0x39 0x38 0x36 ("8986")开头
func isSimCardNumber(raw []byte) bool {
	if len(raw) != 20 {
		return false
	}
	return raw[0] == 0x38 && raw[1] == 0x39 && raw[2] == 0x38 && raw[3] == 0x36
}
