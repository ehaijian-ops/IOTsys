package model

import "time"

// Device 设备实体
type Device struct {
	ID               string    `json:"id" gorm:"primaryKey;size:36"`
	SN               string    `json:"sn" gorm:"uniqueIndex;size:64;not null"`
	DeviceType       string    `json:"device_type" gorm:"size:20;not null"`    // ebike_charger / ev_charger
	Protocol         string    `json:"protocol" gorm:"size:50;not null"`       // AP3000_v2 / TF100_v1
	Manufacturer     string    `json:"manufacturer" gorm:"size:100"`           // 设备厂家/制造商
	Model            string    `json:"model" gorm:"size:100"`
	SiteID           string    `json:"site_id" gorm:"index;size:36"`
	InstallLocation  string    `json:"install_location" gorm:"size:255"`
	PortCount        int       `json:"port_count" gorm:"default:1;not null"`  // 设备端口数量
	FirmwareVersion  string    `json:"firmware_version" gorm:"size:50"`
	Status           string    `json:"status" gorm:"size:20;default:offline"` // online/offline/fault/maintenance
	LastOnlineAt     *time.Time `json:"last_online_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Device) TableName() string {
	return "devices"
}

// Site 站点
type Site struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Address   string    `json:"address" gorm:"size:255"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Contact   string    `json:"contact" gorm:"size:50"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Status    string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Site) TableName() string {
	return "sites"
}
