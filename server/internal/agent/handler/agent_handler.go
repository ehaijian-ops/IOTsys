package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/agent/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AgentHandler 代理/运营商HTTP处理器
type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// ========== Agent ==========

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req struct {
		Name           string  `json:"name" binding:"required"`
		Contact        string  `json:"contact"`
		Phone          string  `json:"phone"`
		Email          string  `json:"email"`
		Address        string  `json:"address"`
		CommissionRate float64 `json:"commission_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的代理商信息"))
		return
	}
	agent, err := h.svc.CreateAgent(req.Name, req.Contact, req.Phone, req.Email, req.Address, req.CommissionRate)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "代理商创建成功", "data": agent})
}

func (h *AgentHandler) ListAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	agents, total, err := h.svc.ListAgents(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": agents, "total": total, "page": page, "page_size": pageSize})
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	agent, err := h.svc.GetAgent(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": agent})
}

func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	agent, err := h.svc.UpdateAgent(id, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "代理商更新成功", "data": agent})
}

func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteAgent(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "代理商已删除"})
}

// ========== Operator ==========

func (h *AgentHandler) CreateOperator(c *gin.Context) {
	var req struct {
		AgentID string `json:"agent_id"`
		Name    string `json:"name" binding:"required"`
		Contact string `json:"contact"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的运营商信息"))
		return
	}
	op, err := h.svc.CreateOperator(req.AgentID, req.Name, req.Contact, req.Phone, req.Address)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "运营商创建成功", "data": op})
}

func (h *AgentHandler) ListOperators(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	agentID := c.Query("agent_id")
	ops, total, err := h.svc.ListOperators(agentID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": ops, "total": total, "page": page, "page_size": pageSize})
}

func (h *AgentHandler) GetOperator(c *gin.Context) {
	id := c.Param("id")
	op, err := h.svc.GetOperator(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": op})
}

func (h *AgentHandler) UpdateOperator(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	op, err := h.svc.UpdateOperator(id, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "运营商更新成功", "data": op})
}

func (h *AgentHandler) DeleteOperator(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteOperator(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "运营商已删除"})
}
