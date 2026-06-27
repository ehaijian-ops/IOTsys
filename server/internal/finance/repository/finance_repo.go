package repository

import (
	"iot-platform/internal/finance/model"

	"gorm.io/gorm"
)

// FinanceRepository 财务数据访问层
type FinanceRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

// ========== UserWallet ==========

func (r *FinanceRepository) GetWallet(userID uint) (*model.UserWallet, error) {
	var w model.UserWallet
	err := r.db.Where("user_id = ?", userID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *FinanceRepository) CreateWallet(wallet *model.UserWallet) error {
	return r.db.Create(wallet).Error
}

func (r *FinanceRepository) UpdateWalletBalance(userID uint, amount float64) error {
	return r.db.Model(&model.UserWallet{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"balance":       gorm.Expr("balance + ?", amount),
			"total_recharge": gorm.Expr("total_recharge + ?", amount),
		}).Error
}

func (r *FinanceRepository) DeductWalletBalance(userID uint, amount float64) error {
	return r.db.Model(&model.UserWallet{}).Where("user_id = ? AND balance >= ?", userID, amount).
		Updates(map[string]interface{}{
			"balance":       gorm.Expr("balance - ?", amount),
			"total_consume": gorm.Expr("total_consume + ?", amount),
		}).Error
}

func (r *FinanceRepository) FreezeBalance(userID uint, amount float64) error {
	return r.db.Model(&model.UserWallet{}).Where("user_id = ? AND balance >= ?", userID, amount).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", amount),
			"frozen_amount":  gorm.Expr("frozen_amount + ?", amount),
		}).Error
}

func (r *FinanceRepository) UnfreezeBalance(userID uint, amount float64) error {
	return r.db.Model(&model.UserWallet{}).Where("user_id = ? AND frozen_amount >= ?", userID, amount).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance + ?", amount),
			"frozen_amount":  gorm.Expr("frozen_amount - ?", amount),
		}).Error
}

// ========== RechargeRecord ==========

func (r *FinanceRepository) CreateRecharge(record *model.RechargeRecord) error {
	return r.db.Create(record).Error
}

func (r *FinanceRepository) GetRecharge(id string) (*model.RechargeRecord, error) {
	var rec model.RechargeRecord
	err := r.db.First(&rec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *FinanceRepository) ListRecharges(userID *uint, status string, page, pageSize int) ([]model.RechargeRecord, int64, error) {
	var records []model.RechargeRecord
	var total int64

	q := r.db.Model(&model.RechargeRecord{})
	if userID != nil && *userID > 0 {
		q = q.Where("user_id = ?", *userID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (r *FinanceRepository) UpdateRecharge(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.RechargeRecord{}).Where("id = ?", id).Updates(updates).Error
}

// ========== WithdrawRecord ==========

func (r *FinanceRepository) CreateWithdraw(record *model.WithdrawRecord) error {
	return r.db.Create(record).Error
}

func (r *FinanceRepository) GetWithdraw(id string) (*model.WithdrawRecord, error) {
	var rec model.WithdrawRecord
	err := r.db.First(&rec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *FinanceRepository) ListWithdraws(status string, page, pageSize int) ([]model.WithdrawRecord, int64, error) {
	var records []model.WithdrawRecord
	var total int64

	q := r.db.Model(&model.WithdrawRecord{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (r *FinanceRepository) UpdateWithdraw(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.WithdrawRecord{}).Where("id = ?", id).Updates(updates).Error
}

// ========== RevenueSplit ==========

func (r *FinanceRepository) CreateSplit(split *model.RevenueSplit) error {
	return r.db.Create(split).Error
}

func (r *FinanceRepository) GetSplit(id string) (*model.RevenueSplit, error) {
	var s model.RevenueSplit
	err := r.db.First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *FinanceRepository) ListSplits(page, pageSize int) ([]model.RevenueSplit, int64, error) {
	var splits []model.RevenueSplit
	var total int64

	if err := r.db.Model(&model.RevenueSplit{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&splits).Error
	return splits, total, err
}

func (r *FinanceRepository) UpdateSplit(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.RevenueSplit{}).Where("id = ?", id).Updates(updates).Error
}
