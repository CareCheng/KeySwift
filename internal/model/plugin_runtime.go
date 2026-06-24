package model

import "time"

// PluginVersion 记录插件安装版本和校验摘要。
type PluginVersion struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PluginID     string    `gorm:"size:120;uniqueIndex:idx_plugin_version" json:"plugin_id"`
	Version      string    `gorm:"size:50;uniqueIndex:idx_plugin_version" json:"version"`
	ManifestHash string    `gorm:"size:128" json:"manifest_hash"`
	PackageHash  string    `gorm:"size:128" json:"package_hash"`
	BinaryHash   string    `gorm:"size:128" json:"binary_hash"`
	InstallPath  string    `gorm:"size:500" json:"install_path"`
	Status       string    `gorm:"size:50;default:installed" json:"status"`
	InstalledAt  time.Time `json:"installed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PluginVersion) TableName() string { return "plugin_versions" }

// PluginRuntimeSession 记录插件每次运行实例。
type PluginRuntimeSession struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	PluginID        string     `gorm:"size:120;index" json:"plugin_id"`
	Version         string     `gorm:"size:50" json:"version"`
	InstanceID      string     `gorm:"size:120;uniqueIndex" json:"instance_id"`
	PID             int        `gorm:"column:pid" json:"pid"`
	State           string     `gorm:"size:50;default:starting;index" json:"state"`
	StartedAt       time.Time  `json:"started_at"`
	ReadyAt         *time.Time `json:"ready_at"`
	StoppedAt       *time.Time `json:"stopped_at"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`
	FaultReason     string     `gorm:"type:text" json:"fault_reason"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (PluginRuntimeSession) TableName() string { return "plugin_runtime_sessions" }

// PluginTrustRecord 记录插件版本信任状态。
type PluginTrustRecord struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	PluginID        string     `gorm:"size:120;uniqueIndex:idx_plugin_trust" json:"plugin_id"`
	Version         string     `gorm:"size:50;uniqueIndex:idx_plugin_trust" json:"version"`
	TrustLevel      string     `gorm:"size:50;default:local-approved" json:"trust_level"`
	SignatureStatus string     `gorm:"size:50;default:unknown" json:"signature_status"`
	ApprovedBy      string     `gorm:"size:120" json:"approved_by"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RiskSummary     string     `gorm:"type:text" json:"risk_summary"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (PluginTrustRecord) TableName() string { return "plugin_trust_records" }

// PluginConfigSchema 记录插件配置 schema。
type PluginConfigSchema struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PluginID      string    `gorm:"size:120;uniqueIndex:idx_plugin_config_schema" json:"plugin_id"`
	ConfigKey     string    `gorm:"size:180;uniqueIndex:idx_plugin_config_schema;default:default" json:"config_key"`
	SchemaVersion string    `gorm:"size:50" json:"schema_version"`
	SchemaJSON    string    `gorm:"type:text" json:"schema_json"`
	Status        string    `gorm:"size:50;default:active" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (PluginConfigSchema) TableName() string { return "plugin_config_schemas" }

// PluginConfigValue 记录插件配置值，普通配置与敏感配置均存放在主数据库。
type PluginConfigValue struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PluginID   string    `gorm:"size:120;uniqueIndex:idx_plugin_config_value" json:"plugin_id"`
	ConfigKey  string    `gorm:"size:180;uniqueIndex:idx_plugin_config_value;default:default" json:"config_key"`
	ValueJSON  string    `gorm:"type:text" json:"value_json"`
	SecretJSON string    `gorm:"type:text" json:"secret_json"`
	Revision   int       `gorm:"default:1" json:"revision"`
	UpdatedBy  string    `gorm:"size:120" json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (PluginConfigValue) TableName() string { return "plugin_config_values" }

// PluginConfigRevision 记录插件配置变更历史。
type PluginConfigRevision struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PluginID      string    `gorm:"size:120;uniqueIndex:idx_plugin_config_revision" json:"plugin_id"`
	ConfigKey     string    `gorm:"size:180;uniqueIndex:idx_plugin_config_revision;default:default" json:"config_key"`
	Revision      int       `gorm:"uniqueIndex:idx_plugin_config_revision" json:"revision"`
	ValueDigest   string    `gorm:"size:128" json:"value_digest"`
	SecretJSON    string    `gorm:"type:text" json:"secret_json"`
	UpdatedBy     string    `gorm:"size:120" json:"updated_by"`
	ChangeSummary string    `gorm:"type:text" json:"change_summary"`
	CreatedAt     time.Time `json:"created_at"`
}

func (PluginConfigRevision) TableName() string { return "plugin_config_revisions" }

// PluginStateEvent 记录插件状态流转。
type PluginStateEvent struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	PluginID          string    `gorm:"size:120;index" json:"plugin_id"`
	FromState         string    `gorm:"size:50" json:"from_state"`
	ToState           string    `gorm:"size:50" json:"to_state"`
	EventType         string    `gorm:"size:80;index" json:"event_type"`
	Reason            string    `gorm:"type:text" json:"reason"`
	OperatorSubjectID string    `gorm:"size:120" json:"operator_subject_id"`
	CreatedAt         time.Time `json:"created_at"`
}

func (PluginStateEvent) TableName() string { return "plugin_state_events" }

// PluginFaultLog 记录插件故障。
type PluginFaultLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PluginID    string    `gorm:"size:120;index" json:"plugin_id"`
	InstanceID  string    `gorm:"size:120;index" json:"instance_id"`
	FaultType   string    `gorm:"size:80" json:"fault_type"`
	FaultReason string    `gorm:"type:text" json:"fault_reason"`
	StackTrace  string    `gorm:"type:text" json:"stack_trace"`
	CreatedAt   time.Time `json:"created_at"`
}

func (PluginFaultLog) TableName() string { return "plugin_fault_logs" }
