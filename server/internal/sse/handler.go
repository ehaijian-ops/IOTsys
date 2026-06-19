package sse

import (
	"encoding/json"
	"fmt"
	"time"

	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HandleSSE 处理 SSE 连接请求
// GET /api/v1/devices/logs/stream?device_id=xxx&protocol=xxx&keyword=xxx
func (h *Hub) HandleSSE(c *gin.Context) {
	// 解析过滤参数
	filter := ClientFilter{
		DeviceID: c.Query("device_id"),
		Protocol: c.Query("protocol"),
		Keyword:  c.Query("keyword"),
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	client := &Client{
		ID:     uuid.New().String(),
		Events: make(chan *LogEntry, 256),
		Done:   make(chan struct{}),
		Filter: filter,
	}

	h.Register(client)

	ctx := c.Request.Context()
	clientGone := c.Writer.CloseNotify()

	logger.Info("SSE stream started",
		zap.String("client_id", client.ID),
		zap.String("device_id", filter.DeviceID),
		zap.String("protocol", filter.Protocol),
	)

	defer func() {
		h.Unregister(client)
		logger.Debug("SSE stream ended", zap.String("client_id", client.ID))
	}()

	// 发送初始连接确认
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"client_id\":\"%s\"}\n\n", client.ID)
	c.Writer.Flush()

	// 心跳 + 数据推送
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-clientGone:
			return
		case entry, ok := <-client.Events:
			if !ok {
				return
			}
			// 序列化日志条目
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", string(data))
			c.Writer.Flush()
		case <-ticker.C:
			// 心跳保活
			if _, err := fmt.Fprintf(c.Writer, ": heartbeat\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
