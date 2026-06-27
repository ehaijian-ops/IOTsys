package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/system/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统管理HTTP处理器
type SystemHandler struct {
	svc *service.SystemService
}

func NewSystemHandler(svc *service.SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

// ========== Role ==========

func (h *SystemHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
		DataScope   string `json:"data_scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供角色信息"))
		return
	}
	role, err := h.svc.CreateRole(req.Name, req.DisplayName, req.Description, req.Permissions, req.DataScope)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "角色创建成功", "data": role})
}

func (h *SystemHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	roles, total, err := h.svc.ListRoles(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": roles, "total": total, "page": page, "page_size": pageSize})
}

func (h *SystemHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	if err := h.svc.UpdateRole(uint(id), req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "角色更新成功"})
}

func (h *SystemHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteRole(uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "角色已删除"})
}

// ========== Menu ==========

func (h *SystemHandler) GetMenuTree(c *gin.Context) {
	tree, err := h.svc.GetMenuTree()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": tree})
}

func (h *SystemHandler) CreateMenu(c *gin.Context) {
	var req struct {
		ParentID   *uint  `json:"parent_id"`
		Name       string `json:"name" binding:"required"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Icon       string `json:"icon"`
		Title      string `json:"title" binding:"required"`
		Permission string `json:"permission"`
		SortOrder  int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供菜单信息"))
		return
	}
	menu, err := h.svc.CreateMenu(req.ParentID, req.Name, req.Path, req.Component, req.Icon, req.Title, req.Permission, req.SortOrder)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "菜单创建成功", "data": menu})
}

func (h *SystemHandler) UpdateMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	if err := h.svc.UpdateMenu(uint(id), req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "菜单更新成功"})
}

func (h *SystemHandler) DeleteMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteMenu(uint(id)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "菜单已删除"})
}

// ========== Logs ==========

func (h *SystemHandler) ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	logs, total, err := h.svc.ListLoginLogs(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": logs, "total": total, "page": page, "page_size": pageSize})
}

func (h *SystemHandler) ListSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	module := c.Query("module")
	action := c.Query("action")
	logs, total, err := h.svc.ListSystemLogs(module, action, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": logs, "total": total, "page": page, "page_size": pageSize})
}
