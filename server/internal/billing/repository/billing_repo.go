package repository

import (
	"iot-platform/internal/billing/model"

	"gorm.io/gorm"
)

// BillingRepository 收费方案数据访问层
type BillingRepository struct {
	db *gorm.DB
}

func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

// ========== BillingScheme CRUD ==========

func (r *BillingRepository) CreateScheme(scheme *model.BillingScheme) error {
	return r.db.Create(scheme).Error
}

func (r *BillingRepository) GetScheme(id string) (*model.BillingScheme, error) {
	var s model.BillingScheme
	err := r.db.First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BillingRepository) ListSchemes(deviceType, siteID string, page, pageSize int) ([]model.BillingScheme, int64, error) {
	var schemes []model.BillingScheme
	var total int64

	q := r.db.Model(&model.BillingScheme{})
	if deviceType != "" {
		q = q.Where("device_type = ?", deviceType)
	}
	if siteID != "" {
		q = q.Where("site_id = ? OR site_id = ''", siteID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&schemes).Error
	return schemes, total, err
}

func (r *BillingRepository) UpdateScheme(scheme *model.BillingScheme) error {
	return r.db.Save(scheme).Error
}

func (r *BillingRepository) DeleteScheme(id string) error {
	return r.db.Delete(&model.BillingScheme{}, "id = ?", id).Error
}

// ========== BillingPeriod CRUD ==========

func (r *BillingRepository) CreatePeriod(period *model.BillingPeriod) error {
	return r.db.Create(period).Error
}

func (r *BillingRepository) GetPeriodsByScheme(schemeID string) ([]model.BillingPeriod, error) {
	var periods []model.BillingPeriod
	err := r.db.Where("scheme_id = ?", schemeID).Order("sort_order ASC, start_time ASC").Find(&periods).Error
	return periods, err
}

func (r *BillingRepository) UpdatePeriod(period *model.BillingPeriod) error {
	return r.db.Save(period).Error
}

func (r *BillingRepository) DeletePeriod(id uint) error {
	return r.db.Delete(&model.BillingPeriod{}, id).Error
}

func (r *BillingRepository) BatchDeletePeriods(schemeID string) error {
	return r.db.Where("scheme_id = ?", schemeID).Delete(&model.BillingPeriod{}).Error
}

// ========== MonthlyCardScheme CRUD ==========

func (r *BillingRepository) CreateMonthlyScheme(scheme *model.MonthlyCardScheme) error {
	return r.db.Create(scheme).Error
}

func (r *BillingRepository) GetMonthlyScheme(id string) (*model.MonthlyCardScheme, error) {
	var s model.MonthlyCardScheme
	err := r.db.First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BillingRepository) ListMonthlySchemes(deviceType string, page, pageSize int) ([]model.MonthlyCardScheme, int64, error) {
	var schemes []model.MonthlyCardScheme
	var total int64

	q := r.db.Model(&model.MonthlyCardScheme{})
	if deviceType != "" {
		q = q.Where("device_type = ?", deviceType)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&schemes).Error
	return schemes, total, err
}

func (r *BillingRepository) UpdateMonthlyScheme(scheme *model.MonthlyCardScheme) error {
	return r.db.Save(scheme).Error
}

func (r *BillingRepository) DeleteMonthlyScheme(id string) error {
	return r.db.Delete(&model.MonthlyCardScheme{}, "id = ?", id).Error
}

// ========== RechargeScheme CRUD ==========

func (r *BillingRepository) CreateRechargeScheme(scheme *model.RechargeScheme) error {
	return r.db.Create(scheme).Error
}

func (r *BillingRepository) GetRechargeScheme(id string) (*model.RechargeScheme, error) {
	var s model.RechargeScheme
	err := r.db.First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BillingRepository) ListRechargeSchemes(page, pageSize int) ([]model.RechargeScheme, int64, error) {
	var schemes []model.RechargeScheme
	var total int64

	if err := r.db.Model(&model.RechargeScheme{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&schemes).Error
	return schemes, total, err
}

func (r *BillingRepository) UpdateRechargeScheme(scheme *model.RechargeScheme) error {
	return r.db.Save(scheme).Error
}

func (r *BillingRepository) DeleteRechargeScheme(id string) error {
	return r.db.Delete(&model.RechargeScheme{}, "id = ?", id).Error
}

// ========== BusinessConfig CRUD ==========

func (r *BillingRepository) GetConfig(key string) (*model.BusinessConfig, error) {
	var cfg model.BusinessConfig
	err := r.db.Where("config_key = ?", key).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *BillingRepository) SetConfig(key, value, desc string) error {
	var cfg model.BusinessConfig
	result := r.db.Where("config_key = ?", key).First(&cfg)
	if result.Error != nil {
		return r.db.Create(&model.BusinessConfig{
			ConfigKey:   key,
			ConfigValue: value,
			Description: desc,
		}).Error
	}
	return r.db.Model(&cfg).Updates(map[string]interface{}{
		"config_value": value,
		"description":  desc,
	}).Error
}

func (r *BillingRepository) ListConfigs() ([]model.BusinessConfig, error) {
	var configs []model.BusinessConfig
	err := r.db.Order("config_key ASC").Find(&configs).Error
	return configs, err
}
