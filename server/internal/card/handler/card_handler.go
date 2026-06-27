package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/card/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// CardHandler 卡片管理HTTP处理器
type CardHandler struct {
	svc *service.CardService
}

func NewCardHandler(svc *service.CardService) *CardHandler {
	return &CardHandler{svc: svc}
}

// ========== IC Card ==========

func (h *CardHandler) CreateICCard(c *gin.Context) {
	var req struct {
		CardNo  string `json:"card_no" binding:"required"`
		CardUID string `json:"card_uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供卡号"))
		return
	}
	card, err := h.svc.CreateICCard(req.CardNo, req.CardUID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "IC卡创建成功", "data": card})
}

func (h *CardHandler) ListICCards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	cards, total, err := h.svc.ListICCards(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": cards, "total": total, "page": page, "page_size": pageSize})
}

func (h *CardHandler) GetICCard(c *gin.Context) {
	id := c.Param("id")
	card, err := h.svc.GetICCard(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": card})
}

func (h *CardHandler) RechargeICCard(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供充值金额"))
		return
	}
	createdBy, _ := c.Get("username")
	operator := ""
	if name, ok := createdBy.(string); ok {
		operator = name
	}
	if err := h.svc.RechargeICCard(id, req.Amount, operator); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "IC卡充值成功"})
}

func (h *CardHandler) BindICCard(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供用户ID"))
		return
	}
	if err := h.svc.BindICCardToUser(id, req.UserID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "IC卡绑定成功"})
}

func (h *CardHandler) ReportLostICCard(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.ReportLostICCard(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "IC卡已挂失"})
}

func (h *CardHandler) DeleteICCard(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteICCard(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "IC卡已删除"})
}

func (h *CardHandler) BatchImportICCards(c *gin.Context) {
	var req struct {
		CardNos []string `json:"card_nos" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供卡号列表"))
		return
	}
	count, err := h.svc.BatchImportICCards(req.CardNos)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "批量导入完成", "data": gin.H{"count": count}})
}

// ========== Traffic Card ==========

func (h *CardHandler) CreateTrafficCard(c *gin.Context) {
	var req struct {
		ICCID   string `json:"iccid" binding:"required"`
		IMSI    string `json:"imsi"`
		Carrier string `json:"carrier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供ICCID"))
		return
	}
	card, err := h.svc.CreateTrafficCard(req.ICCID, req.IMSI, req.Carrier)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "流量卡创建成功", "data": card})
}

func (h *CardHandler) ListTrafficCards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	cards, total, err := h.svc.ListTrafficCards(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": cards, "total": total, "page": page, "page_size": pageSize})
}

func (h *CardHandler) BindTrafficCard(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供设备ID"))
		return
	}
	if err := h.svc.BindTrafficCardToDevice(id, req.DeviceID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "流量卡绑定成功"})
}

// ========== Monthly Card ==========

func (h *CardHandler) ListMonthlyCards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var userID *uint
	if uid := c.Query("user_id"); uid != "" {
		id, _ := strconv.ParseUint(uid, 10, 64)
		u := uint(id)
		userID = &u
	}
	cards, total, err := h.svc.ListMonthlyCards(userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": cards, "total": total, "page": page, "page_size": pageSize})
}
