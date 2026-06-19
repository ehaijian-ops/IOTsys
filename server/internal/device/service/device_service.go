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
	device := &model.Device{
		ID:              uuid.New().String(),
		SN:              req.SN,
		DeviceType:      req.DeviceType,
		Protocol:        req.Protocol,
		Vendor:          req.Vendor,
		Model:           req.Model,
		SiteID:          req.SiteID,
		InstallLocation: req.InstallLocation,
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

	// 从 Redis 获取实时数据
	realtimeKey := fmt.Sprintf("device:realtime:%s", id)
	realtime, err := redisClient.Client.HGetAll(ctx, realtimeKey).Result()
	if err == nil && len(realtime) > 0 {
		detail.RealtimeData = realtime
	}

	// 获取在线状态
	onlineKey := fmt.Sprintf("device:online:%s", id)
	online, _ := redisClient.Client.Get(ctx, onlineKey).Result()
	detail.IsOnline = online == "1"

	return detail, nil
}

// ListDevices 分页查询设备列表
func (s *DeviceService) ListDevices(ctx context.Context, query repository.DeviceQuery) ([]model.Device, int64, error) {
	return s.repo.List(ctx, query)
}

// UpdateDevice 更新设备信息
func (s *DeviceService) UpdateDevice(ctx context.Context, id string, req *UpdateDeviceRequest) (*model.Device, error) {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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
	Vendor          string `json:"vendor"`
	Model           string `json:"model"`
	SiteID          string `json:"site_id"`
	InstallLocation string `json:"install_location"`
	FirmwareVersion string `json:"firmware_version"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	SiteID          string `json:"site_id"`
	InstallLocation string `json:"install_location"`
	FirmwareVersion string `json:"firmware_version"`
	Status          string `json:"status"`
}

// DeviceDetail 设备详情（含实时数据）
type DeviceDetail struct {
	model.Device
	IsOnline     bool                   `json:"is_online"`
	RealtimeData map[string]string      `json:"realtime_data,omitempty"`
}
