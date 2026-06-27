package service

import (
	"fmt"
	"math/rand"
	"time"

	"iot-platform/internal/order/model"
	"iot-platform/internal/order/repository"
	errs "iot-platform/pkg/errors"

	"github.com/google/uuid"
)

// OrderService 订单业务逻辑层
type OrderService struct {
	orderRepo  *repository.OrderRepository
	recordRepo *repository.ChargeRecordRepository
	refundRepo *repository.OrderRefundRepository
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	recordRepo *repository.ChargeRecordRepository,
	refundRepo *repository.OrderRefundRepository,
) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		recordRepo: recordRepo,
		refundRepo: refundRepo,
	}
}

// GenerateOrderSN 生成订单号
func GenerateOrderSN() string {
	now := time.Now()
	return fmt.Sprintf("CD%s%s%04d", now.Format("20060102150405"), now.Format("000000")[0:5], rand.Intn(10000))
}

// CreateOrder 创建充电订单
func (s *OrderService) CreateOrder(deviceID string, orderType string, userID *uint, opts map[string]interface{}) (*model.ChargeOrder, error) {
	order := &model.ChargeOrder{
		ID:       uuid.New().String(),
		OrderSN:  GenerateOrderSN(),
		DeviceID: deviceID,
		UserID:   userID,
		OrderType: orderType,
		Status:   model.OrderStatusPending,
	}

	// 可选参数
	if v, ok := opts["port_number"].(int); ok {
		order.PortNumber = v
	}
	if v, ok := opts["charge_type"].(string); ok {
		order.ChargeType = v
	}
	if v, ok := opts["billing_scheme_id"].(string); ok {
		order.BillingSchemeID = v
	}
	if v, ok := opts["pay_method"].(string); ok {
		order.PayMethod = v
	}
	if v, ok := opts["plate_number"].(string); ok {
		order.PlateNumber = v
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return order, nil
}

// StartCharging 开始充电
func (s *OrderService) StartCharging(orderID string) error {
	now := time.Now()
	return s.orderRepo.UpdateFields(orderID, map[string]interface{}{
		"status":     model.OrderStatusCharging,
		"start_time": now,
	})
}

// AddChargeRecord 添加充电记录
func (s *OrderService) AddChargeRecord(record *model.ChargeRecord) error {
	return s.recordRepo.Create(record)
}

// EndOrder 结束订单
func (s *OrderService) EndOrder(orderID string, amount float64, energyKWh float64) error {
	now := time.Now()

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return errs.NotFound("订单", orderID)
	}
	if order.Status != model.OrderStatusCharging {
		return errs.New("ORDER_STATUS_ERROR", "订单状态不正确", 400)
	}

	duration := int64(now.Sub(*order.StartTime).Seconds())

	return s.orderRepo.UpdateFields(orderID, map[string]interface{}{
		"status":     model.OrderStatusCompleted,
		"end_time":   now,
		"duration":   duration,
		"energy_kwh": energyKWh,
		"amount":     amount,
		"paid_amount": amount,
	})
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(orderID string) (*model.ChargeOrder, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, errs.NotFound("订单", orderID)
	}
	return order, nil
}

// ListOrders 订单列表
func (s *OrderService) ListOrders(query *model.OrderQuery) ([]model.ChargeOrder, int64, error) {
	orders, total, err := s.orderRepo.List(query)
	if err != nil {
		return nil, 0, errs.ErrInternalServer.Wrap(err)
	}
	return orders, total, nil
}

// GetChargeCurve 获取充电功率曲线
func (s *OrderService) GetChargeCurve(orderID string) ([]model.ChargeRecord, error) {
	// 确认订单存在
	if _, err := s.orderRepo.GetByID(orderID); err != nil {
		return nil, errs.NotFound("订单", orderID)
	}
	return s.recordRepo.GetByOrderID(orderID)
}

// RefundOrder 退款
func (s *OrderService) RefundOrder(orderID string, amount float64, reason string, processedBy string) (*model.OrderRefund, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, errs.NotFound("订单", orderID)
	}
	if order.Status != model.OrderStatusCompleted {
		return nil, errs.New("ORDER_CANNOT_REFUND", "仅已完成的订单可退款", 400)
	}
	// 检查是否已退款
	existRefund, _ := s.refundRepo.GetByOrderID(orderID)
	if existRefund != nil {
		return nil, errs.New("ORDER_ALREADY_REFUNDED", "订单已退款", 400)
	}

	refund := &model.OrderRefund{
		ID:           uuid.New().String(),
		OrderID:      orderID,
		RefundAmount: amount,
		RefundReason: reason,
		RefundMethod: "wallet",
		Status:       "pending",
		ProcessedBy:  processedBy,
	}

	if err := s.refundRepo.Create(refund); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	// 更新订单退款金额
	if err := s.orderRepo.UpdateFields(orderID, map[string]interface{}{
		"refund_amount": refund.RefundAmount,
	}); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	return refund, nil
}

// ProcessRefund 处理退款（打款/驳回）
func (s *OrderService) ProcessRefund(refundID string, status string, tradeNo string) error {
	refund, err := s.refundRepo.GetByID(refundID)
	if err != nil {
		return errs.NotFound("退款记录", refundID)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": now,
	}
	if tradeNo != "" {
		updates["trade_no"] = tradeNo
	}

	if err := s.refundRepo.Update(refund); err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}
	_ = updates

	// 根据退款结果更新订单状态
	orderStatus := model.OrderStatusCompleted
	if status == "success" {
		orderStatus = model.OrderStatusRefunded
	}
	return s.orderRepo.UpdateFields(refund.OrderID, map[string]interface{}{
		"status": orderStatus,
	})
}

// DeleteOrder 删除订单
func (s *OrderService) DeleteOrder(orderID string) error {
	if _, err := s.orderRepo.GetByID(orderID); err != nil {
		return errs.NotFound("订单", orderID)
	}
	return s.orderRepo.Delete(orderID)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderID string) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return errs.NotFound("订单", orderID)
	}
	if order.Status != model.OrderStatusPending {
		return errs.New("ORDER_CANNOT_CANCEL", "仅待开始的订单可取消", 400)
	}
	return s.orderRepo.UpdateFields(orderID, map[string]interface{}{
		"status": model.OrderStatusCancelled,
	})
}

// ========== 统计方法 ==========

// GetOrderStats 获取订单统计（用于仪表盘）
func (s *OrderService) GetOrderStats() (map[string]interface{}, error) {
	totalOrders, chargingOrders, completedOrders, totalAmount, err := s.orderRepo.GetStats()
	if err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	return map[string]interface{}{
		"total_orders":     totalOrders,
		"charging_orders":  chargingOrders,
		"completed_orders": completedOrders,
		"total_amount":     totalAmount,
	}, nil
}
