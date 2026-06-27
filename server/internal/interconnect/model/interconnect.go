package model

import "time"

// InterconnectOrg 互联机构
type InterconnectOrg struct {
	ID                string    `json:"id" gorm:"primaryKey;size:36"`
	Name              string    `json:"name" gorm:"size:100;not null"`
	OrgCode           string    `json:"org_code" gorm:"uniqueIndex;size:64;not null"` // 机构编码
	Contact           string    `json:"contact" gorm:"size:50"`
	Phone             string    `json:"phone" gorm:"size:20"`
	PushURL           string    `json:"push_url" gorm:"size:500"`                     // 推送回调地址
	ReconciliationURL string    `json:"reconciliation_url" gorm:"size:500"`           // 对账接口URL
	QueryURL          string    `json:"query_url" gorm:"size:500"`                    // 查询接口URL
	WhiteIPs          string    `json:"white_ips" gorm:"size:500"`                    // IP白名单（逗号分隔）
	Status            string    `json:"status" gorm:"size:20;default:active;index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (InterconnectOrg) TableName() string {
	return "interconnect_orgs"
}

// KeyType 密钥类型
const (
	KeyTypeOrg    = "org_key"    // 机构密钥
	KeyTypeSign   = "sign_key"   // 签名密钥
	KeyTypeMsg    = "msg_key"    // 消息密钥
)

// InterconnectKey 互联密钥
type InterconnectKey struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	OrgID      string    `json:"org_id" gorm:"index;size:36;not null"`
	KeyType    string    `json:"key_type" gorm:"size:20;not null"`          // org_key/sign_key/msg_key
	PublicKey  string    `json:"public_key" gorm:"type:text"`               // 公钥
	PrivateKey string    `json:"private_key" gorm:"type:text"`              // 私钥（加密存储）
	SecretKey  string    `json:"secret_key" gorm:"size:500"`                // 加密密钥/AppSecret
	Remark     string    `json:"remark" gorm:"size:255"`
	Status     string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (InterconnectKey) TableName() string {
	return "interconnect_keys"
}
