package service

import (
	"context"
	"fmt"
	"time"

	"iot-platform/internal/device/model"
	"iot-platform/internal/device/repository"
	redisClient "iot-platform/pkg/database/redis"
	"iot-platform/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeviceService 设备管理服务
type DeviceService struct {
	repo *repository.DeviceRepository
}

// NewDeviceService 创建设备服务
func NewDeviceService(repo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error) {
	portCount := req.PortCount
	if portCount <= 0 {
		portCount = 1
	}
	device := &model.Device{
		ID:              uuid.New().String(),
		SN:              req.SN,
		DeviceType:      req.DeviceType,
		Protocol:        req.Protocol,
		Manufacturer:    req.Manufacturer,
		Model:           req.Model,
		SiteID:          req.SiteID,
		InstallLocation: req.InstallLocation,
		PortCount:       portCount,
		FirmwareVersion: req.FirmwareVersion,
		Status:          "offline",
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	logger.Info("Device created",
		zap.String("id", device.ID),
		zap.String("sn", device.SN),
		zap.String("protocol", device.Protocol),
	)

	return device, nil
}

// GetDevice 获取设备详情
func (s *DeviceService) GetDevice(ctx context.Context, id string) (*DeviceDetail, error) {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &DeviceDetail{
		Device: *device,
	}

	// 从 Redis 获取实时数据（优先用 UUID 查询，失败时用 SN 查询）
	realtimeKey := fmt.Sprintf("device:realtime:%s", id)
	realtime, err := redisClient.Client.HGetAll(ctx, realtimeKey).Result()
	if (err != nil || len(realtime) == 0) && device.SN != "" {
		realtimeKey = fmt.Sprintf("device:realtime:%s", device.SN)
		realtime, err = redisClient.Client.HGetAll(ctx, realtimeKey).Result()
	}
	if err == nil && len(realtime) > 0 {
		detail.RealtimeData = realtime
	}

	// 获取在线状态：先用 DB UUID 查，失败时用 SN（= 协议层 DeviceID）查
	detail.IsOnline = isDeviceOnline(ctx, id, device.SN)

	return detail, nil
}

// ListDevices 分页查询设备列表
func (s *DeviceService) ListDevices(ctx context.Context, query repository.DeviceQuery) ([]model.Device, int64, error) {
	// 在线状态查询通过 Redis 实时获取，不在 SQL 层按 status 筛选
	// （数据库 status 字段是异步更新的缓存值，可能滞后于 Redis 实时状态）
	query.SkipStatusFilter = true
	devices, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 从 Redis 检查每个设备的实时在线状态，覆盖数据库中的旧状态
	// 优先用 DB UUID 查，失败时用 SN（= 协议层 DeviceID）查
	for i := range devices {
		if isDeviceOnline(ctx, devices[i].ID, devices[i].SN) {
			devices[i].Status = "online"
		}
	}

	return devices, total, nil
}

// UpdateDevice 更新设备信息
func (s *DeviceService) UpdateDevice(ctx context.Context, id string, req *UpdateDeviceRequest) (*model.Device, error) {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.DeviceType != "" {
		device.DeviceType = req.DeviceType
	}
	if req.Protocol != "" {
		device.Protocol = req.Protocol
	}
	if req.Manufacturer != "" {
		device.Manufacturer = req.Manufacturer
	}
	if req.Model != "" {
		device.Model = req.Model
	}
	if req.SiteID != "" {
		device.SiteID = req.SiteID
	}
	if req.InstallLocation != "" {
		device.InstallLocation = req.InstallLocation
	}
	if req.FirmwareVersion != "" {
		device.FirmwareVersion = req.FirmwareVersion
	}
	if req.PortCount != nil {
		device.PortCount = *req.PortCount
	}
	if req.Status != "" {
		device.Status = req.Status
	}
	device.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	return device, nil
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	SN              string `json:"sn" binding:"required"`
	DeviceType      string `json:"device_type" binding:"required"`
	Protocol        string `json:"protocol" binding:"required"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	SiteID          string `json:"site_id"`
	InstallLocation string `json:"install_location"`
	PortCount       int    `json:"port_count"`
	FirmwareVersion string `json:"firmware_version"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	DeviceType      string `json:"device_type"`
	Protocol        string `json:"protocol"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	SiteID          string `json:"site_id"`
	InstallLocation string `json:"install_location"`
	PortCount       *int   `json:"port_count"`
	FirmwareVersion string `json:"firmware_version"`
	Status          string `json:"status"`
}

// DeviceDetail 设备详情（含实时数据）
type DeviceDetail struct {
	model.Device
	IsOnline     bool                   `json:"is_online"`
	RealtimeData map[string]string      `json:"realtime_data,omitempty"`
}

// isDeviceOnline 检查设备是否在线（优先用 UUID，fallback 用 SN）
// 原因：tcpserver 使用协议层 DeviceID 设置 Redis key，而协议层 DeviceID == 数据库 SN 字段
func isDeviceOnline(ctx context.Context, dbID string, sn string) bool {
	// 优先用 DB UUID 查询
	onlineKey := "device:online:" + dbID
	online, err := redisClient.Client.Get(ctx, onlineKey).Result()
	if err == nil && online == "1" {
		return true
	}
	// fallback: 用 SN 查询（SN = 协议层 DeviceID）
	if sn != "" {
		onlineKey = "device:online:" + sn
		online, err = redisClient.Client.Get(ctx, onlineKey).Result()
		return err == nil && online == "1"
	}
	return false
}
