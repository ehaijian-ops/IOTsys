package service

import (
	"iot-platform/internal/system/model"
	"iot-platform/internal/system/repository"
	errs "iot-platform/pkg/errors"
)

// SystemService 系统管理业务逻辑层
type SystemService struct {
	repo *repository.SystemRepository
}

func NewSystemService(repo *repository.SystemRepository) *SystemService {
	return &SystemService{repo: repo}
}

// ========== Role ==========

func (s *SystemService) CreateRole(name, displayName, description, permissions, dataScope string) (*model.Role, error) {
	if _, err := s.repo.GetRoleByName(name); err == nil {
		return nil, errs.New("ROLE_EXISTS", "角色已存在", 409)
	}
	role := &model.Role{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Permissions: permissions,
		DataScope:   dataScope,
		Status:      "active",
	}
	if err := s.repo.CreateRole(role); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return role, nil
}

func (s *SystemService) ListRoles(page, pageSize int) ([]model.Role, int64, error) {
	return s.repo.ListRoles(page, pageSize)
}

func (s *SystemService) UpdateRole(id uint, updates map[string]interface{}) error {
	if _, err := s.repo.GetRole(id); err != nil {
		return errs.NotFound("角色", "")
	}
	return s.repo.UpdateRole(id, updates)
}

func (s *SystemService) DeleteRole(id uint) error {
	if _, err := s.repo.GetRole(id); err != nil {
		return errs.NotFound("角色", "")
	}
	return s.repo.DeleteRole(id)
}

// ========== Menu ==========

func (s *SystemService) CreateMenu(parentID *uint, name, path, component, icon, title, permission string, sortOrder int) (*model.Menu, error) {
	menu := &model.Menu{
		ParentID:   parentID,
		Name:       name,
		Path:       path,
		Component:  component,
		Icon:       icon,
		Title:      title,
		Permission: permission,
		SortOrder:  sortOrder,
		Status:     "active",
	}
	if err := s.repo.CreateMenu(menu); err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return menu, nil
}

// GetMenuTree 获取菜单树
func (s *SystemService) GetMenuTree() ([]*model.Menu, error) {
	menus, err := s.repo.ListAllMenus()
	if err != nil {
		return nil, errs.ErrInternalServer.Wrap(err)
	}
	return buildTree(menus), nil
}

func buildTree(menus []model.Menu) []*model.Menu {
	m := make(map[uint]*model.Menu)
	var roots []*model.Menu

	for i := range menus {
		menu := &menus[i]
		menu.Children = []*model.Menu{}
		m[menu.ID] = menu
	}

	for _, menu := range m {
		if menu.ParentID == nil || *menu.ParentID == 0 {
			roots = append(roots, menu)
		} else {
			if parent, ok := m[*menu.ParentID]; ok {
				parent.Children = append(parent.Children, menu)
			}
		}
	}
	return roots
}

func (s *SystemService) UpdateMenu(id uint, updates map[string]interface{}) error {
	if _, err := s.repo.GetMenu(id); err != nil {
		return errs.NotFound("菜单", "")
	}
	return s.repo.UpdateMenu(id, updates)
}

func (s *SystemService) DeleteMenu(id uint) error {
	if _, err := s.repo.GetMenu(id); err != nil {
		return errs.NotFound("菜单", "")
	}
	return s.repo.DeleteMenu(id)
}

// ========== Logs ==========

func (s *SystemService) ListLoginLogs(page, pageSize int) ([]model.LoginLog, int64, error) {
	return s.repo.ListLoginLogs(page, pageSize)
}

func (s *SystemService) ListSystemLogs(module, action string, page, pageSize int) ([]model.SystemLog, int64, error) {
	return s.repo.ListSystemLogs(module, action, page, pageSize)
}
