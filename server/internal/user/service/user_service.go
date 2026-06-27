package service

import (
	"fmt"
	"time"

	"iot-platform/internal/user/model"
	"iot-platform/internal/user/repository"
	"iot-platform/pkg/auth"
	errs "iot-platform/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户业务逻辑层
type UserService struct {
	repo       *repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// Login 用户登录
func (s *UserService) Login(req *model.LoginRequest, clientIP string) (*model.LoginResponse, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, errs.ErrUnauthorized.Wrap(fmt.Errorf("invalid credentials"))
	}

	if !user.Enabled {
		return nil, errs.New("ACCOUNT_DISABLED", "账号已被禁用", 401)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errs.ErrUnauthorized.Wrap(fmt.Errorf("invalid credentials"))
	}

	// 生成 JWT
	roles := []string{user.Role}
	accessToken, err := s.jwtManager.GenerateAccessToken(
		fmt.Sprintf("%d", user.ID), user.Username, roles,
	)
	if err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	// 更新最后登录信息
	now := time.Now()
	s.repo.UpdateFields(user.ID, map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	})

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    7200, // 2小时
		User:         user,
	}, nil
}

// GetUserInfo 获取当前用户信息
func (s *UserService) GetUserInfo(userID uint) (*model.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errs.ErrNotFound.Wrap(err)
	}
	return user, nil
}

// CreateUser 创建用户（仅管理员可操作）
func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if s.repo.ExistsByUsername(req.Username, 0) {
		return nil, errs.New("USER_EXISTS", "用户名已存在", 409)
	}

	// 如果是 super_admin，检查是否已有 super_admin
	if req.Role == model.RoleSuperAdmin {
		count, _ := s.repo.CountByRole(model.RoleSuperAdmin)
		if count > 0 {
			return nil, errs.New("ROLE_LIMIT", "超级管理员只能有一个，如需移交请联系超管", 400)
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Role:     req.Role,
		Email:    req.Email,
		Phone:    req.Phone,
		Enabled:  true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, req *model.UpdateUserRequest, operatorRole string) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errs.ErrNotFound.Wrap(err)
	}

	// 权限检查：super_admin 不可被非 super_admin 修改
	if user.Role == model.RoleSuperAdmin && operatorRole != model.RoleSuperAdmin {
		return nil, errs.ErrForbidden.Wrap(fmt.Errorf("无权修改超级管理员"))
	}

	updates := make(map[string]interface{})

	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Enabled != nil {
		// 不能禁用自己
		if user.Role == model.RoleSuperAdmin && !*req.Enabled {
			return nil, errs.New("CANNOT_DISABLE_SUPERADMIN", "不能禁用超级管理员", 400)
		}
		updates["enabled"] = *req.Enabled
	}
	if req.Role != nil {
		// 仅 super_admin 可设置为 super_admin
		if *req.Role == model.RoleSuperAdmin && operatorRole != model.RoleSuperAdmin {
			return nil, errs.ErrForbidden.Wrap(fmt.Errorf("无权设置超级管理员角色"))
		}
		// 不允许将当前唯一的 super_admin 降级
		if user.Role == model.RoleSuperAdmin && *req.Role != model.RoleSuperAdmin {
			return nil, errs.New("CANNOT_DEMOTE_SUPERADMIN", "不能将唯一的超级管理员降级", 400)
		}
		updates["role"] = *req.Role
	}

	if len(updates) == 0 {
		return user, nil
	}

	if err := s.repo.UpdateFields(id, updates); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}

	// 重新查询返回最新数据
	return s.repo.FindByID(id)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint, operatorRole string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return errs.ErrNotFound.Wrap(err)
	}

	// super_admin 不可删除
	if user.Role == model.RoleSuperAdmin {
		return errs.New("CANNOT_DELETE_SUPERADMIN", "不可删除超级管理员", 400)
	}

	// 仅 super_admin 可删除 admin
	if user.Role == model.RoleAdmin && operatorRole != model.RoleSuperAdmin {
		return errs.ErrForbidden.Wrap(fmt.Errorf("无权删除管理员"))
	}

	return s.repo.Delete(id)
}

// ChangePassword 修改自己的密码
func (s *UserService) ChangePassword(userID uint, req *model.ChangePasswordRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errs.ErrNotFound.Wrap(err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errs.New("WRONG_PASSWORD", "原密码错误", 400)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}

	return s.repo.UpdateFields(userID, map[string]interface{}{
		"password": string(hashedPassword),
	})
}

// ResetPassword 管理员重置用户密码
func (s *UserService) ResetPassword(id uint, req *model.ResetPasswordRequest, operatorRole string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return errs.ErrNotFound.Wrap(err)
	}

	// 仅 super_admin 可重置 super_admin 的密码
	if user.Role == model.RoleSuperAdmin && operatorRole != model.RoleSuperAdmin {
		return errs.ErrForbidden.Wrap(fmt.Errorf("无权重置超级管理员密码"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.ErrInternalServer.Wrap(err)
	}

	return s.repo.UpdateFields(id, map[string]interface{}{
		"password": string(hashedPassword),
	})
}

// ListUsers 列出用户
func (s *UserService) ListUsers(page, pageSize int) ([]*model.UserListItem, int64, error) {
	users, total, err := s.repo.ListAll(page, pageSize)
	if err != nil {
		return nil, 0, errs.ErrInternalServer.Wrap(err)
	}

	items := make([]*model.UserListItem, len(users))
	for i, u := range users {
		items[i] = u.ToListItem()
	}
	return items, total, nil
}

// SeedDefaultAdmin 初始化默认超级管理员
func (s *UserService) SeedDefaultAdmin() error {
	if s.repo.ExistsByUsername("admin", 0) {
		return nil // 已存在，不重复创建
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: "admin",
		Password: string(hashedPassword),
		Nickname: "超级管理员",
		Role:     model.RoleSuperAdmin,
		Enabled:  true,
	}

	return s.repo.Create(admin)
}
