package model

import "time"

// 订单类型常量
const (
	OrderTypeScan    = "scan"    // 扫码充电
	OrderTypeICCard  = "ic_card" // IC卡充电
	OrderTypeFree    = "free"    // 免费充电
	OrderTypeMonthly = "monthly" // 包月充电
)

// 充电方式常量
const (
	ChargeTypeTime  = "time"  // 按时间
	ChargeTypeEnergy = "energy" // 按电量
	ChargeTypePower  = "power"  // 按功率
)

// 订单状态常量
const (
	OrderStatusPending   = "pending"   // 待开始
	OrderStatusCharging  = "charging"  // 充电中
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusRefunded  = "refunded"  // 已退款
	OrderStatusCancelled = "cancelled" // 已取消
)

// ChargeOrder 充电订单（统一表，含所有订单类型）
type ChargeOrder struct {
	ID            string     `json:"id" gorm:"primaryKey;size:36"`
	OrderSN       string     `json:"order_sn" gorm:"uniqueIndex;size:64;not null"`
	UserID        *uint      `json:"user_id" gorm:"index"`
	DeviceID      string     `json:"device_id" gorm:"index;size:36;not null"`
	SiteID        string     `json:"site_id" gorm:"index;size:36"`
	PortNumber    int        `json:"port_number" gorm:"default:1"`
	OrderType     string     `json:"order_type" gorm:"size:20;not null;index"`     // scan/ic_card/free/monthly
	ChargeType    string     `json:"charge_type" gorm:"size:20"`                    // time/energy/power
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Duration      int64      `json:"duration" gorm:"default:0"`                     // 充电时长（秒）
	EnergyKWh     float64    `json:"energy_kwh" gorm:"type:decimal(10,2);default:0"` // 充电电量（度）
	PowerWatt     float64    `json:"power_watt" gorm:"type:decimal(10,2);default:0"` // 实时功率
	Amount        float64    `json:"amount" gorm:"type:decimal(10,2);default:0"`    // 订单金额（元）
	PaidAmount    float64    `json:"paid_amount" gorm:"type:decimal(10,2);default:0"` // 实付金额
	RefundAmount  float64    `json:"refund_amount" gorm:"type:decimal(10,2);default:0"` // 退款金额
	PlateNumber   string     `json:"plate_number" gorm:"size:20"`                   // 车牌号（汽车充电）
	BillingSchemeID string   `json:"billing_scheme_id" gorm:"size:36"`             // 收费方案ID
	PayMethod     string     `json:"pay_method" gorm:"size:20"`                     // 支付方式: wallet/ic_card/monthly
	Remark        string     `json:"remark" gorm:"size:500"`
	Status        string     `json:"status" gorm:"size:20;default:pending;index"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (ChargeOrder) TableName() string {
	return "charge_orders"
}

// ChargeRecord 充电过程记录（功率曲线数据）
type ChargeRecord struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID     string    `json:"order_id" gorm:"index;size:36;not null"`
	DeviceID    string    `json:"device_id" gorm:"index;size:36;not null"`
	RecordAt    time.Time `json:"record_at" gorm:"index;not null"`
	Voltage     float64   `json:"voltage" gorm:"type:decimal(8,2);default:0"`    // 电压 (V)
	Current     float64   `json:"current" gorm:"type:decimal(8,2);default:0"`    // 电流 (A)
	PowerWatt   float64   `json:"power_watt" gorm:"type:decimal(10,2);default:0"` // 功率 (W)
	EnergyKWh   float64   `json:"energy_kwh" gorm:"type:decimal(10,2);default:0"` // 累计电量 (kWh)
	Temperature float64   `json:"temperature" gorm:"type:decimal(5,1);default:0"` // 温度 (℃)
	CreatedAt   time.Time `json:"created_at"`
}

func (ChargeRecord) TableName() string {
	return "charge_records"
}

// OrderRefund 订单退款记录
type OrderRefund struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	OrderID      string    `json:"order_id" gorm:"index;size:36;not null"`
	RefundAmount float64   `json:"refund_amount" gorm:"type:decimal(10,2);not null"`
	RefundReason string    `json:"refund_reason" gorm:"size:500"`
	RefundMethod string    `json:"refund_method" gorm:"size:20;default:original"`  // original/wallet
	TradeNo      string    `json:"trade_no" gorm:"size:64"`                        // 退款交易号
	Status       string    `json:"status" gorm:"size:20;default:pending;index"`    // pending/success/failed
	ProcessedBy  string    `json:"processed_by" gorm:"size:64"`
	ProcessedAt  *time.Time `json:"processed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (OrderRefund) TableName() string {
	return "order_refunds"
}

// 订单列表查询参数
type OrderQuery struct {
	OrderSN    string `json:"order_sn" form:"order_sn"`
	OrderType  string `json:"order_type" form:"order_type"`
	Status     string `json:"status" form:"status"`
	DeviceID   string `json:"device_id" form:"device_id"`
	UserID     *uint  `json:"user_id" form:"user_id"`
	StartDate  string `json:"start_date" form:"start_date"`
	EndDate    string `json:"end_date" form:"end_date"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
}
