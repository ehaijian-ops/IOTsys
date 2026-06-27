package model

import "time"

// 收费方案类型常量
const (
	BillingTypeCar   = "car"   // 汽车收费方案（尖峰平谷）
	BillingTypeTime  = "time"  // 时间收费方案
	BillingTypeEnergy = "energy" // 电量收费方案
	BillingTypePower  = "power"  // 功率收费方案
)

// 时段类型常量
const (
	PeriodPeak  = "peak"  // 尖峰
	PeriodFlat  = "flat"  // 平段
	PeriodValley = "valley" // 谷段
)

// BillingScheme 收费方案
type BillingScheme struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	Name          string    `json:"name" gorm:"size:100;not null"`
	Type          string    `json:"type" gorm:"size:20;not null;index"`            // car/time/energy/power
	DeviceType    string    `json:"device_type" gorm:"size:20;index"`              // ebike_charger/ev_charger
	SiteID        string    `json:"site_id" gorm:"index;size:36"`                  // 0=全局
	BaseServiceFee float64  `json:"base_service_fee" gorm:"type:decimal(10,4);default:0"` // 基础服务费（元/度）
	MaxPrice      float64   `json:"max_price" gorm:"type:decimal(10,2);default:0"` // 封顶价格
	FullStopFee   float64   `json:"full_stop_fee" gorm:"type:decimal(10,2);default:0"` // 充满自停费用
	PrepaidFee    float64   `json:"prepaid_fee" gorm:"type:decimal(10,2);default:0"`  // 预收费金额（功率方案）
	UnitPrice     float64   `json:"unit_price" gorm:"type:decimal(10,4);default:0"`   // 单价（时间/电量/功率方案）
	Unit          string    `json:"unit" gorm:"size:10"`                              // 计费单位: hour/kwh/min
	Status        string    `json:"status" gorm:"size:10;default:active"`            // active/inactive
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (BillingScheme) TableName() string {
	return "billing_schemes"
}

// BillingPeriod 收费方案时段配置（尖峰平谷）
type BillingPeriod struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SchemeID      string    `json:"scheme_id" gorm:"index;size:36;not null"`
	PeriodType    string    `json:"period_type" gorm:"size:20;not null"`             // peak/flat/valley
	StartTime     string    `json:"start_time" gorm:"size:5;not null"`               // HH:MM 格式
	EndTime       string    `json:"end_time" gorm:"size:5;not null"`                 // HH:MM 格式
	PricePerKWh   float64   `json:"price_per_kwh" gorm:"type:decimal(10,4);default:0"` // 电价（元/度）
	ServiceFee    float64   `json:"service_fee" gorm:"type:decimal(10,4);default:0"`   // 服务费（元/度）
	SortOrder     int       `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
}

func (BillingPeriod) TableName() string {
	return "billing_periods"
}

// MonthlyCardScheme 月卡套餐方案
type MonthlyCardScheme struct {
	ID          string    `json:"id" gorm:"primaryKey;size:36"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	DeviceType  string    `json:"device_type" gorm:"size:20;not null"`              // ebike_charger/ev_charger
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null"`         // 月卡价格
	DurationDays int      `json:"duration_days" gorm:"default:30"`                  // 有效天数
	FreeHours   int       `json:"free_hours" gorm:"default:0"`                      // 每日免费充电时长（小时）
	FreeEnergy  float64   `json:"free_energy" gorm:"type:decimal(10,2);default:0"`  // 每月免费电量（度）
	Description string    `json:"description" gorm:"size:500"`
	Status      string    `json:"status" gorm:"size:10;default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (MonthlyCardScheme) TableName() string {
	return "monthly_card_schemes"
}

// RechargeScheme 余额充值套餐方案
type RechargeScheme struct {
	ID          string    `json:"id" gorm:"primaryKey;size:36"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Amount      float64   `json:"amount" gorm:"type:decimal(10,2);not null"`        // 充值金额
	BonusAmount float64   `json:"bonus_amount" gorm:"type:decimal(10,2);default:0"` // 赠送金额
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	Status      string    `json:"status" gorm:"size:10;default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RechargeScheme) TableName() string {
	return "recharge_schemes"
}

// BusinessConfig 业务配置
type BusinessConfig struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey        string    `json:"config_key" gorm:"uniqueIndex;size:64;not null"`
	ConfigValue      string    `json:"config_value" gorm:"type:text;not null"`
	Description      string    `json:"description" gorm:"size:255"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (BusinessConfig) TableName() string {
	return "business_configs"
}
