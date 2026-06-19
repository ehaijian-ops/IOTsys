package repository

import (
	"context"
	"time"

	"iot-platform/internal/device/model"
	"iot-platform/pkg/errors"

	"gorm.io/gorm"
)

// DeviceRepository 设备数据访问层
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 创建设备仓库
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create 创建设备
func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

// GetByID 根据 ID 查询设备
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Device", id)
		}
		return nil, err
	}
	return &device, nil
}

// GetBySN 根据 SN 查询设备
func (r *DeviceRepository) GetBySN(ctx context.Context, sn string) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Where("sn = ?", sn).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Device", sn)
		}
		return nil, err
	}
	return &device, nil
}

// Update 更新设备
func (r *DeviceRepository) Update(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

// Delete 删除设备（软删除）
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Device{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.NotFound("Device", id)
	}
	return result.Error
}

// List 分页查询设备列表
func (r *DeviceRepository) List(ctx context.Context, query DeviceQuery) ([]model.Device, int64, error) {
	var devices []model.Device
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Device{})

	// 筛选条件
	if query.DeviceType != "" {
		db = db.Where("device_type = ?", query.DeviceType)
	}
	if query.Protocol != "" {
		db = db.Where("protocol = ?", query.Protocol)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.SiteID != "" {
		db = db.Where("site_id = ?", query.SiteID)
	}
	if query.Keyword != "" {
		db = db.Where("sn LIKE ? OR vendor LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	db.Count(&total)

	if err := db.Order("created_at DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&devices).Error; err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// UpdateStatus 更新设备状态
func (r *DeviceRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).Model(&model.Device{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         status,
			"last_online_at": time.Now(),
		}).Error
}

// DeviceQuery 设备查询参数
type DeviceQuery struct {
	DeviceType string `form:"device_type"`
	Protocol   string `form:"protocol"`
	Status     string `form:"status"`
	SiteID     string `form:"site_id"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}
