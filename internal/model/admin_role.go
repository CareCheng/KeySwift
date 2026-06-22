package model

import "time"

// AdminRole 管理员角色。
type AdminRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	Permissions string    `gorm:"type:text" json:"permissions"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Admin 管理员账户。
type Admin struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	RoleID       uint       `gorm:"index" json:"role_id"`
	Role         *AdminRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Email        string     `gorm:"size:255" json:"email"`
	Nickname     string     `gorm:"size:100" json:"nickname"`
	Avatar       string     `gorm:"size:500" json:"avatar"`
	Enable2FA    bool       `gorm:"default:false" json:"enable_2fa"`
	TOTPSecret   string     `gorm:"size:64" json:"-"`
	Status       int        `gorm:"default:1" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `gorm:"size:50" json:"last_login_ip"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Permission 权限定义。
type Permission struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Group       string `json:"group"`
}

// AllPermissions 核心权限列表。
var AllPermissions = []Permission{
	{Code: "dashboard:view", Name: "查看仪表盘", Description: "查看系统仪表盘统计数据", Group: "仪表盘"},
	{Code: "product:view", Name: "查看商品", Description: "查看商品列表和详情", Group: "商品管理"},
	{Code: "product:create", Name: "创建商品", Description: "创建新商品", Group: "商品管理"},
	{Code: "product:edit", Name: "编辑商品", Description: "编辑商品信息", Group: "商品管理"},
	{Code: "product:delete", Name: "删除商品", Description: "删除商品", Group: "商品管理"},
	{Code: "category:view", Name: "查看分类", Description: "查看商品分类", Group: "分类管理"},
	{Code: "category:create", Name: "创建分类", Description: "创建新分类", Group: "分类管理"},
	{Code: "category:edit", Name: "编辑分类", Description: "编辑分类信息", Group: "分类管理"},
	{Code: "category:delete", Name: "删除分类", Description: "删除分类", Group: "分类管理"},
	{Code: "order:view", Name: "查看订单", Description: "查看订单列表和详情", Group: "订单管理"},
	{Code: "order:edit", Name: "编辑订单", Description: "编辑订单状态", Group: "订单管理"},
	{Code: "user:view", Name: "查看用户", Description: "查看用户列表和详情", Group: "用户管理"},
	{Code: "user:edit", Name: "编辑用户", Description: "编辑用户状态", Group: "用户管理"},
	{Code: "balance:view", Name: "查看余额", Description: "查看用户余额和余额日志", Group: "余额管理"},
	{Code: "balance:adjust", Name: "调整余额", Description: "后台调整用户余额", Group: "余额管理"},
	{Code: "settings:view", Name: "查看设置", Description: "查看系统设置", Group: "系统设置"},
	{Code: "settings:edit", Name: "编辑设置", Description: "编辑系统设置", Group: "系统设置"},
	{Code: "settings:payment", Name: "支付配置", Description: "查看余额支付配置", Group: "系统设置"},
	{Code: "settings:email", Name: "邮箱配置", Description: "配置邮箱服务", Group: "系统设置"},
	{Code: "settings:database", Name: "数据库配置", Description: "配置数据库", Group: "系统设置"},
	{Code: "settings:security", Name: "安全设置", Description: "配置安全选项", Group: "系统设置"},
	{Code: "plugin:view", Name: "查看插件", Description: "查看插件列表、能力声明和运行状态", Group: "插件管理"},
	{Code: "plugin:manage", Name: "管理插件", Description: "刷新插件目录和管理插件启停治理", Group: "插件管理"},
	{Code: "log:view", Name: "查看日志", Description: "查看操作日志", Group: "日志管理"},
	{Code: "admin:view", Name: "查看管理员", Description: "查看管理员列表", Group: "管理员管理"},
	{Code: "admin:create", Name: "创建管理员", Description: "创建新管理员", Group: "管理员管理"},
	{Code: "admin:edit", Name: "编辑管理员", Description: "编辑管理员信息", Group: "管理员管理"},
	{Code: "admin:delete", Name: "删除管理员", Description: "删除管理员", Group: "管理员管理"},
	{Code: "role:view", Name: "查看角色", Description: "查看角色列表", Group: "角色管理"},
	{Code: "role:create", Name: "创建角色", Description: "创建新角色", Group: "角色管理"},
	{Code: "role:edit", Name: "编辑角色", Description: "编辑角色权限", Group: "角色管理"},
	{Code: "role:delete", Name: "删除角色", Description: "删除角色", Group: "角色管理"},
}

// GetPermissionGroups 获取按分组整理的权限列表。
func GetPermissionGroups() map[string][]Permission {
	groups := make(map[string][]Permission)
	for _, p := range AllPermissions {
		groups[p.Group] = append(groups[p.Group], p)
	}
	return groups
}

// PermissionTemplate 权限模板。
type PermissionTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// PermissionTemplates 核心权限模板。
var PermissionTemplates = []PermissionTemplate{
	{
		Name:        "viewer",
		Description: "只读用户，只能查看核心数据",
		Permissions: []string{
			"dashboard:view",
			"product:view",
			"category:view",
			"order:view",
			"user:view",
			"balance:view",
			"log:view",
		},
	},
	{
		Name:        "operator",
		Description: "运营人员，负责商品、分类和订单处理",
		Permissions: []string{
			"dashboard:view",
			"product:view", "product:create", "product:edit",
			"category:view", "category:create", "category:edit",
			"order:view", "order:edit",
			"user:view",
			"balance:view",
		},
	},
	{
		Name:        "admin",
		Description: "普通管理员，拥有核心管理权限",
		Permissions: []string{
			"dashboard:view",
			"product:view", "product:create", "product:edit", "product:delete",
			"category:view", "category:create", "category:edit", "category:delete",
			"order:view", "order:edit",
			"user:view", "user:edit",
			"balance:view", "balance:adjust",
			"settings:view",
			"plugin:view",
			"log:view",
		},
	},
	{
		Name:        "super_admin",
		Description: "超级管理员，拥有所有核心权限",
		Permissions: func() []string {
			perms := make([]string, 0, len(AllPermissions))
			for _, p := range AllPermissions {
				perms = append(perms, p.Code)
			}
			return perms
		}(),
	},
}

// GetPermissionTemplate 根据名称获取权限模板。
func GetPermissionTemplate(name string) *PermissionTemplate {
	for _, t := range PermissionTemplates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// TableName 设置表名。
func (AdminRole) TableName() string {
	return "admin_roles"
}

// TableName 设置表名。
func (Admin) TableName() string {
	return "admins"
}
