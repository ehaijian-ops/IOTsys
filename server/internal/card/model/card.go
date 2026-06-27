package model

import "time"

// 卡片状态常量
const (
	CardStatusActive  = "active"  // 正常
	CardStatusLost    = "lost"    // 已挂失
	CardStatusFrozen  = "frozen"  // 已冻结
	CardStatusExpired = "expired" // 已过期
)

// ICCard IC充电卡
type ICCard struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	CardNo    string    `json:"card_no" gorm:"uniqueIndex;size:64;not null"`        // 卡号（物理卡号）
	CardUID   string    `json:"card_uid" gorm:"index;size:64"`                      // IC卡UID
	UserID    *uint     `json:"user_id" gorm:"index"`                               // 绑定用户ID
	Balance   float64   `json:"balance" gorm:"type:decimal(10,2);default:0"`        // 卡内余额
	TotalRecharge float64 `json:"total_recharge" gorm:"type:decimal(10,2);default:0"` // 累计充值
	Status    string    `json:"status" gorm:"size:20;default:active;index"`
	IssuedAt  *time.Time `json:"issued_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (ICCard) TableName() string {
	return "ic_cards"
}

// ICCardRecharge IC卡充值记录
type ICCardRecharge struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	CardID    string    `json:"card_id" gorm:"index;size:36;not null"`
	Amount    float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	PayMethod string    `json:"pay_method" gorm:"size:20"`
	TradeNo   string    `json:"trade_no" gorm:"size:64"`
	CreatedBy string    `json:"created_by" gorm:"size:64"`
	CreatedAt time.Time `json:"created_at"`
}

func (ICCardRecharge) TableName() string {
	return "ic_card_recharges"
}

// TrafficCard 流量卡（物联卡）
type TrafficCard struct {
	ID          string     `json:"id" gorm:"primaryKey;size:36"`
	ICCID       string     `json:"iccid" gorm:"uniqueIndex;size:32;not null"`       // SIM卡ICCID
	IMSI        string     `json:"imsi" gorm:"size:20"`
	DeviceID    *string    `json:"device_id" gorm:"index;size:36"`                  // 绑定设备
	Carrier     string     `json:"carrier" gorm:"size:20"`    // 运营商: china_mobile/china_telecom/china_unicom
	PlanName    string     `json:"plan_name" gorm:"size:100"` // 套餐名称
	DataUsedMB  float64    `json:"data_used_mb" gorm:"type:decimal(10,2);default:0"` // 已用流量(MB)
	DataTotalMB float64    `json:"data_total_mb" gorm:"type:decimal(10,2);default:0"` // 总流量(MB)
	Status      string     `json:"status" gorm:"size:20;default:active;index"`
	ActivatedAt *time.Time `json:"activated_at"`
	ExpiredAt   *time.Time `json:"expired_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (TrafficCard) TableName() string {
	return "traffic_cards"
}

// MonthlyCard 月卡记录
type MonthlyCard struct {
	ID          string     `json:"id" gorm:"primaryKey;size:36"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	SchemeID    string     `json:"scheme_id" gorm:"size:36;not null"`
	SchemeName  string     `json:"scheme_name" gorm:"size:100"`
	DeviceType  string     `json:"device_type" gorm:"size:20"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	UsedHours   int        `json:"used_hours" gorm:"default:0"`                     // 已使用小时数
	UsedEnergy  float64    `json:"used_energy" gorm:"type:decimal(10,2);default:0"`  // 已使用电量
	Status      string     `json:"status" gorm:"size:20;default:active;index"`      // active/expired/cancelled
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (MonthlyCard) TableName() string {
	return "monthly_cards"
}

// MonthlyCardRecord 月卡操作记录
type MonthlyCardRecord struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CardID      string    `json:"card_id" gorm:"index;size:36;not null"`
	Operation   string    `json:"operation" gorm:"size:30;not null"` // create/renew/cancel/refund
	Remark      string    `json:"remark" gorm:"size:500"`
	OperatorBy  string    `json:"operator_by" gorm:"size:64"`
	CreatedAt   time.Time `json:"created_at"`
}

func (MonthlyCardRecord) TableName() string {
	return "monthly_card_records"
}
