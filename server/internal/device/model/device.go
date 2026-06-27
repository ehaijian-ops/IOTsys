package model

import "time"

// Device 设备实体
type Device struct {
	ID               string     `json:"id" gorm:"primaryKey;size:36"`
	SN               string     `json:"sn" gorm:"uniqueIndex;size:64;not null"`
	DeviceType       string     `json:"device_type" gorm:"size:20;not null"`    // ebike_charger / ev_charger
	Protocol         string     `json:"protocol" gorm:"size:50;not null"`       // AP3000_v2 / TF100_v1
	Manufacturer     string     `json:"manufacturer" gorm:"size:100"`           // 设备厂家/制造商
	Model            string     `json:"model" gorm:"size:100"`
	SiteID           string     `json:"site_id" gorm:"index;size:36"`
	InstallLocation  string     `json:"install_location" gorm:"size:255"`
	PortCount        int        `json:"port_count" gorm:"default:1;not null"`   // 设备端口数量
	BillingSchemeID  string     `json:"billing_scheme_id" gorm:"size:36"`       // 关联收费方案
	FirmwareVersion  string     `json:"firmware_version" gorm:"size:50"`
	RateVoltage      float64    `json:"rate_voltage" gorm:"type:decimal(8,2);default:220"`   // 额定电压
	RateCurrent      float64    `json:"rate_current" gorm:"type:decimal(8,2);default:32"`    // 额定电流
	RatePower        float64    `json:"rate_power" gorm:"type:decimal(10,2);default:7000"`   // 额定功率(W)
	Disabled         bool       `json:"disabled" gorm:"default:false"`                       // 是否禁用
	Status           string     `json:"status" gorm:"size:20;default:offline"` // online/offline/fault/maintenance
	LastOnlineAt     *time.Time `json:"last_online_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (Device) TableName() string {
	return "devices"
}

// Site 站点
type Site struct {
	ID             string    `json:"id" gorm:"primaryKey;size:36"`
	Name           string    `json:"name" gorm:"size:100;not null"`
	Address        string    `json:"address" gorm:"size:255"`
	Province       string    `json:"province" gorm:"size:32;index"`
	City           string    `json:"city" gorm:"size:32"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Contact        string    `json:"contact" gorm:"size:50"`
	Phone          string    `json:"phone" gorm:"size:20"`
	AgentID        string    `json:"agent_id" gorm:"index;size:36"`                        // 所属代理商
	OperatorID     string    `json:"operator_id" gorm:"index;size:36"`                     // 所属运营商
	CommissionRate float64   `json:"commission_rate" gorm:"type:decimal(5,4);default:0"`   // 分佣比例
	BillingSchemeID string   `json:"billing_scheme_id" gorm:"size:36"`                     // 默认收费方案
	BusinessHours  string    `json:"business_hours" gorm:"size:100"`                       // 营业时间
	ImageURL       string    `json:"image_url" gorm:"size:500"`                            // 站点图片
	Status         string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Site) TableName() string {
	return "sites"
}
