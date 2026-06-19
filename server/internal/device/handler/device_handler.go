package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/device/repository"
	"iot-platform/internal/device/service"
	"iot-platform/pkg/errors"
	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeviceHandler 设备管理 HTTP 处理器
type DeviceHandler struct {
	svc *service.DeviceService
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// CreateDevice 创建设备
// POST /api/v1/devices
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req service.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	device, err := h.svc.CreateDevice(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create device", zap.Error(err))
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "SUCCESS",
		"message": "设备创建成功",
		"data":    device,
	})
}

// GetDevice 获取设备详情
// GET /api/v1/devices/:id
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	device, err := h.svc.GetDevice(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": device,
	})
}

// ListDevices 设备列表
// GET /api/v1/devices
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	var query repository.DeviceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	query.Page = page
	query.PageSize = pageSize

	devices, total, err := h.svc.ListDevices(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  "SUCCESS",
		"data":  devices,
		"total": total,
		"page":  page,
	})
}

// UpdateDevice 更新设备
// PUT /api/v1/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	var req service.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	device, err := h.svc.UpdateDevice(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "设备更新成功",
		"data":    device,
	})
}

// DeleteDevice 删除设备
// DELETE /api/v1/devices/:id
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	if err := h.svc.DeleteDevice(c.Request.Context(), c.Param("id")); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "设备已删除",
	})
}
