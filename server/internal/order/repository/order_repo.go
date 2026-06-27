package repository

import (
	"iot-platform/internal/order/model"

	"gorm.io/gorm"
)

// OrderRepository 订单数据访问层
type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.ChargeOrder) error {
	return r.db.Create(order).Error
}

// GetByID 根据ID查找
func (r *OrderRepository) GetByID(id string) (*model.ChargeOrder, error) {
	var order model.ChargeOrder
	err := r.db.First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOrderSN 根据订单号查找
func (r *OrderRepository) GetByOrderSN(sn string) (*model.ChargeOrder, error) {
	var order model.ChargeOrder
	err := r.db.Where("order_sn = ?", sn).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// List 分页列表（支持多条件筛选）
func (r *OrderRepository) List(query *model.OrderQuery) ([]model.ChargeOrder, int64, error) {
	var orders []model.ChargeOrder
	var total int64

	q := r.db.Model(&model.ChargeOrder{})

	if query.OrderSN != "" {
		q = q.Where("order_sn LIKE ?", "%"+query.OrderSN+"%")
	}
	if query.OrderType != "" {
		q = q.Where("order_type = ?", query.OrderType)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.DeviceID != "" {
		q = q.Where("device_id = ?", query.DeviceID)
	}
	if query.UserID != nil && *query.UserID > 0 {
		q = q.Where("user_id = ?", *query.UserID)
	}
	if query.StartDate != "" {
		q = q.Where("created_at >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		q = q.Where("created_at <= ?", query.EndDate+" 23:59:59")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	err := q.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&orders).Error
	return orders, total, err
}

// GetStats 获取订单统计（用于仪表盘）
func (r *OrderRepository) GetStats() (totalOrders, chargingOrders, completedOrders int64, totalAmount float64, err error) {
	q := r.db.Model(&model.ChargeOrder{})
	if err = q.Count(&totalOrders).Error; err != nil {
		return
	}
	q.Where("status = ?", model.OrderStatusCharging).Count(&chargingOrders)
	q.Where("status = ?", model.OrderStatusCompleted).Count(&completedOrders)
	err = r.db.Model(&model.ChargeOrder{}).Where("status = ?", model.OrderStatusCompleted).
		Select("COALESCE(SUM(paid_amount),0)").Row().Scan(&totalAmount)
	return
}

// Update 全量更新
func (r *OrderRepository) Update(order *model.ChargeOrder) error {
	return r.db.Save(order).Error
}

// UpdateFields 按字段更新
func (r *OrderRepository) UpdateFields(id string, updates map[string]interface{}) error {
	return r.db.Model(&model.ChargeOrder{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 软删除（如果模型有软删除字段）
func (r *OrderRepository) Delete(id string) error {
	return r.db.Delete(&model.ChargeOrder{}, "id = ?", id).Error
}

// ========== ChargeRecord Repository ==========

// ChargeRecordRepository 充电记录数据访问层
type ChargeRecordRepository struct {
	db *gorm.DB
}

func NewChargeRecordRepository(db *gorm.DB) *ChargeRecordRepository {
	return &ChargeRecordRepository{db: db}
}

// Create 创建充电记录
func (r *ChargeRecordRepository) Create(record *model.ChargeRecord) error {
	return r.db.Create(record).Error
}

// BatchCreate 批量创建
func (r *ChargeRecordRepository) BatchCreate(records []*model.ChargeRecord) error {
	return r.db.Create(&records).Error
}

// GetByOrderID 根据订单ID获取充电曲线
func (r *ChargeRecordRepository) GetByOrderID(orderID string) ([]model.ChargeRecord, error) {
	var records []model.ChargeRecord
	err := r.db.Where("order_id = ?", orderID).Order("record_at ASC").Find(&records).Error
	return records, err
}

// ========== OrderRefund Repository ==========

// OrderRefundRepository 退款记录数据访问层
type OrderRefundRepository struct {
	db *gorm.DB
}

func NewOrderRefundRepository(db *gorm.DB) *OrderRefundRepository {
	return &OrderRefundRepository{db: db}
}

// Create 创建退款记录
func (r *OrderRefundRepository) Create(refund *model.OrderRefund) error {
	return r.db.Create(refund).Error
}

// GetByID 根据ID查找
func (r *OrderRefundRepository) GetByID(id string) (*model.OrderRefund, error) {
	var refund model.OrderRefund
	err := r.db.First(&refund, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// GetByOrderID 根据订单ID查找退款记录
func (r *OrderRefundRepository) GetByOrderID(orderID string) (*model.OrderRefund, error) {
	var refund model.OrderRefund
	err := r.db.Where("order_id = ?", orderID).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// List 退款列表
func (r *OrderRefundRepository) List(page, pageSize int) ([]model.OrderRefund, int64, error) {
	var refunds []model.OrderRefund
	var total int64

	if err := r.db.Model(&model.OrderRefund{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&refunds).Error
	return refunds, total, err
}

// Update 更新
func (r *OrderRefundRepository) Update(refund *model.OrderRefund) error {
	return r.db.Save(refund).Error
}
