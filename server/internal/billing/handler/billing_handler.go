package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/billing/model"
	"iot-platform/internal/billing/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// BillingHandler 收费方案HTTP处理器
type BillingHandler struct {
	svc *service.BillingService
}

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// ========== BillingScheme ==========

func (h *BillingHandler) CreateScheme(c *gin.Context) {
	var req struct {
		Name           string  `json:"name" binding:"required"`
		Type           string  `json:"type" binding:"required"`
		DeviceType     string  `json:"device_type"`
		SiteID         string  `json:"site_id"`
		BaseServiceFee float64 `json:"base_service_fee"`
		MaxPrice       float64 `json:"max_price"`
		UnitPrice      float64 `json:"unit_price"`
		Unit           string  `json:"unit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的收费方案信息"))
		return
	}

	scheme, err := h.svc.CreateScheme(req.Name, req.Type, req.DeviceType, req.SiteID,
		req.BaseServiceFee, req.MaxPrice, req.UnitPrice, req.Unit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "收费方案创建成功",
		"data":    scheme,
	})
}

func (h *BillingHandler) GetScheme(c *gin.Context) {
	id := c.Param("id")
	scheme, err := h.svc.GetScheme(id)
	if err != nil {
		c.Error(err)
		return
	}

	// 获取时段配置
	periods, _ := h.svc.GetPeriods(id)

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"data":    scheme,
		"periods": periods,
	})
}

func (h *BillingHandler) ListSchemes(c *gin.Context) {
	deviceType := c.Query("device_type")
	siteID := c.Query("site_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	schemes, total, err := h.svc.ListSchemes(deviceType, siteID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      schemes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *BillingHandler) UpdateScheme(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}

	scheme, err := h.svc.UpdateScheme(id, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "收费方案更新成功",
		"data":    scheme,
	})
}

func (h *BillingHandler) DeleteScheme(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteScheme(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "收费方案已删除"})
}

// ========== BillingPeriod ==========

func (h *BillingHandler) BatchSetPeriods(c *gin.Context) {
	schemeID := c.Param("id")

	var req struct {
		Periods []model.BillingPeriod `json:"periods" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供时段配置"))
		return
	}

	if err := h.svc.BatchSetPeriods(schemeID, req.Periods); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "时段配置保存成功"})
}

func (h *BillingHandler) GetPeriods(c *gin.Context) {
	schemeID := c.Param("id")
	periods, err := h.svc.GetPeriods(schemeID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": periods})
}

// ========== MonthlyCardScheme ==========

func (h *BillingHandler) CreateMonthlyScheme(c *gin.Context) {
	var req struct {
		Name         string  `json:"name" binding:"required"`
		DeviceType   string  `json:"device_type" binding:"required"`
		Price        float64 `json:"price" binding:"required"`
		DurationDays int     `json:"duration_days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的月卡方案信息"))
		return
	}

	scheme, err := h.svc.CreateMonthlyScheme(req.Name, req.DeviceType, req.Price, req.DurationDays)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "月卡方案创建成功",
		"data":    scheme,
	})
}

func (h *BillingHandler) ListMonthlySchemes(c *gin.Context) {
	deviceType := c.Query("device_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	schemes, total, err := h.svc.ListMonthlySchemes(deviceType, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      schemes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *BillingHandler) UpdateMonthlyScheme(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}

	scheme, err := h.svc.UpdateMonthlyScheme(id, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "月卡方案更新成功", "data": scheme})
}

func (h *BillingHandler) DeleteMonthlyScheme(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteMonthlyScheme(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "月卡方案已删除"})
}

// ========== RechargeScheme ==========

func (h *BillingHandler) CreateRechargeScheme(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		BonusAmount float64 `json:"bonus_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的充值方案信息"))
		return
	}

	scheme, err := h.svc.CreateRechargeScheme(req.Name, req.Amount, req.BonusAmount)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "充值方案创建成功",
		"data":    scheme,
	})
}

func (h *BillingHandler) ListRechargeSchemes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	schemes, total, err := h.svc.ListRechargeSchemes(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      schemes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *BillingHandler) UpdateRechargeScheme(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}

	scheme, err := h.svc.UpdateRechargeScheme(id, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "充值方案更新成功", "data": scheme})
}

func (h *BillingHandler) DeleteRechargeScheme(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteRechargeScheme(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "充值方案已删除"})
}

// ========== BusinessConfig ==========

func (h *BillingHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	cfg, err := h.svc.GetConfig(key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": cfg})
}

func (h *BillingHandler) SetConfig(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供配置值"))
		return
	}

	if err := h.svc.SetConfig(key, req.Value, req.Description); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "配置保存成功"})
}

func (h *BillingHandler) ListConfigs(c *gin.Context) {
	configs, err := h.svc.ListConfigs()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": configs})
}
