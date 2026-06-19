package handler

import (
	"net/http"

	"iot-platform/internal/command/service"
	"iot-platform/pkg/errors"
	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommandHandler 指令 HTTP 处理器
type CommandHandler struct {
	svc *service.CommandService
}

// NewCommandHandler 创建指令处理器
func NewCommandHandler(svc *service.CommandService) *CommandHandler {
	return &CommandHandler{svc: svc}
}

// CreateCommand 下发指令
// POST /api/v1/devices/:id/commands
func (h *CommandHandler) CreateCommand(c *gin.Context) {
	deviceID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req struct {
		CmdType string                 `json:"cmd_type" binding:"required"`
		Params  map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	cmd, err := h.svc.CreateCommand(c.Request.Context(), deviceID, req.CmdType, req.Params, userID.(string))
	if err != nil {
		logger.Error("Failed to create command",
			zap.String("device_id", deviceID),
			zap.String("cmd_type", req.CmdType),
			zap.Error(err),
		)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "指令已下发",
		"data":    cmd,
	})
}

// GetCommand 获取指令详情
// GET /api/v1/commands/:id
func (h *CommandHandler) GetCommand(c *gin.Context) {
	cmd, err := h.svc.GetCommand(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": cmd,
	})
}

// ListCommands 查询设备指令列表
// GET /api/v1/devices/:id/commands
func (h *CommandHandler) ListCommands(c *gin.Context) {
	deviceID := c.Param("id")
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	commands, total, err := h.svc.ListCommands(c.Request.Context(), deviceID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  "SUCCESS",
		"data":  commands,
		"total": total,
		"page":  page,
	})
}
