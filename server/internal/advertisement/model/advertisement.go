package model

import "time"

// Advertisement 广告轮播图
type Advertisement struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Title     string    `json:"title" gorm:"size:100;not null"`
	ImageURL  string    `json:"image_url" gorm:"size:500;not null"`   // 图片地址
	LinkURL   string    `json:"link_url" gorm:"size:500"`             // 跳转链接
	SortOrder int       `json:"sort_order" gorm:"default:0"`          // 排序
	Platform  string    `json:"platform" gorm:"size:20;default:mini"`  // mini/pc/all
	Status    string    `json:"status" gorm:"size:20;default:active;index"` // active/inactive
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Advertisement) TableName() string {
	return "advertisements"
}

// FranchiseApplication 加盟合作申请
type FranchiseApplication struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	Phone     string    `json:"phone" gorm:"size:20;not null"`
	Company   string    `json:"company" gorm:"size:100"`
	Address   string    `json:"address" gorm:"size:255"`
	Remark    string    `json:"remark" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20;default:pending;index"` // pending/approved/rejected
	ProcessedBy string  `json:"processed_by" gorm:"size:64"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FranchiseApplication) TableName() string {
	return "franchise_applications"
}

// WechatUser 微信小程序用户
type WechatUser struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	OpenID    string    `json:"open_id" gorm:"uniqueIndex;size:64;not null"`
	UnionID   string    `json:"union_id" gorm:"index;size:64"`
	Nickname  string    `json:"nickname" gorm:"size:64"`
	AvatarURL string    `json:"avatar_url" gorm:"size:500"`
	Phone     string    `json:"phone" gorm:"size:20;index"`
	Gender    int       `json:"gender" gorm:"default:0"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WechatUser) TableName() string {
	return "wechat_users"
}
