package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色定义
const (
	RoleSuperAdmin = "super_admin" // 超级管理员：所有权限 + 管理其他管理员
	RoleAdmin      = "admin"       // 管理员：设备管理、指令下发、查看报表
	RoleOperator   = "operator"    // 运维人员：查看设备、下发指令
	RoleViewer     = "viewer"      // 查看者：仅查看权限
)

// AllRoles 所有角色列表
var AllRoles = []string{RoleSuperAdmin, RoleAdmin, RoleOperator, RoleViewer}

// User 后台管理用户
type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password     string         `gorm:"type:varchar(256);not null" json:"-"` // json:"-" 禁止序列化到前端
	Nickname     string         `gorm:"type:varchar(64)" json:"nickname"`
	Role         string         `gorm:"type:varchar(32);not null;default:viewer;index" json:"role"`
	Email        string         `gorm:"type:varchar(128)" json:"email,omitempty"`
	Phone        string         `gorm:"type:varchar(32)" json:"phone,omitempty"`
	Avatar       string         `gorm:"type:varchar(256)" json:"avatar,omitempty"`
	Enabled      bool           `gorm:"default:true;index" json:"enabled"` // 是否启用
	LastLoginAt  *time.Time     `gorm:"" json:"last_login_at,omitempty"`
	LastLoginIP  string         `gorm:"type:varchar(64)" json:"last_login_ip,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // 秒
	User         *User  `json:"user"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"required,min=1,max=64"`
	Role     string `json:"role" binding:"required,oneof=super_admin admin operator viewer"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname *string `json:"nickname,omitempty"`
	Role     *string `json:"role,omitempty" binding:"omitempty,oneof=super_admin admin operator viewer"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=64"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// ResetPasswordRequest 管理员重置密码
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// UserListItem 用户列表项（不含敏感信息）
type UserListItem struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Role        string     `json:"role"`
	Email       string     `json:"email,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	Enabled     bool       `json:"enabled"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToListItem 转为列表项
func (u *User) ToListItem() *UserListItem {
	return &UserListItem{
		ID:          u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Role:        u.Role,
		Email:       u.Email,
		Phone:       u.Phone,
		Enabled:     u.Enabled,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}
