package handler

import (
	"net/http"
	"strconv"

	"iot-platform/internal/user/model"
	"iot-platform/internal/user/service"
	errs "iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请输入正确的用户名和密码"))
		return
	}

	resp, err := h.svc.Login(&req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "登录成功",
		"data":    resp,
	})
}

// GetUserInfo 获取当前用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, err := strconv.ParseUint(userID.(string), 10, 64)
	if err != nil {
		c.Error(errs.ErrUnauthorized)
		return
	}

	user, err := h.svc.GetUserInfo(uint(uid))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "OK",
		"data": user,
	})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := strconv.ParseUint(userID.(string), 10, 64)

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供原密码和新密码"))
		return
	}

	if err := h.svc.ChangePassword(uint(uid), &req); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "密码修改成功",
	})
}

// CreateUser 创建用户（管理员操作）
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供完整的用户信息"))
		return
	}

	user, err := h.svc.CreateUser(&req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "用户创建成功",
		"data":    user.ToListItem(),
	})
}

// ListUsers 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.svc.ListUsers(page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  "OK",
		"data":  users,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.ErrBadRequest.Wrap(err))
		return
	}

	user, err := h.svc.GetUserInfo(uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	// 判断是否可以修改角色
	roles, _ := c.Get("roles")
	canEditRole := false
	if roleList, ok := roles.([]string); ok {
		for _, r := range roleList {
			if r == model.RoleSuperAdmin || r == model.RoleAdmin {
				canEditRole = true
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":          "OK",
		"data":          user,
		"can_edit_role": canEditRole,
	})
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.ErrBadRequest.Wrap(err))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请求参数错误"))
		return
	}

	// 获取操作者角色
	roles, _ := c.Get("roles")
	operatorRole := "viewer"
	if roleList, ok := roles.([]string); ok && len(roleList) > 0 {
		operatorRole = roleList[0]
	}

	user, err := h.svc.UpdateUser(uint(id), &req, operatorRole)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "用户更新成功",
		"data":    user.ToListItem(),
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.ErrBadRequest.Wrap(err))
		return
	}

	roles, _ := c.Get("roles")
	operatorRole := "viewer"
	if roleList, ok := roles.([]string); ok && len(roleList) > 0 {
		operatorRole = roleList[0]
	}

	if err := h.svc.DeleteUser(uint(id), operatorRole); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "用户删除成功",
	})
}

// ResetPassword 管理员重置用户密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.ErrBadRequest.Wrap(err))
		return
	}

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errs.ValidationError("请提供新密码"))
		return
	}

	roles, _ := c.Get("roles")
	operatorRole := "viewer"
	if roleList, ok := roles.([]string); ok && len(roleList) > 0 {
		operatorRole = roleList[0]
	}

	if err := h.svc.ResetPassword(uint(id), &req, operatorRole); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "密码重置成功",
	})
}

// GetRoles 获取可用角色列表
func (h *UserHandler) GetRoles(c *gin.Context) {
	roles, _ := c.Get("roles")
	operatorRole := "viewer"
	if roleList, ok := roles.([]string); ok && len(roleList) > 0 {
		operatorRole = roleList[0]
	}

	// 根据当前用户角色返回可分配的角色列表
	assignableRoles := map[string][]map[string]string{
		model.RoleSuperAdmin: {
			{"value": model.RoleSuperAdmin, "label": "超级管理员"},
			{"value": model.RoleAdmin, "label": "管理员"},
			{"value": model.RoleOperator, "label": "运维人员"},
			{"value": model.RoleViewer, "label": "查看者"},
		},
		model.RoleAdmin: {
			{"value": model.RoleAdmin, "label": "管理员"},
			{"value": model.RoleOperator, "label": "运维人员"},
			{"value": model.RoleViewer, "label": "查看者"},
		},
		model.RoleOperator: {},
		model.RoleViewer:    {},
	}

	rolesList := assignableRoles[operatorRole]
	if rolesList == nil {
		rolesList = []map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "OK",
		"data": rolesList,
	})
}

// ========== 角色常量导出（供 middleware 使用） ==========
var (
	RoleSuperAdmin = model.RoleSuperAdmin
	RoleAdmin      = model.RoleAdmin
	RoleOperator   = model.RoleOperator
	RoleViewer     = model.RoleViewer
)
