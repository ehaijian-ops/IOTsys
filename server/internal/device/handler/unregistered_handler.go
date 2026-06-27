package handler

import (
	"fmt"
	"net/http"

	"iot-platform/internal/connector/tcpserver"
	"iot-platform/internal/device/model"
	"iot-platform/internal/device/service"
	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UnregisteredDeviceHandler 未注册设备处理器
type UnregisteredDeviceHandler struct {
	tcpSrv *tcpserver.Server
	devSvc *service.DeviceService
	db     *gorm.DB
}

// NewUnregisteredDeviceHandler 创建未注册设备处理器
func NewUnregisteredDeviceHandler(tcpSrv *tcpserver.Server, devSvc *service.DeviceService, db *gorm.DB) *UnregisteredDeviceHandler {
	return &UnregisteredDeviceHandler{
		tcpSrv: tcpSrv,
		devSvc: devSvc,
		db:     db,
	}
}

// UnregisteredDevice 未注册设备信息
type UnregisteredDevice struct {
	DeviceID      string `json:"device_id"`
	Protocol      string `json:"protocol"`
	SimCardNumber string `json:"sim_card_number,omitempty"`
	PortCount     int    `json:"port_count"` // 设备上报的端口/枪数量，0 表示尚未上报
	RemoteAddr    string `json:"remote_addr"`
	ConnectedAt   string `json:"connected_at"`
	LastActive    string `json:"last_active"`
	LastMsgType   string `json:"last_msg_type,omitempty"` // 最近报文类型: register/login/heartbeat
}

// ListUnregisteredDevices 获取未注册设备列表
// GET /api/v1/devices/unregistered
func (h *UnregisteredDeviceHandler) ListUnregisteredDevices(c *gin.Context) {
	// 1. 获取所有在线设备会话
	briefs := h.tcpSrv.GetOnlineSessionBriefs()
	if len(briefs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": "SUCCESS",
			"data": []UnregisteredDevice{},
			"total": 0,
		})
		return
	}

	// 2. 收集所有 deviceID，批量查询 MySQL 中已注册的设备
	deviceIDs := make([]string, 0, len(briefs))
	for _, b := range briefs {
		deviceIDs = append(deviceIDs, b.DeviceID)
	}

	var registeredSNs []string
	if err := h.db.Model(&model.Device{}).
		Where("sn IN ?", deviceIDs).
		Pluck("sn", &registeredSNs).Error; err != nil {
		logger.Error("Failed to query registered devices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "查询设备失败"})
		return
	}

	// 3. 构建已注册SN集合
	registeredSet := make(map[string]bool, len(registeredSNs))
	for _, sn := range registeredSNs {
		registeredSet[sn] = true
	}

	// 4. 筛选未注册设备
	unregistered := make([]UnregisteredDevice, 0)
	for _, b := range briefs {
		if registeredSet[b.DeviceID] {
			continue
		}
		unregistered = append(unregistered, UnregisteredDevice{
			DeviceID:      b.DeviceID,
			Protocol:      b.Protocol,
			SimCardNumber: b.SimCardNumber,
			PortCount:     b.PortCount,
			RemoteAddr:    b.RemoteAddr,
			ConnectedAt:   b.ConnectedAt.Format("2006-01-02 15:04:05"),
			LastActive:    b.LastActive.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  "SUCCESS",
		"data":  unregistered,
		"total": len(unregistered),
	})
}

// AddDeviceToSiteRequest 将未注册设备添加到站点请求
type AddDeviceToSiteRequest struct {
	DeviceID        string `json:"device_id" binding:"required"`   // 协议层设备ID
	Protocol        string `json:"protocol" binding:"required"`     // 协议类型（自动识别）
	DeviceType      string `json:"device_type" binding:"required"`  // 设备类型: ebike_charger / ev_charger
	SiteID          string `json:"site_id" binding:"required"`      // 站点ID
	InstallLocation string `json:"install_location"`                // 安装位置
	PortCount       int    `json:"port_count"`                      // 设备端口数量
	Manufacturer    string `json:"manufacturer"`                    // 设备厂家
}

const maxPortCount = 100

// AddToSite 将未注册设备添加到站点
// POST /api/v1/devices/unregistered/add
func (h *UnregisteredDeviceHandler) AddToSite(c *gin.Context) {
	var req AddDeviceToSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ERROR", "message": "参数错误: " + err.Error()})
		return
	}

	// 校验端口数量
	if req.PortCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ERROR", "message": "端口数量必须大于 0"})
		return
	}
	if req.PortCount > maxPortCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERROR",
			"message": fmt.Sprintf("端口数量不能超过最大限制 %d", maxPortCount),
		})
		return
	}

	// 验证站点是否存在
	var site model.Site
	if err := h.db.Where("id = ?", req.SiteID).First(&site).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ERROR", "message": "站点不存在"})
		return
	}

	// 创建设备记录（协议层 deviceID 作为 SN）
	createReq := &service.CreateDeviceRequest{
		SN:              req.DeviceID,
		DeviceType:      req.DeviceType,
		Protocol:        req.Protocol,
		SiteID:          req.SiteID,
		InstallLocation: req.InstallLocation,
		PortCount:       req.PortCount,
		Manufacturer:    req.Manufacturer,
	}

	device, err := h.devSvc.CreateDevice(c.Request.Context(), createReq)
	if err != nil {
		logger.Error("Failed to add device to site", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERROR",
			"message": "添加设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "设备已添加到站点 " + site.Name,
		"data":    device,
	})
}

// SiteBrief 站点简要信息
type SiteBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListSites 获取站点列表（简要，供下拉选择）
// GET /api/v1/sites
func (h *UnregisteredDeviceHandler) ListSites(c *gin.Context) {
	var sites []model.Site
	if err := h.db.Where("status = ?", "active").Order("name ASC").Find(&sites).Error; err != nil {
		logger.Error("Failed to list sites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "查询站点失败"})
		return
	}

	briefs := make([]SiteBrief, 0, len(sites))
	for _, s := range sites {
		briefs = append(briefs, SiteBrief{
			ID:   s.ID,
			Name: s.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": briefs,
	})
}
