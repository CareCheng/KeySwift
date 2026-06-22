package service

import (
	"encoding/json"
	"errors"
	"strings"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// RoleService 角色权限服务
type RoleService struct {
	repo             *repository.Repository
	extraPermissions []model.Permission
}

// NewRoleService 创建角色权限服务
func NewRoleService(repo *repository.Repository) *RoleService {
	return &RoleService{repo: repo}
}

// SetExtraPermissions 设置当前插件等运行时声明的权限点。
func (s *RoleService) SetExtraPermissions(permissions []model.Permission) {
	s.extraPermissions = append([]model.Permission(nil), permissions...)
}

// InitDefaultRoles 初始化默认角色
func (s *RoleService) InitDefaultRoles() error {
	db := s.repo.GetDB()

	// 检查是否已有角色
	var count int64
	db.Model(&model.AdminRole{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建超级管理员角色（拥有所有权限）
	allPerms := make([]string, 0, len(model.AllPermissions))
	for _, p := range model.AllPermissions {
		allPerms = append(allPerms, p.Code)
	}
	allPermsJSON, _ := json.Marshal(allPerms)

	superAdmin := &model.AdminRole{
		Name:        "super_admin",
		Description: "超级管理员，拥有所有权限",
		Permissions: string(allPermsJSON),
		IsSystem:    true,
		Status:      1,
	}
	if err := db.Create(superAdmin).Error; err != nil {
		return err
	}

	// 创建普通管理员角色（基础管理权限）
	normalPerms := []string{
		"dashboard:view",
		"product:view", "product:create", "product:edit",
		"category:view", "category:create", "category:edit",
		"order:view", "order:edit",
		"user:view",
		"balance:view", "balance:adjust",
		"settings:view",
		"plugin:view",
		"log:view",
	}
	normalPermsJSON, _ := json.Marshal(normalPerms)

	normalAdmin := &model.AdminRole{
		Name:        "admin",
		Description: "普通管理员，拥有基础管理权限",
		Permissions: string(normalPermsJSON),
		IsSystem:    true,
		Status:      1,
	}
	if err := db.Create(normalAdmin).Error; err != nil {
		return err
	}

	// 创建运营角色
	operatorPerms := []string{
		"dashboard:view",
		"product:view", "product:create", "product:edit",
		"category:view", "category:create", "category:edit",
		"order:view",
		"user:view",
		"balance:view",
	}
	operatorPermsJSON, _ := json.Marshal(operatorPerms)

	operatorRole := &model.AdminRole{
		Name:        "operator",
		Description: "运营人员，负责商品和活动管理",
		Permissions: string(operatorPermsJSON),
		IsSystem:    true,
		Status:      1,
	}
	return db.Create(operatorRole).Error
}

func (s *RoleService) permissionMap() map[string]bool {
	permMap := make(map[string]bool)
	for _, p := range model.AllPermissions {
		permMap[p.Code] = true
	}
	for _, p := range s.extraPermissions {
		permMap[p.Code] = true
	}
	return permMap
}

// GetAllRoles 获取所有角色
func (s *RoleService) GetAllRoles() ([]model.AdminRole, error) {
	var roles []model.AdminRole
	err := s.repo.GetDB().Where("status = ?", 1).Order("id ASC").Find(&roles).Error
	return roles, err
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id uint) (*model.AdminRole, error) {
	var role model.AdminRole
	err := s.repo.GetDB().First(&role, id).Error
	return &role, err
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(name string) (*model.AdminRole, error) {
	var role model.AdminRole
	err := s.repo.GetDB().Where("name = ?", name).First(&role).Error
	return &role, err
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(name, description string, permissions []string) (*model.AdminRole, error) {
	// 检查名称是否已存在
	var count int64
	s.repo.GetDB().Model(&model.AdminRole{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return nil, errors.New("角色名称已存在")
	}

	// 验证权限代码
	validPerms := make([]string, 0)
	permMap := s.permissionMap()
	for _, p := range permissions {
		if permMap[p] {
			validPerms = append(validPerms, p)
		}
	}

	permsJSON, _ := json.Marshal(validPerms)
	role := &model.AdminRole{
		Name:        name,
		Description: description,
		Permissions: string(permsJSON),
		IsSystem:    false,
		Status:      1,
	}

	err := s.repo.GetDB().Create(role).Error
	return role, err
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(id uint, name, description string, permissions []string) error {
	role, err := s.GetRoleByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 系统内置角色不允许修改名称
	if role.IsSystem && role.Name != name {
		return errors.New("系统内置角色不允许修改名称")
	}

	// 检查名称是否与其他角色冲突
	var count int64
	s.repo.GetDB().Model(&model.AdminRole{}).Where("name = ? AND id != ?", name, id).Count(&count)
	if count > 0 {
		return errors.New("角色名称已存在")
	}

	// 验证权限代码
	validPerms := make([]string, 0)
	permMap := s.permissionMap()
	for _, p := range permissions {
		if permMap[p] {
			validPerms = append(validPerms, p)
		}
	}

	permsJSON, _ := json.Marshal(validPerms)
	return s.repo.GetDB().Model(role).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"permissions": string(permsJSON),
	}).Error
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id uint) error {
	role, err := s.GetRoleByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}

	if role.IsSystem {
		return errors.New("系统内置角色不允许删除")
	}

	// 检查是否有管理员使用此角色
	var count int64
	s.repo.GetDB().Model(&model.Admin{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该角色下还有管理员，无法删除")
	}

	return s.repo.GetDB().Delete(role).Error
}

// GetRolePermissions 获取角色权限列表
func (s *RoleService) GetRolePermissions(roleID uint) ([]string, error) {
	role, err := s.GetRoleByID(roleID)
	if err != nil {
		return nil, err
	}

	var permissions []string
	if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

// HasPermission 检查角色是否有指定权限
func (s *RoleService) HasPermission(roleID uint, permission string) bool {
	permissions, err := s.GetRolePermissions(roleID)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
		// 支持通配符，如 product:* 匹配所有 product 权限
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}
	return false
}

// GetAllPermissions 获取所有权限定义
func (s *RoleService) GetAllPermissions() []model.Permission {
	permissions := append([]model.Permission(nil), model.AllPermissions...)
	permissions = append(permissions, s.extraPermissions...)
	return permissions
}

// GetPermissionGroups 获取按分组整理的权限
func (s *RoleService) GetPermissionGroups() map[string][]model.Permission {
	groups := make(map[string][]model.Permission)
	for _, p := range s.GetAllPermissions() {
		groups[p.Group] = append(groups[p.Group], p)
	}
	return groups
}

// ==================== 管理员管理 ====================

// GetAllAdmins 获取所有管理员
func (s *RoleService) GetAllAdmins(page, pageSize int) ([]model.Admin, int64, error) {
	var admins []model.Admin
	var total int64

	db := s.repo.GetDB().Model(&model.Admin{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Preload("Role").Order("id ASC").Offset(offset).Limit(pageSize).Find(&admins).Error
	return admins, total, err
}

// GetAdminByID 根据ID获取管理员
func (s *RoleService) GetAdminByID(id uint) (*model.Admin, error) {
	var admin model.Admin
	err := s.repo.GetDB().Preload("Role").First(&admin, id).Error
	return &admin, err
}

// GetAdminByUsername 根据用户名获取管理员
func (s *RoleService) GetAdminByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	err := s.repo.GetDB().Preload("Role").Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, err
}

// CreateSuperAdmin 创建超级管理员（用于初始化设置）
func (s *RoleService) CreateSuperAdmin(username, password string) error {
	// 确保 super_admin 角色存在
	var superRole model.AdminRole
	err := s.repo.GetDB().Where("name = ?", "super_admin").First(&superRole).Error
	if err != nil {
		// 创建超级管理员角色
		superRole = model.AdminRole{
			Name:        "super_admin",
			Description: "超级管理员，拥有所有权限",
			Permissions: "[]", // 超级管理员不需要定义具体权限
			IsSystem:    true,
			Status:      1,
		}
		if err := s.repo.GetDB().Create(&superRole).Error; err != nil {
			return err
		}
	}

	// 检查用户名是否已存在
	var count int64
	s.repo.GetDB().Model(&model.Admin{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.Admin{
		Username:     username,
		PasswordHash: string(hashedPassword),
		RoleID:       superRole.ID,
		Nickname:     "超级管理员",
		Status:       1,
	}

	return s.repo.GetDB().Create(admin).Error
}

// CreateAdmin 创建管理员
func (s *RoleService) CreateAdmin(username, password, email, nickname string, roleID uint) (*model.Admin, error) {
	// 检查用户名是否已存在于 admins 表
	var count int64
	s.repo.GetDB().Model(&model.Admin{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 检查用户名是否与系统配置账户冲突
	var sysConfig model.SystemConfigDB
	if err := s.repo.GetDB().First(&sysConfig).Error; err == nil {
		if sysConfig.AdminUsername == username {
			return nil, errors.New("用户名与系统配置账户冲突，请使用其他用户名")
		}
	}

	// 检查角色是否存在
	_, err := s.GetRoleByID(roleID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &model.Admin{
		Username:     username,
		PasswordHash: string(hashedPassword),
		RoleID:       roleID,
		Email:        email,
		Nickname:     nickname,
		Status:       1,
	}

	err = s.repo.GetDB().Create(admin).Error
	return admin, err
}

// UpdateAdmin 更新管理员
func (s *RoleService) UpdateAdmin(id uint, email, nickname string, roleID uint, status int) error {
	admin, err := s.GetAdminByID(id)
	if err != nil {
		return errors.New("管理员不存在")
	}

	// 检查角色是否存在
	_, err = s.GetRoleByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}

	return s.repo.GetDB().Model(admin).Updates(map[string]interface{}{
		"email":    email,
		"nickname": nickname,
		"role_id":  roleID,
		"status":   status,
	}).Error
}

// UpdateAdminPassword 更新管理员密码
func (s *RoleService) UpdateAdminPassword(id uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.GetDB().Model(&model.Admin{}).Where("id = ?", id).
		Update("password_hash", string(hashedPassword)).Error
}

// DeleteAdmin 删除管理员
func (s *RoleService) DeleteAdmin(id uint) error {
	admin, err := s.GetAdminByID(id)
	if err != nil {
		return errors.New("管理员不存在")
	}

	// 检查是否是最后一个超级管理员
	if admin.Role != nil && admin.Role.Name == "super_admin" {
		var count int64
		s.repo.GetDB().Model(&model.Admin{}).
			Joins("JOIN admin_roles ON admins.role_id = admin_roles.id").
			Where("admin_roles.name = ? AND admins.status = 1", "super_admin").
			Count(&count)
		if count <= 1 {
			return errors.New("不能删除最后一个超级管理员")
		}
	}

	return s.repo.GetDB().Delete(admin).Error
}

// VerifyAdminPassword 验证管理员密码
func (s *RoleService) VerifyAdminPassword(username, password string) (*model.Admin, error) {
	admin, err := s.GetAdminByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if admin.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return admin, nil
}

// EnableAdmin2FA 启用管理员两步验证。
func (s *RoleService) EnableAdmin2FA(username, secret string) error {
	admin, err := s.GetAdminByUsername(username)
	if err != nil {
		return errors.New("管理员不存在")
	}

	admin.Enable2FA = true
	admin.TOTPSecret = secret
	return s.repo.GetDB().Select("Enable2FA", "TOTPSecret").Save(admin).Error
}

// DisableAdmin2FA 禁用管理员两步验证。
func (s *RoleService) DisableAdmin2FA(username string) error {
	admin, err := s.GetAdminByUsername(username)
	if err != nil {
		return errors.New("管理员不存在")
	}

	admin.Enable2FA = false
	admin.TOTPSecret = ""
	return s.repo.GetDB().Select("Enable2FA", "TOTPSecret").Save(admin).Error
}

// GetAdmin2FAStatus 获取管理员两步验证状态。
func (s *RoleService) GetAdmin2FAStatus(username string) (bool, string, error) {
	admin, err := s.GetAdminByUsername(username)
	if err != nil {
		return false, "", err
	}
	return admin.Enable2FA, admin.TOTPSecret, nil
}

// UpdateAdminLoginInfo 更新管理员登录信息
func (s *RoleService) UpdateAdminLoginInfo(id uint, ip string) error {
	return s.repo.GetDB().Model(&model.Admin{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": s.repo.GetDB().NowFunc(),
		"last_login_ip": ip,
	}).Error
}

// AdminHasPermission 检查管理员是否有指定权限
func (s *RoleService) AdminHasPermission(adminID uint, permission string) bool {
	admin, err := s.GetAdminByID(adminID)
	if err != nil || admin.Status != 1 {
		return false
	}

	// 超级管理员拥有所有权限
	if admin.Role != nil && admin.Role.Name == "super_admin" {
		return true
	}

	return s.HasPermission(admin.RoleID, permission)
}

// GetAdminPermissions 获取管理员的所有权限
func (s *RoleService) GetAdminPermissions(adminID uint) ([]string, error) {
	admin, err := s.GetAdminByID(adminID)
	if err != nil {
		return nil, err
	}

	return s.GetRolePermissions(admin.RoleID)
}
