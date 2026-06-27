package model

import "time"

// 提现状态常量
const (
	WithdrawPending  = "pending"  // 待处理
	WithdrawApproved = "approved" // 已打款
	WithdrawRejected = "rejected" // 已驳回
)

// 充值状态常量
const (
	RechargePending = "pending" // 待支付
	RechargeSuccess = "success" // 已到账
	RechargeFailed  = "failed"  // 已失败
)

// 分成状态常量
const (
	SplitPending   = "pending"   // 待分账
	SplitCompleted = "completed" // 已分账
)

// UserWallet 用户钱包
type UserWallet struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	Balance        float64   `json:"balance" gorm:"type:decimal(12,2);default:0"`       // 可用余额
	TotalRecharge  float64   `json:"total_recharge" gorm:"type:decimal(12,2);default:0"` // 累计充值
	TotalConsume   float64   `json:"total_consume" gorm:"type:decimal(12,2);default:0"`  // 累计消费
	FrozenAmount   float64   `json:"frozen_amount" gorm:"type:decimal(12,2);default:0"`  // 冻结金额
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (UserWallet) TableName() string {
	return "user_wallets"
}

// RechargeRecord 充值记录
type RechargeRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;size:36"`
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	Amount      float64   `json:"amount" gorm:"type:decimal(10,2);not null"`
	BonusAmount float64   `json:"bonus_amount" gorm:"type:decimal(10,2);default:0"`  // 赠送金额
	PayMethod   string    `json:"pay_method" gorm:"size:20;not null"`                // wechat/alipay/admin/card
	TradeNo     string    `json:"trade_no" gorm:"size:64"`                           // 第三方交易号
	Status      string    `json:"status" gorm:"size:20;default:pending;index"`
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RechargeRecord) TableName() string {
	return "recharge_records"
}

// WithdrawRecord 提现记录
type WithdrawRecord struct {
	ID           string     `json:"id" gorm:"primaryKey;size:36"`
	UserID       *uint      `json:"user_id" gorm:"index"`              // 终端用户提现
	AgentID      *string    `json:"agent_id" gorm:"index;size:36"`    // 代理商/商户提现
	Amount       float64    `json:"amount" gorm:"type:decimal(12,2);not null"`
	ServiceFee   float64    `json:"service_fee" gorm:"type:decimal(10,2);default:0"` // 手续费
	ActualAmount float64    `json:"actual_amount" gorm:"type:decimal(12,2);default:0"` // 实际到账
	BankName     string     `json:"bank_name" gorm:"size:100"`
	BankCardNo   string     `json:"bank_card_no" gorm:"size:50"`
	BankAccount  string     `json:"bank_account" gorm:"size:50"`
	Status       string     `json:"status" gorm:"size:20;default:pending;index"`
	Remark       string     `json:"remark" gorm:"size:500"`
	ProcessedBy  string     `json:"processed_by" gorm:"size:64"`
	ProcessedAt  *time.Time `json:"processed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (WithdrawRecord) TableName() string {
	return "withdraw_records"
}

// RevenueSplit 分成分账记录
type RevenueSplit struct {
	ID              string    `json:"id" gorm:"primaryKey;size:36"`
	OrderID         string    `json:"order_id" gorm:"index;size:36;not null"`
	TotalAmount     float64   `json:"total_amount" gorm:"type:decimal(10,2);not null"` // 分成总金额
	AgentID         *string   `json:"agent_id" gorm:"index;size:36"`                    // 代理商ID
	AgentAmount     float64   `json:"agent_amount" gorm:"type:decimal(10,2);default:0"`  // 代理商分成
	AgentRate       float64   `json:"agent_rate" gorm:"type:decimal(5,4);default:0"`     // 代理商分成比例
	OperatorID      *string   `json:"operator_id" gorm:"index;size:36"`                  // 运营商ID
	OperatorAmount  float64   `json:"operator_amount" gorm:"type:decimal(10,2);default:0"` // 运营商分成
	OperatorRate    float64   `json:"operator_rate" gorm:"type:decimal(5,4);default:0"`   // 运营商分成比例
	PlatformAmount  float64   `json:"platform_amount" gorm:"type:decimal(10,2);default:0"` // 平台分成
	Status          string    `json:"status" gorm:"size:20;default:pending;index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (RevenueSplit) TableName() string {
	return "revenue_splits"
}
