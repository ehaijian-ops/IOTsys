package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/interconnect/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// InterconnectHandler 互联互通HTTP处理器
type InterconnectHandler struct {
	svc *service.InterconnectService
}

func NewInterconnectHandler(svc *service.InterconnectService) *InterconnectHandler {
	return &InterconnectHandler{svc: svc}
}

// ========== Org ==========

func (h *InterconnectHandler) CreateOrg(c *gin.Context) {
	var req struct {
		Name              string `json:"name" binding:"required"`
		OrgCode           string `json:"org_code" binding:"required"`
		Contact           string `json:"contact"`
		Phone             string `json:"phone"`
		PushURL           string `json:"push_url"`
		ReconciliationURL string `json:"reconciliation_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的机构信息"))
		return
	}
	org, err := h.svc.CreateOrg(req.Name, req.OrgCode, req.Contact, req.Phone, req.PushURL, req.ReconciliationURL)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "互联机构创建成功", "data": org})
}

func (h *InterconnectHandler) ListOrgs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orgs, total, err := h.svc.ListOrgs(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": orgs, "total": total, "page": page, "page_size": pageSize})
}

func (h *InterconnectHandler) GetOrg(c *gin.Context) {
	id := c.Param("id")
	org, err := h.svc.GetOrg(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": org})
}

func (h *InterconnectHandler) UpdateOrg(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	if err := h.svc.UpdateOrg(id, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "机构更新成功"})
}

func (h *InterconnectHandler) DeleteOrg(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteOrg(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "机构已删除"})
}

// ========== Key ==========

func (h *InterconnectHandler) CreateKey(c *gin.Context) {
	var req struct {
		OrgID      string `json:"org_id" binding:"required"`
		KeyType    string `json:"key_type" binding:"required"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
		SecretKey  string `json:"secret_key"`
		Remark     string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的密钥信息"))
		return
	}
	key, err := h.svc.CreateKey(req.OrgID, req.KeyType, req.PublicKey, req.PrivateKey, req.SecretKey, req.Remark)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "密钥创建成功", "data": key})
}

func (h *InterconnectHandler) ListKeys(c *gin.Context) {
	orgID := c.Query("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keys, total, err := h.svc.ListKeys(orgID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": keys, "total": total, "page": page, "page_size": pageSize})
}

func (h *InterconnectHandler) DeleteKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteKey(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "密钥已删除"})
}
