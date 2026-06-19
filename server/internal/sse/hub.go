package sse

import (
	"encoding/json"
	"sync"

	"iot-platform/internal/protocol/model"
	"iot-platform/pkg/logger"

	"go.uber.org/zap"
)

// Client 代表一个 SSE 订阅连接
type Client struct {
	ID       string
	Events   chan *LogEntry
	Done     chan struct{}
	Filter   ClientFilter
}

// ClientFilter 客户端过滤条件
type ClientFilter struct {
	DeviceID string // 按设备ID过滤（空=不过滤）
	Protocol string // 按协议过滤（空=不过滤）
	Keyword  string // 关键字搜索（匹配消息内容）
}

// LogEntry 单条日志记录
type LogEntry struct {
	DeviceID    string  `json:"device_id"`
	Protocol    string  `json:"protocol"`
	Timestamp   string  `json:"timestamp"`
	MsgID       int     `json:"msg_id"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Power       float64 `json:"power"`
	EnergyTotal float64 `json:"energy_total"`
	Status      string  `json:"charging_status"`
	Temperature float64 `json:"temperature"`
	FaultCode   string  `json:"fault_code,omitempty"`
	RemoteAddr  string  `json:"remote_addr,omitempty"`
	RawHex      string  `json:"raw_hex,omitempty"`
	Type        string  `json:"type"` // "data" / "connect" / "disconnect"
}

// Hub SSE 消息分发中心
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	broadcast  chan *LogEntry
	register   chan *Client
	unregister chan *Client
	stopped    chan struct{}
}

// NewHub 创建新的 SSE Hub
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *LogEntry, 1024),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stopped:    make(chan struct{}),
	}
	go h.run()
	return h
}

// run Hub 主循环
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			logger.Debug("SSE client connected",
				zap.String("client_id", client.ID),
				zap.Int("total", len(h.clients)),
			)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Events)
			}
			h.mu.Unlock()
			logger.Debug("SSE client disconnected",
				zap.String("client_id", client.ID),
				zap.Int("total", len(h.clients)),
			)

		case entry := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if h.matchFilter(client.Filter, entry) {
					select {
					case client.Events <- entry:
					default:
						// 客户端消费太慢，跳过
					}
				}
			}
			h.mu.RUnlock()

		case <-h.stopped:
			h.mu.Lock()
			for _, client := range h.clients {
				close(client.Events)
			}
			h.clients = make(map[string]*Client)
			h.mu.Unlock()
			return
		}
	}
}

// matchFilter 检查日志是否匹配客户端过滤条件
func (h *Hub) matchFilter(f ClientFilter, entry *LogEntry) bool {
	if f.DeviceID != "" && entry.DeviceID != f.DeviceID {
		return false
	}
	if f.Protocol != "" && entry.Protocol != f.Protocol {
		return false
	}
	if f.Keyword != "" {
		data, _ := json.Marshal(entry)
		if !contains(string(data), f.Keyword) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Register 注册 SSE 客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销 SSE 客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast 广播设备日志
func (h *Hub) Broadcast(entry *LogEntry) {
	select {
	case h.broadcast <- entry:
	default:
		// 广播队列满时丢弃
	}
}

// FromStandardData 将 StandardData 转为 LogEntry
func FromStandardData(data *model.StandardData, remoteAddr string) *LogEntry {
	return &LogEntry{
		DeviceID:    data.DeviceID,
		Protocol:    data.Protocol,
		Timestamp:   data.Timestamp.Format("2006-01-02 15:04:05.000"),
		MsgID:       data.MsgID,
		Voltage:     data.Voltage,
		Current:     data.Current,
		Power:       data.Power,
		EnergyTotal: data.EnergyTotal,
		Status:      data.ChargingStatus,
		Temperature: data.Temperature,
		FaultCode:   data.FaultCode,
		RemoteAddr:  remoteAddr,
		Type:        "data",
	}
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	close(h.stopped)
}

// ClientCount 返回当前客户端数量
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
