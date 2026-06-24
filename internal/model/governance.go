package model

import "time"

// PermissionDefinition 是宿主和插件权限点的统一落库事实。
type PermissionDefinition struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	PermissionCode     string    `gorm:"size:200;uniqueIndex" json:"permission_code"`
	OwnerType          string    `gorm:"size:50;index" json:"owner_type"`
	OwnerPluginID      string    `gorm:"size:120;index" json:"owner_plugin_id"`
	RiskLevel          string    `gorm:"size:50;default:normal" json:"risk_level"`
	GroupKey           string    `gorm:"size:120;index" json:"group_key"`
	Name               string    `gorm:"size:200" json:"name"`
	Description        string    `gorm:"type:text" json:"description"`
	DefaultGrantPolicy string    `gorm:"size:50;default:manual" json:"default_grant_policy"`
	Status             string    `gorm:"size:50;default:active;index" json:"status"`
	ExtensionsJSON     string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (PermissionDefinition) TableName() string { return "permission_definitions" }

// RolePermissionGrant 记录角色被授予的权限点。
type RolePermissionGrant struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	RoleID             uint      `gorm:"uniqueIndex:idx_role_permission_grant" json:"role_id"`
	PermissionCode     string    `gorm:"size:200;uniqueIndex:idx_role_permission_grant" json:"permission_code"`
	GrantedBySubjectID string    `gorm:"size:120" json:"granted_by_subject_id"`
	GrantedAt          time.Time `json:"granted_at"`
}

func (RolePermissionGrant) TableName() string { return "role_permission_grants" }

// SubjectPermissionGrant 记录直接授予主体的权限点。
type SubjectPermissionGrant struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	SubjectID          string     `gorm:"size:120;uniqueIndex:idx_subject_permission_grant" json:"subject_id"`
	SubjectType        string     `gorm:"size:50;index" json:"subject_type"`
	PermissionCode     string     `gorm:"size:200;uniqueIndex:idx_subject_permission_grant" json:"permission_code"`
	GrantedBySubjectID string     `gorm:"size:120" json:"granted_by_subject_id"`
	GrantedAt          time.Time  `json:"granted_at"`
	ExpiresAt          *time.Time `json:"expires_at"`
}

func (SubjectPermissionGrant) TableName() string { return "subject_permission_grants" }

// ResourceScopeDefinition 定义资源可授权的数据范围类型。
type ResourceScopeDefinition struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ResourceType   string    `gorm:"size:120;uniqueIndex:idx_resource_scope_definition" json:"resource_type"`
	ScopeType      string    `gorm:"size:80;uniqueIndex:idx_resource_scope_definition" json:"scope_type"`
	OwnerPluginID  string    `gorm:"size:120;uniqueIndex:idx_resource_scope_definition" json:"owner_plugin_id"`
	Name           string    `gorm:"size:200" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	Status         string    `gorm:"size:50;default:active" json:"status"`
	ExtensionsJSON string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ResourceScopeDefinition) TableName() string { return "resource_scope_definitions" }

// SubjectDataScopeGrant 记录主体被授予的数据范围。
type SubjectDataScopeGrant struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	SubjectID          string     `gorm:"size:120;uniqueIndex:idx_subject_data_scope_grant" json:"subject_id"`
	SubjectType        string     `gorm:"size:50;index" json:"subject_type"`
	ResourceType       string     `gorm:"size:120;uniqueIndex:idx_subject_data_scope_grant" json:"resource_type"`
	ScopeType          string     `gorm:"size:80;uniqueIndex:idx_subject_data_scope_grant" json:"scope_type"`
	ScopeValue         string     `gorm:"size:200;uniqueIndex:idx_subject_data_scope_grant" json:"scope_value"`
	OwnerPluginID      string     `gorm:"size:120;uniqueIndex:idx_subject_data_scope_grant" json:"owner_plugin_id"`
	GrantedBySubjectID string     `gorm:"size:120" json:"granted_by_subject_id"`
	GrantedAt          time.Time  `json:"granted_at"`
	ExpiresAt          *time.Time `json:"expires_at"`
}

func (SubjectDataScopeGrant) TableName() string { return "subject_data_scope_grants" }

// EventLog 是宿主统一事件日志。
type EventLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EventID       string    `gorm:"size:120;uniqueIndex" json:"event_id"`
	EventType     string    `gorm:"size:120;index" json:"event_type"`
	SourceType    string    `gorm:"size:80;index" json:"source_type"`
	SourceID      string    `gorm:"size:120;index" json:"source_id"`
	OwnerPluginID string    `gorm:"size:120;index" json:"owner_plugin_id"`
	PayloadJSON   string    `gorm:"type:text" json:"payload_json"`
	Status        string    `gorm:"size:50;default:recorded" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (EventLog) TableName() string { return "event_logs" }

// SystemJob 是宿主统一任务记录。
type SystemJob struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	JobID         string    `gorm:"size:120;uniqueIndex" json:"job_id"`
	JobType       string    `gorm:"size:120;index" json:"job_type"`
	OwnerPluginID string    `gorm:"size:120;index" json:"owner_plugin_id"`
	Status        string    `gorm:"size:50;default:pending;index" json:"status"`
	RunAt         time.Time `gorm:"index" json:"run_at"`
	Attempts      int       `gorm:"default:0" json:"attempts"`
	MaxAttempts   int       `gorm:"default:3" json:"max_attempts"`
	PayloadJSON   string    `gorm:"type:text" json:"payload_json"`
	LastError     string    `gorm:"type:text" json:"last_error"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (SystemJob) TableName() string { return "system_jobs" }

// AuditLog 记录高危操作审计。
type AuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ActorSubjectID string    `gorm:"size:120;index" json:"actor_subject_id"`
	Action         string    `gorm:"size:160;index" json:"action"`
	ResourceType   string    `gorm:"size:120;index" json:"resource_type"`
	ResourceID     string    `gorm:"size:120;index" json:"resource_id"`
	RiskLevel      string    `gorm:"size:50;default:normal" json:"risk_level"`
	IP             string    `gorm:"size:80" json:"ip"`
	UserAgent      string    `gorm:"size:500" json:"user_agent"`
	PayloadDigest  string    `gorm:"size:128" json:"payload_digest"`
	PayloadJSON    string    `gorm:"type:text" json:"payload_json"`
	CreatedAt      time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
