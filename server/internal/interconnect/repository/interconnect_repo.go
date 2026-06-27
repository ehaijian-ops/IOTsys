package repository

import (
	"iot-platform/internal/interconnect/model"

	"gorm.io/gorm"
)

// InterconnectRepository 互联互通数据访问层
type InterconnectRepository struct {
	db *gorm.DB
}

func NewInterconnectRepository(db *gorm.DB) *InterconnectRepository {
	return &InterconnectRepository{db: db}
}

// ========== InterconnectOrg ==========

func (r *InterconnectRepository) CreateOrg(org *model.InterconnectOrg) error {
	return r.db.Create(org).Error
}

func (r *InterconnectRepository) GetOrg(id string) (*model.InterconnectOrg, error) {
	var o model.InterconnectOrg
	err := r.db.First(&o, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *InterconnectRepository) ListOrgs(page, pageSize int) ([]model.InterconnectOrg, int64, error) {
	var orgs []model.InterconnectOrg
	var total int64
	if err := r.db.Model(&model.InterconnectOrg{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orgs).Error
	return orgs, total, err
}

func (r *InterconnectRepository) UpdateOrg(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.InterconnectOrg{}).Where("id = ?", id).Updates(updates).Error
}

func (r *InterconnectRepository) DeleteOrg(id string) error {
	return r.db.Delete(&model.InterconnectOrg{}, "id = ?", id).Error
}

// ========== InterconnectKey ==========

func (r *InterconnectRepository) CreateKey(key *model.InterconnectKey) error {
	return r.db.Create(key).Error
}

func (r *InterconnectRepository) GetKey(id string) (*model.InterconnectKey, error) {
	var k model.InterconnectKey
	err := r.db.First(&k, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *InterconnectRepository) ListKeys(orgID string, page, pageSize int) ([]model.InterconnectKey, int64, error) {
	var keys []model.InterconnectKey
	var total int64
	q := r.db.Model(&model.InterconnectKey{})
	if orgID != "" {
		q = q.Where("org_id = ?", orgID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error
	return keys, total, err
}

func (r *InterconnectRepository) UpdateKey(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.InterconnectKey{}).Where("id = ?", id).Updates(updates).Error
}

func (r *InterconnectRepository) DeleteKey(id string) error {
	return r.db.Delete(&model.InterconnectKey{}, "id = ?", id).Error
}
