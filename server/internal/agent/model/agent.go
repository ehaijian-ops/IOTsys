package model

import "time"

// Agent 代理商
type Agent struct {
	ID             string    `json:"id" gorm:"primaryKey;size:36"`
	Name           string    `json:"name" gorm:"size:100;not null"`
	Contact        string    `json:"contact" gorm:"size:50"`
	Phone          string    `json:"phone" gorm:"size:20;index"`
	Email          string    `json:"email" gorm:"size:128"`
	Address        string    `json:"address" gorm:"size:255"`
	CommissionRate float64   `json:"commission_rate" gorm:"type:decimal(5,4);default:0"` // 分佣比例
	Balance        float64   `json:"balance" gorm:"type:decimal(12,2);default:0"`         // 可提现余额
	TotalRevenue   float64   `json:"total_revenue" gorm:"type:decimal(12,2);default:0"`    // 累计收益
	Status         string    `json:"status" gorm:"size:20;default:active;index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Agent) TableName() string {
	return "agents"
}

// Operator 运营商
type Operator struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	AgentID   string    `json:"agent_id" gorm:"index;size:36"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Contact   string    `json:"contact" gorm:"size:50"`
	Phone     string    `json:"phone" gorm:"size:20;index"`
	Address   string    `json:"address" gorm:"size:255"`
	SiteIDs   string    `json:"site_ids" gorm:"type:text"`                             // 关联站点ID列表（JSON数组）
	Status    string    `json:"status" gorm:"size:20;default:active;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Operator) TableName() string {
	return "operators"
}
