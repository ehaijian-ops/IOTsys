package handler

import (
	"net/http"
	"time"

	"iot-platform/internal/device/model"
	"iot-platform/internal/device/repository"
	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SiteHandler 站点管理处理器
type SiteHandler struct {
	db      *gorm.DB
	devRepo *repository.DeviceRepository
}

// NewSiteHandler 创建站点处理器
func NewSiteHandler(db *gorm.DB, devRepo *repository.DeviceRepository) *SiteHandler {
	return &SiteHandler{db: db, devRepo: devRepo}
}

// SiteResponse 站点响应（含设备数量）
type SiteResponse struct {
	model.Site
	DeviceCount int64 `json:"device_count"`
}

// ListSites 获取站点列表
// GET /api/v1/sites
func (h *SiteHandler) ListSites(c *gin.Context) {
	var sites []model.Site
	if err := h.db.Order("created_at DESC").Find(&sites).Error; err != nil {
		logger.Error("Failed to list sites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "查询站点失败"})
		return
	}

	result := make([]SiteResponse, 0, len(sites))
	for _, s := range sites {
		var count int64
		h.db.Model(&model.Device{}).Where("site_id = ?", s.ID).Count(&count)
		result = append(result, SiteResponse{
			Site:        s,
			DeviceCount: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": result,
	})
}

// GetSite 获取站点详情
// GET /api/v1/sites/:id
func (h *SiteHandler) GetSite(c *gin.Context) {
	id := c.Param("id")

	var site model.Site
	if err := h.db.Where("id = ?", id).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": "ERROR", "message": "站点不存在"})
			return
		}
		logger.Error("Failed to get site", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "查询站点失败"})
		return
	}

	var count int64
	h.db.Model(&model.Device{}).Where("site_id = ?", site.ID).Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"data": SiteResponse{Site: site, DeviceCount: count},
	})
}

// CreateSiteRequest 创建站点请求
type CreateSiteRequest struct {
	Name      string  `json:"name" binding:"required"`
	Address   string  `json:"address"`
	Contact   string  `json:"contact"`
	Phone     string  `json:"phone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

// CreateSite 创建站点
// POST /api/v1/sites
func (h *SiteHandler) CreateSite(c *gin.Context) {
	var req CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ERROR", "message": "参数错误: " + err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	site := model.Site{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Address:   req.Address,
		Contact:   req.Contact,
		Phone:     req.Phone,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.db.Create(&site).Error; err != nil {
		logger.Error("Failed to create site", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "创建站点失败: " + err.Error()})
		return
	}

	logger.Info("Site created", zap.String("id", site.ID), zap.String("name", site.Name))

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "站点创建成功",
		"data":    SiteResponse{Site: site, DeviceCount: 0},
	})
}

// UpdateSiteRequest 更新站点请求
type UpdateSiteRequest struct {
	Name      *string  `json:"name"`
	Address   *string  `json:"address"`
	Contact   *string  `json:"contact"`
	Phone     *string  `json:"phone"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Status    *string  `json:"status"`
}

// UpdateSite 更新站点
// PUT /api/v1/sites/:id
func (h *SiteHandler) UpdateSite(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "ERROR", "message": "参数错误: " + err.Error()})
		return
	}

	var site model.Site
	if err := h.db.Where("id = ?", id).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": "ERROR", "message": "站点不存在"})
			return
		}
		logger.Error("Failed to find site", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "查询站点失败"})
		return
	}

	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.Address != nil {
		site.Address = *req.Address
	}
	if req.Contact != nil {
		site.Contact = *req.Contact
	}
	if req.Phone != nil {
		site.Phone = *req.Phone
	}
	if req.Latitude != nil {
		site.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		site.Longitude = *req.Longitude
	}
	if req.Status != nil {
		site.Status = *req.Status
	}
	site.UpdatedAt = time.Now()

	if err := h.db.Save(&site).Error; err != nil {
		logger.Error("Failed to update site", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "更新站点失败: " + err.Error()})
		return
	}

	var count int64
	h.db.Model(&model.Device{}).Where("site_id = ?", site.ID).Count(&count)

	logger.Info("Site updated", zap.String("id", site.ID), zap.String("name", site.Name))

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "站点更新成功",
		"data":    SiteResponse{Site: site, DeviceCount: count},
	})
}

// DeleteSite 删除站点
// DELETE /api/v1/sites/:id
func (h *SiteHandler) DeleteSite(c *gin.Context) {
	id := c.Param("id")

	// 检查站点下是否有设备
	var deviceCount int64
	h.db.Model(&model.Device{}).Where("site_id = ?", id).Count(&deviceCount)
	if deviceCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERROR",
			"message": "该站点下还有设备，请先将设备迁移或删除后再删除站点",
		})
		return
	}

	result := h.db.Delete(&model.Site{}, "id = ?", id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": "ERROR", "message": "站点不存在"})
		return
	}
	if result.Error != nil {
		logger.Error("Failed to delete site", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ERROR", "message": "删除站点失败"})
		return
	}

	logger.Info("Site deleted", zap.String("id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "站点已删除",
	})
}
