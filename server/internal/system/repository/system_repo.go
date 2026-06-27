package repository

import (
	"iot-platform/internal/system/model"

	"gorm.io/gorm"
)

// SystemRepository 系统管理数据访问层
type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// ========== Role ==========

func (r *SystemRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *SystemRepository) GetRole(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SystemRepository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SystemRepository) ListRoles(page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	if err := r.db.Model(&model.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&roles).Error
	return roles, total, err
}

func (r *SystemRepository) UpdateRole(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Role{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteRole(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

// ========== Menu ==========

func (r *SystemRepository) CreateMenu(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

func (r *SystemRepository) GetMenu(id uint) (*model.Menu, error) {
	var m model.Menu
	err := r.db.First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SystemRepository) ListAllMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Order("sort_order ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *SystemRepository) UpdateMenu(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Menu{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SystemRepository) DeleteMenu(id uint) error {
	return r.db.Delete(&model.Menu{}, id).Error
}

// ========== LoginLog ==========

func (r *SystemRepository) CreateLoginLog(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) ListLoginLogs(page, pageSize int) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64
	if err := r.db.Model(&model.LoginLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// ========== SystemLog ==========

func (r *SystemRepository) CreateSystemLog(log *model.SystemLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) ListSystemLogs(module, action string, page, pageSize int) ([]model.SystemLog, int64, error) {
	var logs []model.SystemLog
	var total int64
	q := r.db.Model(&model.SystemLog{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
