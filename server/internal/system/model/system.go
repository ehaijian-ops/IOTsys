package model

import "time"

// Role 角色（扩展RBAC）
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:50;not null"`
	DisplayName string    `json:"display_name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:255"`
	Permissions string    `json:"permissions" gorm:"type:text"` // 权限列表（JSON数组）
	DataScope   string    `json:"data_scope" gorm:"size:20;default:all"` // all/site/self
	Status      string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

// Menu 菜单（动态菜单管理）
type Menu struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ParentID   *uint     `json:"parent_id" gorm:"index;default:null"`
	Name       string    `json:"name" gorm:"size:50;not null"`              // 路由名称
	Path       string    `json:"path" gorm:"size:255"`                      // 路由路径
	Component  string    `json:"component" gorm:"size:255"`                 // 组件路径
	Icon       string    `json:"icon" gorm:"size:50"`                       // 图标
	Title      string    `json:"title" gorm:"size:50;not null"`             // 菜单标题
	Permission string    `json:"permission" gorm:"size:100"`                // 权限标识
	SortOrder  int       `json:"sort_order" gorm:"default:0"`
	Hidden     bool      `json:"hidden" gorm:"default:false"`
	Status     string    `json:"status" gorm:"size:20;default:active"`
	Children   []*Menu   `json:"children,omitempty" gorm:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Menu) TableName() string {
	return "menus"
}

// LoginLog 登陆日志
type LoginLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Username  string    `json:"username" gorm:"size:64"`
	IP        string    `json:"ip" gorm:"size:64"`
	UserAgent string    `json:"user_agent" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20"` // success/failed
	Message   string    `json:"message" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}

// SystemLog 系统操作日志
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Username  string    `json:"username" gorm:"size:64;index"`
	Action    string    `json:"action" gorm:"size:50;index"`          // create/update/delete/login/logout
	Module    string    `json:"module" gorm:"size:50;index"`          // device/order/user/system
	Target    string    `json:"target" gorm:"size:100"`               // 操作目标
	Detail    string    `json:"detail" gorm:"type:text"`              // 操作详情（JSON）
	IP        string    `json:"ip" gorm:"size:64"`
	CreatedAt time.Time `json:"created_at"`
}

func (SystemLog) TableName() string {
	return "system_logs"
}
