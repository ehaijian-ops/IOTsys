package repository

import (
	"iot-platform/internal/advertisement/model"

	"gorm.io/gorm"
)

// AdvertisementRepository 广告/用户数据访问层
type AdvertisementRepository struct {
	db *gorm.DB
}

func NewAdvertisementRepository(db *gorm.DB) *AdvertisementRepository {
	return &AdvertisementRepository{db: db}
}

// ========== Advertisement ==========

func (r *AdvertisementRepository) Create(ad *model.Advertisement) error {
	return r.db.Create(ad).Error
}

func (r *AdvertisementRepository) GetByID(id string) (*model.Advertisement, error) {
	var a model.Advertisement
	err := r.db.First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdvertisementRepository) List(platform string, page, pageSize int) ([]model.Advertisement, int64, error) {
	var ads []model.Advertisement
	var total int64
	q := r.db.Model(&model.Advertisement{})
	if platform != "" {
		q = q.Where("platform = ? OR platform = 'all'", platform)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&ads).Error
	return ads, total, err
}

func (r *AdvertisementRepository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.Advertisement{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AdvertisementRepository) Delete(id string) error {
	return r.db.Delete(&model.Advertisement{}, "id = ?", id).Error
}

// ========== FranchiseApplication ==========

func (r *AdvertisementRepository) CreateFranchise(app *model.FranchiseApplication) error {
	return r.db.Create(app).Error
}

func (r *AdvertisementRepository) GetFranchise(id string) (*model.FranchiseApplication, error) {
	var a model.FranchiseApplication
	err := r.db.First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdvertisementRepository) ListFranchises(status string, page, pageSize int) ([]model.FranchiseApplication, int64, error) {
	var apps []model.FranchiseApplication
	var total int64
	q := r.db.Model(&model.FranchiseApplication{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&apps).Error
	return apps, total, err
}

func (r *AdvertisementRepository) UpdateFranchise(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.FranchiseApplication{}).Where("id = ?", id).Updates(updates).Error
}

// ========== WechatUser ==========

func (r *AdvertisementRepository) CreateWechatUser(user *model.WechatUser) error {
	return r.db.Create(user).Error
}

func (r *AdvertisementRepository) GetWechatUserByOpenID(openID string) (*model.WechatUser, error) {
	var u model.WechatUser
	err := r.db.Where("open_id = ?", openID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdvertisementRepository) GetWechatUserByID(id uint) (*model.WechatUser, error) {
	var u model.WechatUser
	err := r.db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdvertisementRepository) ListWechatUsers(page, pageSize int) ([]model.WechatUser, int64, error) {
	var users []model.WechatUser
	var total int64
	if err := r.db.Model(&model.WechatUser{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *AdvertisementRepository) UpdateWechatUser(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.WechatUser{}).Where("id = ?", id).Updates(updates).Error
}
