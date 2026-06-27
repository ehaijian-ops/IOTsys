package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/maintenance/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// MaintenanceHandler 运维HTTP处理器
type MaintenanceHandler struct {
	svc *service.MaintenanceService
}

func NewMaintenanceHandler(svc *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

// ========== FaultReport ==========

func (h *MaintenanceHandler) CreateFault(c *gin.Context) {
	var req struct {
		DeviceID    string `json:"device_id" binding:"required"`
		FaultType   string `json:"fault_type" binding:"required"`
		Description string `json:"description"`
		Images      string `json:"images"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供故障信息"))
		return
	}
	fault, err := h.svc.CreateFault(req.DeviceID, req.FaultType, req.Description, req.Images, nil)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "故障已上报", "data": fault})
}

func (h *MaintenanceHandler) ListFaults(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceID := c.Query("device_id")
	status := c.Query("status")

	faults, total, err := h.svc.ListFaults(deviceID, status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": faults, "total": total, "page": page, "page_size": pageSize})
}

func (h *MaintenanceHandler) HandleFault(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Result string `json:"result" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供处理结果"))
		return
	}
	handledBy, _ := c.Get("username")
	operator := ""
	if name, ok := handledBy.(string); ok {
		operator = name
	}
	if err := h.svc.HandleFault(id, req.Result, operator); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "故障已处理"})
}

// ========== ScheduledTask ==========

func (h *MaintenanceHandler) CreateTask(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		TaskType    string `json:"task_type" binding:"required"`
		CronExpr    string `json:"cron_expr" binding:"required"`
		Params      string `json:"params"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的任务信息"))
		return
	}
	createdBy, _ := c.Get("username")
	operator := ""
	if name, ok := createdBy.(string); ok {
		operator = name
	}
	task, err := h.svc.CreateTask(req.Name, req.TaskType, req.CronExpr, req.Params, req.Description, operator)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "定时任务创建成功", "data": task})
}

func (h *MaintenanceHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tasks, total, err := h.svc.ListTasks(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": tasks, "total": total, "page": page, "page_size": pageSize})
}

func (h *MaintenanceHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}
	if err := h.svc.UpdateTask(id, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "任务更新成功"})
}

func (h *MaintenanceHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteTask(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "任务已删除"})
}

func (h *MaintenanceHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	logs, total, err := h.svc.GetTaskLogs(taskID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": logs, "total": total, "page": page, "page_size": pageSize})
}

// ========== DownloadTask ==========

func (h *MaintenanceHandler) CreateDownload(c *gin.Context) {
	var req struct {
		TaskType string `json:"task_type" binding:"required"`
		Params   string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供导出参数"))
		return
	}
	createdBy, _ := c.Get("username")
	operator := ""
	if name, ok := createdBy.(string); ok {
		operator = name
	}
	task, err := h.svc.CreateDownload(req.TaskType, req.Params, operator)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "下载任务已创建", "data": task})
}

func (h *MaintenanceHandler) ListDownloads(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	createdBy, _ := c.Get("username")
	operator := ""
	if name, ok := createdBy.(string); ok {
		operator = name
	}
	tasks, total, err := h.svc.ListDownloads(operator, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": tasks, "total": total, "page": page, "page_size": pageSize})
}
