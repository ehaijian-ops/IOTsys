package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/order/model"
	"iot-platform/internal/order/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单HTTP处理器
type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req struct {
		DeviceID        string `json:"device_id" binding:"required"`
		OrderType       string `json:"order_type" binding:"required"`
		UserID          *uint  `json:"user_id"`
		PortNumber      int    `json:"port_number"`
		ChargeType      string `json:"charge_type"`
		BillingSchemeID string `json:"billing_scheme_id"`
		PayMethod       string `json:"pay_method"`
		PlateNumber     string `json:"plate_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的订单信息"))
		return
	}

	opts := map[string]interface{}{
		"port_number":       req.PortNumber,
		"charge_type":       req.ChargeType,
		"billing_scheme_id": req.BillingSchemeID,
		"pay_method":        req.PayMethod,
		"plate_number":      req.PlateNumber,
	}

	order, err := h.svc.CreateOrder(req.DeviceID, req.OrderType, req.UserID, opts)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "订单创建成功",
		"data":    order,
	})
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.svc.GetOrder(id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "OK",
		"data": order,
	})
}

// ListOrders 订单列表
func (h *OrderHandler) ListOrders(c *gin.Context) {
	query := &model.OrderQuery{
		OrderSN:   c.Query("order_sn"),
		OrderType: c.Query("order_type"),
		Status:    c.Query("status"),
		DeviceID:  c.Query("device_id"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	if uid := c.Query("user_id"); uid != "" {
		id, _ := strconv.ParseUint(uid, 10, 64)
		u := uint(id)
		query.UserID = &u
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query.Page = page
	query.PageSize = pageSize

	orders, total, err := h.svc.ListOrders(query)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// EndOrder 结束订单
func (h *OrderHandler) EndOrder(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Amount    float64 `json:"amount"`
		EnergyKWh float64 `json:"energy_kwh"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供结算信息"))
		return
	}

	if err := h.svc.EndOrder(id, req.Amount, req.EnergyKWh); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "订单结束成功",
	})
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelOrder(id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "订单取消成功",
	})
}

// StartCharging 开始充电
func (h *OrderHandler) StartCharging(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StartCharging(id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "充电已开始",
	})
}

// GetChargeCurve 获取充电功率曲线
func (h *OrderHandler) GetChargeCurve(c *gin.Context) {
	id := c.Param("id")
	records, err := h.svc.GetChargeCurve(id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "OK",
		"data": records,
	})
}

// RefundOrder 退款
func (h *OrderHandler) RefundOrder(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Reason string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供退款金额"))
		return
	}

	processedBy, _ := c.Get("username")
	operator := ""
	if name, ok := processedBy.(string); ok {
		operator = name
	}

	refund, err := h.svc.RefundOrder(id, req.Amount, req.Reason, operator)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "退款申请成功",
		"data":    refund,
	})
}

// ProcessRefund 处理退款
func (h *OrderHandler) ProcessRefund(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status  string `json:"status" binding:"required"`
		TradeNo string `json:"trade_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供处理状态"))
		return
	}

	if err := h.svc.ProcessRefund(id, req.Status, req.TradeNo); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "退款处理成功",
	})
}

// DeleteOrder 删除订单
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteOrder(id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "订单删除成功",
	})
}
