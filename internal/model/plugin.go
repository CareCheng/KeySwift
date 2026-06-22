package model

import "time"

// PluginRegistry 宿主管理的插件注册表。
type PluginRegistry struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	PluginID             string     `gorm:"size:120;uniqueIndex" json:"plugin_id"`
	InstallID            string     `gorm:"size:120;index" json:"install_id"`
	CurrentVersion       string     `gorm:"size:50" json:"current_version"`
	InstallRoot          string     `gorm:"size:500" json:"install_root"`
	SourceType           string     `gorm:"size:50" json:"source_type"`
	SourceURI            string     `gorm:"size:500" json:"source_uri"`
	Enabled              bool       `gorm:"default:false" json:"enabled"`
	Autostart            bool       `gorm:"default:false" json:"autostart"`
	DesiredState         string     `gorm:"size:50;default:approved-disabled" json:"desired_state"`
	ActualState          string     `gorm:"size:50;default:stopped" json:"actual_state"`
	LifecycleState       string     `gorm:"size:50;default:discovered" json:"lifecycle_state"`
	TrustLevel           string     `gorm:"size:50;default:local-approved" json:"trust_level"`
	SignatureStatus      string     `gorm:"size:50;default:unknown" json:"signature_status"`
	ApprovedManifestHash string     `gorm:"size:128" json:"approved_manifest_hash"`
	ApprovedPackageHash  string     `gorm:"size:128" json:"approved_package_hash"`
	ApprovedBinaryHash   string     `gorm:"size:128" json:"approved_binary_hash"`
	CurrentManifestHash  string     `gorm:"size:128" json:"current_manifest_hash"`
	LastVerifiedAt       *time.Time `json:"last_verified_at"`
	LastVerifyStatus     string     `gorm:"size:50" json:"last_verify_status"`
	TamperStatus         string     `gorm:"size:50" json:"tamper_status"`
	QuarantineReason     string     `gorm:"type:text" json:"quarantine_reason"`
	ConfigVersion        int        `gorm:"default:1" json:"config_version"`
	SelectedOS           string     `gorm:"size:50" json:"selected_os"`
	SelectedArch         string     `gorm:"size:50" json:"selected_arch"`
	HealthStatus         string     `gorm:"size:50;default:stopped" json:"health_status"`
	LastStartAt          *time.Time `json:"last_start_at"`
	LastReadyAt          *time.Time `json:"last_ready_at"`
	LastStopAt           *time.Time `json:"last_stop_at"`
	LastFaultAt          *time.Time `json:"last_fault_at"`
	ManifestJSON         string     `gorm:"type:text" json:"manifest_json"`
	ExtensionsJSON       string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PluginArtifact 记录插件包文件摘要与平台产物。
type PluginArtifact struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PluginID       string     `gorm:"size:120;index" json:"plugin_id"`
	ArtifactID     string     `gorm:"size:120;index" json:"artifact_id"`
	RelativePath   string     `gorm:"size:500" json:"relative_path"`
	ArtifactType   string     `gorm:"size:50" json:"artifact_type"`
	Platform       string     `gorm:"size:50" json:"platform"`
	Arch           string     `gorm:"size:50" json:"arch"`
	SizeBytes      int64      `json:"size_bytes"`
	HashAlgorithm  string     `gorm:"size:50" json:"hash_algorithm"`
	HashValue      string     `gorm:"size:128" json:"hash_value"`
	IsExecutable   bool       `gorm:"default:false" json:"is_executable"`
	IsRequired     bool       `gorm:"default:false" json:"is_required"`
	GroupName      string     `gorm:"size:100" json:"group_name"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PluginBinding 记录插件菜单、页面、路由、权限等宿主绑定关系。
type PluginBinding struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PluginID        string    `gorm:"size:120;index" json:"plugin_id"`
	BindingType     string    `gorm:"size:50;index" json:"binding_type"`
	BindingKey      string    `gorm:"size:180;index" json:"binding_key"`
	TargetScope     string    `gorm:"size:80" json:"target_scope"`
	MountArea       string    `gorm:"size:80" json:"mount_area"`
	RouteOrViewID   string    `gorm:"size:180" json:"route_or_view_id"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	OrderHint       int       `gorm:"default:0" json:"order_hint"`
	PermissionGuard string    `gorm:"size:200" json:"permission_guard"`
	ExtensionsJSON  string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PluginConfig 存储插件配置值与 schema 摘要。
type PluginConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PluginID        string    `gorm:"size:120;index" json:"plugin_id"`
	ConfigKey       string    `gorm:"size:180;index" json:"config_key"`
	ConfigVersion   int       `gorm:"default:1" json:"config_version"`
	SchemaJSON      string    `gorm:"type:text" json:"schema_json"`
	ValueJSON       string    `gorm:"type:text" json:"value_json"`
	EncryptedFields string    `gorm:"type:text" json:"encrypted_fields"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	UpdatedBy       string    `gorm:"size:120" json:"updated_by"`
	ExtensionsJSON  string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PluginEventLog 记录宿主与插件事件总线事件。
type PluginEventLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	EventID        string    `gorm:"size:120;uniqueIndex" json:"event_id"`
	EventType      string    `gorm:"size:120;index" json:"event_type"`
	EventVersion   string    `gorm:"size:50" json:"event_version"`
	OccurredAt     time.Time `gorm:"index" json:"occurred_at"`
	ProducerType   string    `gorm:"size:50" json:"producer_type"`
	ProducerID     string    `gorm:"size:120" json:"producer_id"`
	RequestID      string    `gorm:"size:120;index" json:"request_id"`
	SubjectID      string    `gorm:"size:120" json:"subject_id"`
	ResourceType   string    `gorm:"size:80;index" json:"resource_type"`
	ResourceID     string    `gorm:"size:120;index" json:"resource_id"`
	PayloadJSON    string    `gorm:"type:text" json:"payload_json"`
	ExtensionsJSON string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time `json:"created_at"`
}

// PluginJob 记录插件异步任务。
type PluginJob struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	JobID          string     `gorm:"size:120;uniqueIndex" json:"job_id"`
	JobType        string     `gorm:"size:120;index" json:"job_type"`
	ResourceType   string     `gorm:"size:80;index" json:"resource_type"`
	ResourceID     string     `gorm:"size:120;index" json:"resource_id"`
	Status         string     `gorm:"size:50;index;default:pending" json:"status"`
	Priority       int        `gorm:"default:0" json:"priority"`
	RunAt          time.Time  `gorm:"index" json:"run_at"`
	RetryCount     int        `gorm:"default:0" json:"retry_count"`
	MaxRetries     int        `gorm:"default:3" json:"max_retries"`
	LastError      string     `gorm:"type:text" json:"last_error"`
	PayloadJSON    string     `gorm:"type:text" json:"payload_json"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PluginMigration 记录插件迁移声明和执行状态。
type PluginMigration struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PluginID       string     `gorm:"size:120;index" json:"plugin_id"`
	MigrationID    string     `gorm:"size:180;index" json:"migration_id"`
	Version        string     `gorm:"size:50" json:"version"`
	Direction      string     `gorm:"size:20" json:"direction"`
	Path           string     `gorm:"size:500" json:"path"`
	Checksum       string     `gorm:"size:128" json:"checksum"`
	Status         string     `gorm:"size:50;default:declared" json:"status"`
	ExecutedAt     *time.Time `json:"executed_at"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (PluginRegistry) TableName() string {
	return "plugin_registry"
}

func (PluginArtifact) TableName() string {
	return "plugin_artifacts"
}

func (PluginBinding) TableName() string {
	return "plugin_bindings"
}

func (PluginConfig) TableName() string {
	return "plugin_configs"
}

func (PluginEventLog) TableName() string {
	return "plugin_event_logs"
}

func (PluginJob) TableName() string {
	return "plugin_jobs"
}

func (PluginMigration) TableName() string {
	return "plugin_migrations"
}
