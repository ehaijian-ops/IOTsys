package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/finance/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// FinanceHandler 财务HTTP处理器
type FinanceHandler struct {
	svc *service.FinanceService
}

func NewFinanceHandler(svc *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{svc: svc}
}

// GetWallet 获取钱包
func (h *FinanceHandler) GetWallet(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.Error(errs.ValidationError("请提供用户ID"))
		return
	}
	uid, _ := strconv.ParseUint(userID, 10, 64)
	wallet, err := h.svc.GetWallet(uint(uid))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": wallet})
}

// Recharge 用户充值
func (h *FinanceHandler) Recharge(c *gin.Context) {
	var req struct {
		UserID      uint    `json:"user_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		BonusAmount float64 `json:"bonus_amount"`
		PayMethod   string  `json:"pay_method" binding:"required"`
		TradeNo     string  `json:"trade_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的充值信息"))
		return
	}

	record, err := h.svc.Recharge(req.UserID, req.Amount, req.BonusAmount, req.PayMethod, req.TradeNo)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "充值成功",
		"data":    record,
	})
}

// AdminRecharge 管理员充值
func (h *FinanceHandler) AdminRecharge(c *gin.Context) {
	var req struct {
		UserID uint    `json:"user_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
		Remark string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的充值信息"))
		return
	}

	record, err := h.svc.AdminRecharge(req.UserID, req.Amount, req.Remark)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "充值成功",
		"data":    record,
	})
}

// ListRecharges 充值记录列表
func (h *FinanceHandler) ListRecharges(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	var userID *uint
	if uid := c.Query("user_id"); uid != "" {
		id, _ := strconv.ParseUint(uid, 10, 64)
		u := uint(id)
		userID = &u
	}

	records, total, err := h.svc.ListRecharges(userID, status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ApplyWithdraw 申请提现
func (h *FinanceHandler) ApplyWithdraw(c *gin.Context) {
	var req struct {
		Amount      float64 `json:"amount" binding:"required"`
		BankName    string  `json:"bank_name"`
		BankCardNo  string  `json:"bank_card_no"`
		BankAccount string  `json:"bank_account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的提现信息"))
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := strconv.ParseUint(userID.(string), 10, 64)

	record, err := h.svc.ApplyWithdraw(uint(uid), req.Amount, req.BankName, req.BankCardNo, req.BankAccount)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "提现申请已提交",
		"data":    record,
	})
}

// ListWithdraws 提现列表
func (h *FinanceHandler) ListWithdraws(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	records, total, err := h.svc.ListWithdraws(status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ProcessWithdraw 处理提现
func (h *FinanceHandler) ProcessWithdraw(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status       string  `json:"status" binding:"required"`
		ActualAmount float64 `json:"actual_amount"`
		Remark       string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供处理信息"))
		return
	}

	processedBy, _ := c.Get("username")
	operator := ""
	if name, ok := processedBy.(string); ok {
		operator = name
	}

	if err := h.svc.ProcessWithdraw(id, req.Status, req.ActualAmount, req.Remark, operator); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "提现处理成功"})
}

// ListSplits 分成列表
func (h *FinanceHandler) ListSplits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	splits, total, err := h.svc.ListSplits(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      splits,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
