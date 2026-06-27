package repository

import (
	"iot-platform/internal/card/model"

	"gorm.io/gorm"
)

// CardRepository 卡片数据访问层
type CardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{db: db}
}

// ========== IC Card ==========

func (r *CardRepository) CreateICCard(card *model.ICCard) error {
	return r.db.Create(card).Error
}

func (r *CardRepository) GetICCard(id string) (*model.ICCard, error) {
	var c model.ICCard
	err := r.db.First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CardRepository) GetICCardByNo(cardNo string) (*model.ICCard, error) {
	var c model.ICCard
	err := r.db.Where("card_no = ?", cardNo).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CardRepository) ListICCards(page, pageSize int) ([]model.ICCard, int64, error) {
	var cards []model.ICCard
	var total int64
	if err := r.db.Model(&model.ICCard{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&cards).Error
	return cards, total, err
}

func (r *CardRepository) UpdateICCard(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.ICCard{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CardRepository) DeleteICCard(id string) error {
	return r.db.Delete(&model.ICCard{}, "id = ?", id).Error
}

func (r *CardRepository) BatchCreateICCards(cards []*model.ICCard) error {
	return r.db.Create(&cards).Error
}

// ICCardRecharge
func (r *CardRepository) CreateICCardRecharge(record *model.ICCardRecharge) error {
	return r.db.Create(record).Error
}

// ========== Traffic Card ==========

func (r *CardRepository) CreateTrafficCard(card *model.TrafficCard) error {
	return r.db.Create(card).Error
}

func (r *CardRepository) GetTrafficCard(id string) (*model.TrafficCard, error) {
	var c model.TrafficCard
	err := r.db.First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CardRepository) ListTrafficCards(page, pageSize int) ([]model.TrafficCard, int64, error) {
	var cards []model.TrafficCard
	var total int64
	if err := r.db.Model(&model.TrafficCard{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&cards).Error
	return cards, total, err
}

func (r *CardRepository) UpdateTrafficCard(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.TrafficCard{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CardRepository) DeleteTrafficCard(id string) error {
	return r.db.Delete(&model.TrafficCard{}, "id = ?", id).Error
}

// ========== Monthly Card ==========

func (r *CardRepository) CreateMonthlyCard(card *model.MonthlyCard) error {
	return r.db.Create(card).Error
}

func (r *CardRepository) GetMonthlyCard(id string) (*model.MonthlyCard, error) {
	var c model.MonthlyCard
	err := r.db.First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CardRepository) ListMonthlyCards(userID *uint, page, pageSize int) ([]model.MonthlyCard, int64, error) {
	var cards []model.MonthlyCard
	var total int64
	q := r.db.Model(&model.MonthlyCard{})
	if userID != nil && *userID > 0 {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&cards).Error
	return cards, total, err
}

func (r *CardRepository) UpdateMonthlyCard(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.MonthlyCard{}).Where("id = ?", id).Updates(updates).Error
}

// MonthlyCardRecord
func (r *CardRepository) CreateMonthlyRecord(record *model.MonthlyCardRecord) error {
	return r.db.Create(record).Error
}
