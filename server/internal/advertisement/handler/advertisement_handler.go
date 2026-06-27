package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/advertisement/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AdvertisementHandler 广告/运营HTTP处理器
type AdvertisementHandler struct {
	svc *service.AdvertisementService
}

func NewAdvertisementHandler(svc *service.AdvertisementService) *AdvertisementHandler {
	return &AdvertisementHandler{svc: svc}
}

// ========== Advertisement ==========

func (h *AdvertisementHandler) CreateAd(c *gin.Context) {
	var req struct {
		Title     string `json:"title" binding:"required"`
		ImageURL  string `json:"image_url" binding:"required"`
		LinkURL   string `json:"link_url"`
		Platform  string `json:"platform"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供广告信息"))
		return
	}
	ad, err := h.svc.CreateAd(req.Title, req.ImageURL, req.LinkURL, req.Platform, req.SortOrder)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "广告创建成功", "data": ad})
}

func (h *AdvertisementHandler) ListAds(c *gin.Context) {
	platform := c.Query("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ads, total, err := h.svc.ListAds(platform, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": ads, "total": total, "page": page, "page_size": pageSize})
}

func (h *AdvertisementHandler) UpdateAd(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	if err := h.svc.UpdateAd(id, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "广告更新成功"})
}

func (h *AdvertisementHandler) DeleteAd(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteAd(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "广告已删除"})
}

// ========== Franchise ==========

func (h *AdvertisementHandler) ApplyFranchise(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Phone   string `json:"phone" binding:"required"`
		Company string `json:"company"`
		Address string `json:"address"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供加盟信息"))
		return
	}
	app, err := h.svc.ApplyFranchise(req.Name, req.Phone, req.Company, req.Address, req.Remark)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "加盟申请已提交", "data": app})
}

func (h *AdvertisementHandler) ListFranchises(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	apps, total, err := h.svc.ListFranchises(status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": apps, "total": total, "page": page, "page_size": pageSize})
}

func (h *AdvertisementHandler) ProcessFranchise(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供处理状态"))
		return
	}
	processedBy, _ := c.Get("username")
	operator := ""
	if name, ok := processedBy.(string); ok {
		operator = name
	}
	if err := h.svc.ProcessFranchise(id, req.Status, operator); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "申请已处理"})
}

// ========== WechatUser ==========

func (h *AdvertisementHandler) ListWechatUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	users, total, err := h.svc.ListWechatUsers(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": users, "total": total, "page": page, "page_size": pageSize})
}

func (h *AdvertisementHandler) FreezeWechatUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.FreezeWechatUser(uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "用户已冻结"})
}
