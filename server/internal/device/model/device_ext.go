package model

import "time"

// DeviceType 设备类型定义
type DeviceType struct {
	ID          string    `json:"id" gorm:"primaryKey;size:36"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:30;not null"` // ebike_charger/ev_charger
	Protocol    string    `json:"protocol" gorm:"size:50"`                  // 默认协议
	PortCount   int       `json:"port_count" gorm:"default:1"`              // 默认端口数
	Icon        string    `json:"icon" gorm:"size:255"`                     // 图标
	Description string    `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DeviceType) TableName() string {
	return "device_types"
}

// OTAFirmware OTA固件升级包
type OTAFirmware struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	DeviceTypeID string    `json:"device_type_id" gorm:"index;size:36"`
	Version      string    `json:"version" gorm:"size:50;not null"`          // 固件版本号
	FileName     string    `json:"file_name" gorm:"size:255"`
	FileURL      string    `json:"file_url" gorm:"size:500;not null"`        // 下载地址
	FileSize     int64     `json:"file_size" gorm:"default:0"`               // 文件大小
	MD5          string    `json:"md5" gorm:"size:32"`                       // 校验和
	SHA256       string    `json:"sha256" gorm:"size:64"`
	Changelog    string    `json:"changelog" gorm:"type:text"`               // 更新日志
	ForceUpdate  bool      `json:"force_update" gorm:"default:false"`        // 是否强制升级
	Status       string    `json:"status" gorm:"size:20;default:active"`     // active/inactive
	CreatedBy    string    `json:"created_by" gorm:"size:64"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (OTAFirmware) TableName() string {
	return "ota_firmwares"
}

// DeviceQRCode 设备二维码
type DeviceQRCode struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	DeviceID   string    `json:"device_id" gorm:"index;size:36;not null"`
	PortNumber int       `json:"port_number" gorm:"default:1"`              // 端口号（多口设备）
	QRCodeURL  string    `json:"qrcode_url" gorm:"size:500"`                // 二维码图片地址
	QRContent  string    `json:"qr_content" gorm:"size:500"`                // 二维码内容
	CreatedAt  time.Time `json:"created_at"`
}

func (DeviceQRCode) TableName() string {
	return "device_qrcodes"
}

// DeviceOnlineLog 设备上下线记录
type DeviceOnlineLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	DeviceID  string    `json:"device_id" gorm:"index;size:36;not null"`
	Event     string    `json:"event" gorm:"size:20;not null"` // online/offline
	Reason    string    `json:"reason" gorm:"size:255"`        // 下线原因
	IP        string    `json:"ip" gorm:"size:64"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func (DeviceOnlineLog) TableName() string {
	return "device_online_logs"
}

// VirtualDevice 虚拟设备（模拟器用）
type VirtualDevice struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	DeviceID  string    `json:"device_id" gorm:"uniqueIndex;size:36;not null"`
	Protocol  string    `json:"protocol" gorm:"size:50"`
	SimData   string    `json:"sim_data" gorm:"type:text"`           // 模拟数据配置（JSON）
	Status    string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (VirtualDevice) TableName() string {
	return "virtual_devices"
}
